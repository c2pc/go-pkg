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
	"github.com/c2pc/go-pkg/v2/utils/mcontext"
	"github.com/c2pc/go-pkg/v2/utils/secret"
	"github.com/c2pc/go-pkg/v2/utils/sso"
	"github.com/c2pc/go-pkg/v2/utils/sso/ldap"
	"github.com/c2pc/go-pkg/v2/utils/sso/oidc"
	"github.com/c2pc/go-pkg/v2/utils/sso/saml"
	"github.com/c2pc/go-pkg/v2/utils/tokenverify"
	"github.com/golang-jwt/jwt/v4"
	"github.com/rs/xid"
	"gorm.io/gorm"
)

var (
	ErrAuthNoAccess    = apperr.New("auth_blocked", apperr.WithTextTranslate(i18n.ErrAuthNoAccess), apperr.WithCode(code.PermissionDenied))
	ErrSSONotSupported = apperr.New("sso_not_supported", apperr.WithTextTranslate(i18n.ErrSSONotSupported), apperr.WithCode(code.Aborted))
)

type IAuthService[Model, CreateInput, UpdateInput, UpdateProfileInput any] interface {
	Trx(db *gorm.DB) IAuthService[Model, CreateInput, UpdateInput, UpdateProfileInput]
	Login(ctx context.Context, input AuthLogin) (*model2.AuthToken, int, error)
	Refresh(ctx context.Context, input AuthRefresh) (*model2.AuthToken, int, error)
	Logout(ctx context.Context, input AuthLogout) (int, error)
	Account(ctx context.Context) (*model2.User, error)
	UpdateAccountData(ctx context.Context, input AuthUpdateAccountData, profileInput *UpdateProfileInput) error
	SSO(ctx context.Context, input SSO) (*model2.AuthToken, int, error)
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
	ldapAuth        ldap.AuthService
	oidcAuth        oidc.AuthService
	samlAuth        saml.AuthService
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
	ldapAuth ldap.AuthService,
	oidcAuth oidc.AuthService,
	samlAuth saml.AuthService,
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
		oidcAuth:        oidcAuth,
		samlAuth:        samlAuth,
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

	var provider, refreshToken string
	var refreshExpiredAt, accessExpiredAt time.Duration
	if input.IsDomain && s.ldapAuth != nil && s.ldapAuth.IsEnabled() {
		err = s.ldapAuth.CheckAuth(input.Login, input.Password)
		if err != nil {
			return nil, user.ID, apperr.ErrUnauthenticated.WithError(err)
		}

		provider = "ldap"
		refreshExpiredAt = s.refreshExpire
		accessExpiredAt = s.accessExpire
		refreshToken = xid.New().String()
	} else {
		if !s.hasher.HashMatchesString(user.Password, input.Password) {
			return nil, user.ID, apperr.ErrUnauthenticated.WithErrorText("hash matches password error")
		}
		refreshExpiredAt = s.refreshExpire
		accessExpiredAt = s.accessExpire
		refreshToken = xid.New().String()
	}

	data, err := s.createSession(ctx, createSessionInput{
		IsLogin:          true,
		UserID:           user.ID,
		DeviceID:         input.DeviceID,
		Provider:         provider,
		RefreshExpiredAt: refreshExpiredAt,
		RefreshToken:     refreshToken,
		AccessExpiredAt:  accessExpiredAt,
	})

	return data, user.ID, err
}

type SSO struct {
	Provider     string
	RefreshToken string
	AccessExpire time.Duration
	Login        string
	DeviceID     int
}

