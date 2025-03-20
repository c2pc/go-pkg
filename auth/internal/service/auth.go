package service

import (
	"context"
	"time"

	cache2 "github.com/c2pc/go-pkg/v2/auth/internal/cache"
	"github.com/c2pc/go-pkg/v2/auth/internal/i18n"
	model2 "github.com/c2pc/go-pkg/v2/auth/internal/model"
	repository2 "github.com/c2pc/go-pkg/v2/auth/internal/repository"
	"github.com/c2pc/go-pkg/v2/auth/profile"
	"github.com/c2pc/go-pkg/v2/utils/apperr"
	"github.com/c2pc/go-pkg/v2/utils/apperr/code"
	"github.com/c2pc/go-pkg/v2/utils/constant"
	"github.com/c2pc/go-pkg/v2/utils/ldapauth"
	"github.com/c2pc/go-pkg/v2/utils/mcontext"
	"github.com/c2pc/go-pkg/v2/utils/secret"
	"github.com/c2pc/go-pkg/v2/utils/tokenverify"
	"github.com/golang-jwt/jwt/v4"
	"github.com/rs/xid"
	"gorm.io/gorm"
)

var (
	ErrAuthNoAccess = apperr.New("auth_blocked", apperr.WithTextTranslate(i18n.ErrAuthNoAccess), apperr.WithCode(code.PermissionDenied))
)

type IAuthService[Model, CreateInput, UpdateInput, UpdateProfileInput any] interface {
	Trx(db *gorm.DB) IAuthService[Model, CreateInput, UpdateInput, UpdateProfileInput]
	Login(ctx context.Context, input AuthLogin) (*model2.AuthToken, int, error)
	Refresh(ctx context.Context, input AuthRefresh) (*model2.AuthToken, int, error)
	Logout(ctx context.Context, input AuthLogout) (int, error)
	Account(ctx context.Context) (*model2.User, error)
	UpdateAccountData(ctx context.Context, input AuthUpdateAccountData, profileInput *UpdateProfileInput) error
}

type AuthService[Model, CreateInput, UpdateInput, UpdateProfileInput any] struct {
	profileService  profile.IProfileService[Model, CreateInput, UpdateInput, UpdateProfileInput]
	userRepository  repository2.IUserRepository
	tokenRepository repository2.ITokenRepository
	tokenCache      cache2.ITokenCache
	userCache       cache2.IUserCache
	hasher          secret.Hasher
	accessExpire    time.Duration
	refreshExpire   time.Duration
	ldapAuth        ldapauth.AuthService
	accessSecret    string
}

func NewAuthService[Model, CreateInput, UpdateInput, UpdateProfileInput any](
	profileService profile.IProfileService[Model, CreateInput, UpdateInput, UpdateProfileInput],
	userRepository repository2.IUserRepository,
	tokenRepository repository2.ITokenRepository,
	tokenCache cache2.ITokenCache,
	userCache cache2.IUserCache,
	hasher secret.Hasher,
	accessExpire time.Duration,
	refreshExpire time.Duration,
	accessSecret string,
	ldapAuth ldapauth.AuthService,
) AuthService[Model, CreateInput, UpdateInput, UpdateProfileInput] {
	return AuthService[Model, CreateInput, UpdateInput, UpdateProfileInput]{
		profileService:  profileService,
		userRepository:  userRepository,
		tokenRepository: tokenRepository,
		tokenCache:      tokenCache,
		userCache:       userCache,
		hasher:          hasher,
		accessExpire:    accessExpire,
		refreshExpire:   refreshExpire,
		accessSecret:    accessSecret,
		ldapAuth:        ldapAuth,
	}
}

func (s AuthService[Model, CreateInput, UpdateInput, UpdateProfileInput]) Trx(db *gorm.DB) IAuthService[Model, CreateInput, UpdateInput, UpdateProfileInput] {
	s.userRepository = s.userRepository.Trx(db)
	s.tokenRepository = s.tokenRepository.Trx(db)

	if s.profileService != nil {
		s.profileService = s.profileService.Trx(db)
	}

	return s
}

type AuthLogin struct {
	Login    string
	Password string
	DeviceID int
	IsDomain bool
}

func (s AuthService[Model, CreateInput, UpdateInput, UpdateProfileInput]) Login(ctx context.Context, input AuthLogin) (*model2.AuthToken, int, error) {
	user, err := s.userRepository.Find(ctx, "login = ?", input.Login)
	if err != nil {
		return nil, 0, apperr.ErrUnauthenticated.WithError(err)
	}

	if user.Blocked {
		return nil, user.ID, ErrAuthNoAccess.WithErrorText("user is blocked")
	}

	var accessTokenTTL int64
	var refreshToken string
	var isDomain bool

	if input.IsDomain && s.ldapAuth != nil {
		ldapClaims, err := s.ldapAuth.Login(input.Login, input.Password)
		if err != nil {
			if apperr.Is(err, apperr.ErrUnauthenticated) {
				return nil, user.ID, apperr.ErrUnauthenticated.WithError(err)
			}
			if apperr.Is(err, apperr.ErrForbidden) {
				return nil, user.ID, ErrAuthNoAccess.WithError(err)
			}
			return nil, user.ID, err
		}
		refreshToken = ldapClaims.Refresh
		isDomain = true

		accessTokenTTL = ldapClaims.UserData.ExpiresAt - ldapClaims.UserData.IssuedAt
	} else {
		if !s.hasher.HashMatchesString(user.Password, input.Password) {
			return nil, user.ID, apperr.ErrUnauthenticated.WithErrorText("hash matches password error")
		}
		refreshToken = xid.New().String()
		isDomain = false
		accessTokenTTL = 0
	}

	data, err := s.createSession(ctx, createSessionInput{
		IsLogin:        true,
		UserID:         user.ID,
		DeviceID:       input.DeviceID,
		RefreshToken:   refreshToken,
		AccessTokenTTL: accessTokenTTL,
		IsDomain:       isDomain,
	})

	return data, user.ID, err
}

type AuthRefresh struct {
	Token    string
	DeviceID int
}

func (s AuthService[Model, CreateInput, UpdateInput, UpdateProfileInput]) Refresh(ctx context.Context, input AuthRefresh) (*model2.AuthToken, int, error) {
	token, err := s.tokenRepository.With("user").Find(ctx, "token = ? AND device_id = ?", input.Token, input.DeviceID)
	if err != nil {
		return nil, 0, apperr.ErrUnauthenticated.WithError(err)
	}

	var accessTokenTTL int64
	var refreshToken string
	var isDomain bool

	if token.User.Blocked {
		if err := s.userCache.DelUsersInfo(token.User.ID).ChainExecDel(ctx); err != nil {
			return nil, token.User.ID, apperr.ErrUnauthenticated.WithError(err)
		}

		if err := s.tokenCache.DeleteAllUserTokens(ctx, token.User.ID); err != nil {
			return nil, token.User.ID, apperr.ErrUnauthenticated.WithError(err)
		}

		return nil, token.User.ID, ErrAuthNoAccess.WithErrorText("user is blocked")
	}

	if token.Domain {
		if s.ldapAuth == nil {
			return nil, token.UserID, apperr.ErrUnauthenticated.WithErrorText("domain auth disabled")
		}

		ldapClaims, err := s.ldapAuth.Refresh(input.Token)
		if err != nil {
			if apperr.Is(err, apperr.ErrUnauthenticated) {
				return nil, token.UserID, apperr.ErrUnauthenticated.WithError(err)
			}
			if apperr.Is(err, apperr.ErrForbidden) {
				return nil, token.UserID, ErrAuthNoAccess.WithError(err)
			}

			return nil, token.UserID, err
		}

		refreshToken = ldapClaims.Refresh
		isDomain = true
		accessTokenTTL = ldapClaims.UserData.ExpiresAt - ldapClaims.UserData.IssuedAt
	} else {
		if time.Now().UTC().After(token.ExpiresAt) {
			_ = s.tokenRepository.Delete(ctx, "token = ? ", input.Token)
			return nil, token.UserID, apperr.ErrUnauthenticated.WithErrorText("token is expired")
		}

		refreshToken = xid.New().String()
		isDomain = false
		accessTokenTTL = 0
	}

	data, err := s.createSession(ctx, createSessionInput{
		IsLogin:        false,
		UserID:         token.UserID,
		DeviceID:       token.DeviceID,
		RefreshToken:   refreshToken,
		AccessTokenTTL: accessTokenTTL,
		IsDomain:       isDomain,
	})
	return data, token.UserID, err
}

type AuthLogout struct {
	Token string
}