func (s AuthService[Model, CreateInput, UpdateInput, UpdateProfileInput]) SSO(ctx context.Context, input SSO) (*model2.AuthToken, int, error) {
	user, err := s.userRepository.Find(ctx, "login = ?", input.Login)
	if err != nil {
		return nil, 0, ErrAuthNoAccess.WithError(err)
	}

	if user.Blocked {
		return nil, user.ID, ErrAuthNoAccess.WithErrorText("user is blocked")
	}

	if input.Provider == sso.OIDC && !s.oidcAuth.IsEnabled() {
		return nil, user.ID, ErrSSONotSupported
	}

	if input.Provider == sso.SAML {
		if !s.samlAuth.IsEnabled() {
			return nil, user.ID, ErrSSONotSupported
		}
		input.RefreshToken = xid.New().String()
		input.AccessExpire = s.accessExpire
	}

	data, err := s.createSession(ctx, createSessionInput{
		IsLogin:          true,
		UserID:           user.ID,
		DeviceID:         input.DeviceID,
		Provider:         input.Provider,
		RefreshExpiredAt: s.refreshExpire,
		RefreshToken:     input.RefreshToken,
		AccessExpiredAt:  input.AccessExpire,
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

	if token.User.Blocked {
		err := s.clearSession(ctx, token.UserID, token.DeviceID, true)
		if err != nil {
			return nil, token.User.ID, ErrAuthNoAccess.WithError(err)
		}

		return nil, token.User.ID, ErrAuthNoAccess.WithErrorText("user is blocked")
	}

	var provider, refreshToken string
	var refreshExpiredAt, accessExpiredAt time.Duration
	err = func() error {
		if time.Now().UTC().After(token.ExpiresAt) {
			return apperr.ErrUnauthenticated.WithErrorText("token is expired")
		}

		if token.Provider == nil || (token.Provider != nil && *token.Provider == "ldap") {
			if token.Provider != nil {
				provider = *token.Provider
			}
			refreshExpiredAt = s.refreshExpire
			accessExpiredAt = s.accessExpire
			refreshToken = xid.New().String()
		} else if token.Provider != nil {
			if *token.Provider == sso.OIDC && s.oidcAuth.IsEnabled() {
				oidcToken, err := s.oidcAuth.Refresh(ctx, input.Token)
				if err != nil {
					_ = s.tokenRepository.Delete(ctx, "token = ? ", input.Token)
					return ErrAuthNoAccess.WithErrorText("sso not enabled")
				}
				provider = sso.OIDC
				refreshExpiredAt = s.refreshExpire
				accessExpiredAt = time.Duration(oidcToken.IDToken.Expiry.UTC().Sub(time.Now().UTC()).Minutes()) * time.Minute
				refreshToken = oidcToken.IDToken.RefreshToken
			} else if *token.Provider == sso.SAML && s.samlAuth.IsEnabled() {
				provider = sso.SAML
				refreshExpiredAt = s.refreshExpire
				accessExpiredAt = s.accessExpire
				refreshToken = xid.New().String()
			} else {
				return apperr.ErrUnauthenticated.WithError(ErrSSONotSupported)
			}
		} else {
			return apperr.ErrUnauthenticated.WithErrorText("invalid name service")
		}

		return nil
	}()
	if err != nil {
		_ = s.tokenRepository.Delete(ctx, "token = ? ", input.Token)
		return nil, token.User.ID, err
	}

	data, err := s.createSession(ctx, createSessionInput{
		IsLogin:          false,
		UserID:           token.UserID,
		DeviceID:         token.DeviceID,
		Provider:         provider,
		RefreshExpiredAt: refreshExpiredAt,
		RefreshToken:     refreshToken,
		AccessExpiredAt:  accessExpiredAt,
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

	return claims.UserID, s.clearSession(ctx, claims.UserID, claims.DeviceID, true)
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
		password, err := s.hasher.HashString(*input.Password)
		if err != nil {
			return apperr.ErrUnauthenticated.WithError(err)
		}
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
	IsLogin          bool
	UserID           int
	DeviceID         int
	Provider         string
	RefreshExpiredAt time.Duration
	RefreshToken     string
	AccessExpiredAt  time.Duration
}

func (s AuthService[Model, CreateInput, UpdateInput, UpdateProfileInput]) createSession(ctx context.Context, input createSessionInput) (*model2.AuthToken, error) {
	claims := tokenverify.BuildClaims(input.UserID, input.DeviceID, input.AccessExpiredAt)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.accessSecret))
	if err != nil {
		return nil, apperr.ErrUnauthenticated.WithError(err)
	}

	doUpdate := []interface{}{"token", "expires_at", "updated_at"}
	doCreate := []interface{}{"logged_at", "provider"}
	if input.IsLogin {
		doUpdate = append(doUpdate, []string{"logged_at"})
	}

	var provider *string
	if input.Provider != "" {
		provider = &input.Provider
	}

	if _, err := s.tokenRepository.CreateOrUpdate(ctx, &model2.RefreshToken{
		UserID:    input.UserID,
		DeviceID:  input.DeviceID,
		Token:     input.RefreshToken,
		LoggedAt:  time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		ExpiresAt: time.Now().UTC().Add(input.RefreshExpiredAt),
		Provider:  provider,
	}, []interface{}{"user_id", "device_id"}, doUpdate, doCreate); err != nil {
		return nil, apperr.ErrUnauthenticated.WithError(err)
	}

	err = s.clearSession(ctx, input.UserID, input.DeviceID, false)
	if err != nil {
		return nil, apperr.ErrUnauthenticated.WithError(err)
	}

	if err = s.tokenCache.SetTokenFlagEx(ctx, input.UserID, input.DeviceID, tokenString, constant.NormalToken); err != nil {
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
			ExpiresAt:    input.RefreshExpiredAt.Seconds(),
			TokenType:    "Bearer",
			UserID:       input.UserID,
		},
		User: *user,
	}, nil
}

func (s AuthService[Model, CreateInput, UpdateInput, UpdateProfileInput]) clearSession(ctx context.Context, userID, deviceID int, clearRefresh bool) error {
	if clearRefresh {
		if err := s.tokenRepository.Delete(ctx, `user_id = ? AND device_id = ?`, userID, deviceID); err != nil {
			if !apperr.Is(err, apperr.ErrDBRecordNotFound) {
				return apperr.ErrUnauthenticated.WithError(err)
			}
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