func (s AuthService[Model, CreateInput, UpdateInput, UpdateProfileInput]) Logout(ctx context.Context, input AuthLogout) (int, error) {
	claims, err := tokenverify.GetClaimFromToken(input.Token, tokenverify.Secret(s.accessSecret))
	if err != nil {
		return 0, apperr.ErrUnauthenticated.WithErrorText("invalid token")
	}

	return claims.UserID, s.clearSession(ctx, claims.UserID, claims.DeviceID)
}

func (s AuthService[Model, CreateInput, UpdateInput, UpdateProfileInput]) Account(ctx context.Context) (*model2.User, error) {
	userID, ok := mcontext.GetOpUserID(ctx)
	if !ok {
		return nil, apperr.ErrUnauthenticated.WithErrorText("operation user id is empty")
	}

	user, err := s.userRepository.GetUserWithPermissions(ctx, "id = ?", userID)
	if err != nil {
		return nil, apperr.ErrUnauthenticated.WithError(err)
	}

	var prof *Model
	if s.profileService != nil {
		prof, err = s.profileService.GetById(ctx, userID)
		if err != nil {
			if !apperr.Is(err, profile.ErrNotFound) {
				return nil, apperr.ErrUnauthenticated.WithError(err)
			}
		}
	}

	user.Profile = prof

	return user, nil
}

type AuthUpdateAccountData struct {
	Login      *string
	FirstName  *string
	SecondName *string
	LastName   *string
	Password   *string
	Email      *string
	Phone      *string
}

func (s AuthService[Model, CreateInput, UpdateInput, UpdateProfileInput]) UpdateAccountData(ctx context.Context, input AuthUpdateAccountData, profileInput *UpdateProfileInput) error {
	userID, ok := mcontext.GetOpUserID(ctx)
	if !ok {
		return apperr.ErrUnauthenticated.WithErrorText("operation user id is empty")
	}

	user := &model2.User{}

	var selects []interface{}
	if input.Login != nil && *input.Login != "" {
		user.Login = *input.Login
		selects = append(selects, "login")
	}

	if input.FirstName != nil && *input.FirstName != "" {
		user.FirstName = *input.FirstName
		selects = append(selects, "first_name")
	}

	if input.Password != nil && *input.Password != "" {
		password := s.hasher.HashString(*input.Password)
		user.Password = password
		selects = append(selects, "password")
	}

	if input.SecondName != nil {
		if *input.SecondName == "" {
			user.SecondName = nil
		} else {
			user.SecondName = input.SecondName
		}
		selects = append(selects, "second_name")
	}

	if input.LastName != nil {
		if *input.LastName == "" {
			user.LastName = nil
		} else {
			user.LastName = input.LastName
		}
		selects = append(selects, "last_name")
	}

	if input.Email != nil {
		if *input.Email == "" {
			user.Email = nil
		} else {
			user.Email = input.Email
		}
		selects = append(selects, "email")
	}

	if input.Phone != nil {
		if *input.Phone == "" {
			user.Phone = nil
		} else {
			user.Phone = input.Phone
		}
		selects = append(selects, "phone")
	}

	if len(selects) > 0 {
		if err := s.userRepository.Update(ctx, user, selects, `id = ?`, userID); err != nil {
			if apperr.Is(err, apperr.ErrDBDuplicated) {
				return ErrUserExists
			}
			return err
		}
	}

	if s.profileService != nil && profileInput != nil {
		err := s.profileService.UpdateProfile(ctx, userID, *profileInput)
		if err != nil {
			return err
		}
	}

	if len(selects) > 0 {
		if err := s.userCache.DelUsersInfo(userID).ChainExecDel(ctx); err != nil {
			return apperr.ErrInternal.WithError(err)
		}
	}

	return nil
}

type createSessionInput struct {
	IsLogin        bool   //Авторизация или refresh
	UserID         int    //ID пользователя
	DeviceID       int    //ID устройства
	RefreshToken   string //Refresh токен
	AccessTokenTTL int64  //Количество секунд до истечения access токена
	IsDomain       bool   //Доменная авторизация
}

func (s AuthService[Model, CreateInput, UpdateInput, UpdateProfileInput]) createSession(ctx context.Context, input createSessionInput) (*model2.AuthToken, error) {
	accessTokenTTL := s.accessExpire
	if input.AccessTokenTTL > 0 {
		accessTokenTTL = time.Duration(input.AccessTokenTTL) * time.Second
	}

	claims := tokenverify.BuildClaims(input.UserID, input.DeviceID, accessTokenTTL)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.accessSecret))
	if err != nil {
		return nil, apperr.ErrUnauthenticated.WithError(err)
	}

	var expiresAt time.Time
	if input.IsDomain {
		expiresAt = time.Now().UTC()
	} else {
		expiresAt = time.Now().UTC().Add(s.refreshExpire)
	}

	doUpdate := []interface{}{"token", "expires_at", "updated_at", "domain"}
	doCreate := []interface{}{"logged_at", "domain"}
	if input.IsLogin {
		doUpdate = append(doUpdate, []string{"logged_at"})
	}

	if _, err := s.tokenRepository.CreateOrUpdate(ctx, &model2.RefreshToken{
		UserID:    input.UserID,
		DeviceID:  input.DeviceID,
		Token:     input.RefreshToken,
		LoggedAt:  time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		ExpiresAt: expiresAt,
		Domain:    input.IsDomain,
	}, []interface{}{"user_id", "device_id"}, doUpdate, doCreate); err != nil {
		return nil, apperr.ErrUnauthenticated.WithError(err)
	}

	tokens, err := s.tokenCache.GetTokensWithoutError(ctx, input.UserID, input.DeviceID)
	if err != nil {
		return nil, apperr.ErrUnauthenticated.WithError(err)
	}

	var deleteTokenKey []string
	for k, _ := range tokens {
		//_, err = tokenverify.GetClaimFromToken(k, tokenverify.Secret(s.accessSecret))
		//if err != nil || v != constant.NormalToken {
		//	deleteTokenKey = append(deleteTokenKey, k)
		//}
		deleteTokenKey = append(deleteTokenKey, k)
	}

	if len(deleteTokenKey) != 0 {
		err = s.tokenCache.DeleteTokenByUidPid(ctx, input.UserID, input.DeviceID, deleteTokenKey)
		if err != nil {
			return nil, apperr.ErrUnauthenticated.WithError(err)
		}
	}

	if err = s.tokenCache.SetTokenFlagEx(ctx, input.UserID, input.DeviceID, tokenString, constant.NormalToken); err != nil {
		return nil, apperr.ErrUnauthenticated.WithError(err)
	}

	if err := s.userCache.DelUsersInfo(input.UserID).ChainExecDel(ctx); err != nil {
		return nil, apperr.ErrUnauthenticated.WithError(err)
	}

	user, err := s.userCache.GetUserInfo(ctx, input.UserID, func(ctx context.Context) (*model2.User, error) {
		user, err := s.userRepository.GetUserWithPermissions(ctx, "id = ?", input.UserID)
		if err != nil {
			return nil, err
		}

		return user, nil
	})
	if err != nil {
		return nil, apperr.ErrUnauthenticated.WithError(err)
	}

	var prof *Model
	if s.profileService != nil {
		prof, err = s.profileService.GetById(ctx, input.UserID)
		if err != nil {
			if !apperr.Is(err, profile.ErrNotFound) {
				return nil, apperr.ErrUnauthenticated.WithError(err)
			}
		}
	}

	user.Profile = prof

	return &model2.AuthToken{
		Auth: model2.Token{
			Token:        tokenString,
			RefreshToken: input.RefreshToken,
			ExpiresAt:    accessTokenTTL.Seconds(),
			TokenType:    "Bearer",
			UserID:       input.UserID,
		},
		User: *user,
	}, nil
}

func (s AuthService[Model, CreateInput, UpdateInput, UpdateProfileInput]) clearSession(ctx context.Context, userID, deviceID int) error {
	if err := s.tokenRepository.Delete(ctx, `user_id = ? AND device_id = ?`, userID, deviceID); err != nil {
		if !apperr.Is(err, apperr.ErrDBRecordNotFound) {
			return apperr.ErrUnauthenticated.WithError(err)
		}
	}

	tokens, err := s.tokenCache.GetTokensWithoutError(ctx, userID, deviceID)
	if err != nil {
		return apperr.ErrUnauthenticated.WithError(err)
	}

	var deleteTokenKey []string
	for k, _ := range tokens {
		deleteTokenKey = append(deleteTokenKey, k)
	}

	if len(deleteTokenKey) != 0 {
		err = s.tokenCache.DeleteTokenByUidPid(ctx, userID, deviceID, deleteTokenKey)
		if err != nil {
			return apperr.ErrUnauthenticated.WithError(err)
		}
	}

	if err := s.userCache.DelUsersInfo(userID).ChainExecDel(ctx); err != nil {
		return apperr.ErrUnauthenticated.WithError(err)
	}

	return nil
}
