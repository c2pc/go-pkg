package service

import (
	"context"
	"errors"
	"time"

	cache2 "github.com/c2pc/go-pkg/v2/auth/internal/cache"
	model2 "github.com/c2pc/go-pkg/v2/auth/internal/model"
	repository2 "github.com/c2pc/go-pkg/v2/auth/internal/repository"
	"github.com/c2pc/go-pkg/v2/utils/apperr"
	"github.com/c2pc/go-pkg/v2/utils/constant"
	"github.com/c2pc/go-pkg/v2/utils/mcontext"
	"github.com/c2pc/go-pkg/v2/utils/secret"
	"github.com/c2pc/go-pkg/v2/utils/tokenverify"
	"github.com/golang-jwt/jwt/v4"
	"github.com/redis/go-redis/v9"
	"github.com/rs/xid"
	"gorm.io/gorm"
)

type IAuthService interface {
	Trx(db *gorm.DB) IAuthService
	Login(ctx context.Context, input AuthLogin) (*model2.AuthToken, int, error)
	Refresh(ctx context.Context, input AuthRefresh) (*model2.AuthToken, int, error)
	Logout(ctx context.Context, input AuthLogout) (int, error)
	Account(ctx context.Context) (*model2.User, error)
	UpdateAccountData(ctx context.Context, input AuthUpdateAccountData) error
}

type AuthService struct {
	userRepository  repository2.IUserRepository
	tokenRepository repository2.ITokenRepository
	tokenCache      cache2.ITokenCache
	userCache       cache2.IUserCache
	hasher          secret.Hasher
	accessExpire    time.Duration
	refreshExpire   time.Duration
	accessSecret    string
	db              *gorm.DB
}

func NewAuthService(
	userRepository repository2.IUserRepository,
	tokenRepository repository2.ITokenRepository,
	tokenCache cache2.ITokenCache,
	userCache cache2.IUserCache,
	hasher secret.Hasher,
	accessExpire time.Duration,
	refreshExpire time.Duration,
	accessSecret string,
) AuthService {
	return AuthService{
		userRepository:  userRepository,
		tokenRepository: tokenRepository,
		tokenCache:      tokenCache,
		userCache:       userCache,
		hasher:          hasher,
		accessExpire:    accessExpire,
		refreshExpire:   refreshExpire,
		accessSecret:    accessSecret,
	}
}

func (s AuthService) Trx(db *gorm.DB) IAuthService {
	s.userRepository = s.userRepository.Trx(db)
	s.tokenRepository = s.tokenRepository.Trx(db)
	s.db = db
	return s
}

type AuthLogin struct {
	Login    string
	Password string
	DeviceID int
}

func (s AuthService) Login(ctx context.Context, input AuthLogin) (*model2.AuthToken, int, error) {
	user, err := s.userRepository.Find(ctx, "login = ?", input.Login)
	if err != nil {
		return nil, 0, apperr.ErrUnauthenticated.WithError(err)
	}

	if !s.hasher.HashMatchesString(user.Password, input.Password) {
		return nil, user.ID, apperr.ErrUnauthenticated.WithErrorText("hash matches password error")
	}

	if user.Blocked {
		return nil, user.ID, apperr.ErrUnauthenticated.WithErrorText("user is blocked")
	}

	data, err := s.createSession(ctx, true, user.ID, input.DeviceID)
	return data, user.ID, err
}

type AuthRefresh struct {
	Token    string
	DeviceID int
}

func (s AuthService) Refresh(ctx context.Context, input AuthRefresh) (*model2.AuthToken, int, error) {
	token, err := s.tokenRepository.Find(ctx, "token = ? AND device_id = ?", input.Token, input.DeviceID)
	if err != nil {
		return nil, 0, apperr.ErrUnauthenticated.WithError(err)
	}

	if time.Now().UTC().After(token.ExpiresAt) {
		_ = s.tokenRepository.Delete(ctx, "token = ? AND device_id = ?", input.Token, input.DeviceID)
		return nil, token.UserID, apperr.ErrUnauthenticated.WithErrorText("token is expired")
	}

	data, err := s.createSession(ctx, false, token.UserID, token.DeviceID)
	return data, token.UserID, err
}

type AuthLogout struct {
	Token string
}

func (s AuthService) Logout(ctx context.Context, input AuthLogout) (int, error) {
	claims, err := tokenverify.GetClaimFromToken(input.Token, tokenverify.Secret(s.accessSecret))
	if err != nil {
		return 0, apperr.ErrUnauthenticated.WithErrorText("invalid token")
	}

	return claims.UserID, s.clearSession(ctx, claims.UserID, claims.DeviceID)
}

func (s AuthService) Account(ctx context.Context) (*model2.User, error) {
	userID, ok := mcontext.GetOpUserID(ctx)
	if !ok {
		return nil, apperr.ErrUnauthenticated.WithErrorText("operation user id is empty")
	}

	user, err := s.userCache.GetUserInfo(ctx, userID, func(ctx context.Context) (*model2.User, error) {
		user, err := s.userRepository.GetUserWithPermissions(ctx, "id = ?", userID)
		if err != nil {
			return nil, err
		}

		return user, nil
	})
	if err != nil {
		return nil, apperr.ErrUnauthenticated.WithError(err)
	}

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

func (s AuthService) UpdateAccountData(ctx context.Context, input AuthUpdateAccountData) error {
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

	if len(selects) > 0 {
		if err := s.userCache.DelUsersInfo(userID).ChainExecDel(ctx); err != nil {
			return apperr.ErrInternal.WithError(err)
		}
	}

	return nil
}

func (s AuthService) createSession(ctx context.Context, isLogin bool, userID int, deviceID int) (*model2.AuthToken, error) {
	claims := tokenverify.BuildClaims(userID, deviceID, s.accessExpire)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.accessSecret))
	if err != nil {
		return nil, apperr.ErrUnauthenticated.WithError(err)
	}

	refreshToken := xid.New().String()
	expiresAt := time.Now().UTC().Add(s.refreshExpire)

	doUpdate := []interface{}{"token", "expires_at", "updated_at"}
	doCreate := []interface{}{"logged_at"}
	if isLogin {
		doUpdate = append(doUpdate, "logged_at")
	}

	if _, err = s.tokenRepository.CreateOrUpdate(ctx, &model2.RefreshToken{
		UserID:    userID,
		DeviceID:  deviceID,
		Token:     refreshToken,
		LoggedAt:  time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		ExpiresAt: expiresAt,
	}, []interface{}{"user_id", "device_id"}, doUpdate, doCreate); err != nil {
		return nil, apperr.ErrUnauthenticated.WithError(err)
	}

	tokens, err := s.tokenCache.GetTokensWithoutError(ctx, userID, deviceID)
	if err != nil {
		return nil, apperr.ErrUnauthenticated.WithError(err)
	}

	var deleteTokenKey []string
	for k, v := range tokens {
		_, err = tokenverify.GetClaimFromToken(k, tokenverify.Secret(s.accessSecret))
		if err != nil || v != constant.NormalToken {
			deleteTokenKey = append(deleteTokenKey, k)
		}
	}

	if len(deleteTokenKey) != 0 {
		err = s.tokenCache.DeleteTokenByUidPid(ctx, userID, deviceID, deleteTokenKey)
		if err != nil {
			return nil, apperr.ErrUnauthenticated.WithError(err)
		}
	}

	if err = s.tokenCache.SetTokenFlagEx(ctx, userID, deviceID, tokenString, constant.NormalToken); err != nil {
		return nil, apperr.ErrUnauthenticated.WithError(err)
	}

	if err := s.userCache.DelUsersInfo(userID).ChainExecDel(ctx); err != nil {
		return nil, apperr.ErrUnauthenticated.WithError(err)
	}

	user, err := s.userCache.GetUserInfo(ctx, userID, func(ctx context.Context) (*model2.User, error) {
		user, err := s.userRepository.GetUserWithPermissions(ctx, "id = ?", userID)
		if err != nil {
			return nil, err
		}

		return user, nil
	})
	if err != nil {
		return nil, apperr.ErrUnauthenticated.WithError(err)
	}

	return &model2.AuthToken{
		Auth: model2.Token{
			Token:        tokenString,
			RefreshToken: refreshToken,
			ExpiresAt:    s.accessExpire.Seconds(),
			TokenType:    "Bearer",
			UserID:       userID,
		},
		User: *user,
	}, nil
}

func (s AuthService) clearSession(ctx context.Context, userID, deviceID int) error {
	if err := s.tokenRepository.Delete(ctx, `user_id = ? AND device_id = ?`, userID, deviceID); err != nil {
		if !apperr.Is(err, apperr.ErrDBRecordNotFound) {
			return apperr.ErrUnauthenticated.WithError(err)
		}
	}

	m, err := s.tokenCache.GetTokensWithoutError(ctx, userID, deviceID)
	if err != nil && !errors.Is(err, redis.Nil) {
		return apperr.ErrUnauthenticated.WithError(err)
	}
	for k := range m {
		m[k] = constant.KickedToken
		err = s.tokenCache.SetTokenMapByUidPid(ctx, userID, deviceID, m)
		if err != nil {
			return apperr.ErrUnauthenticated.WithError(err)
		}
	}

	if err := s.userCache.DelUsersInfo(userID).ChainExecDel(ctx); err != nil {
		return apperr.ErrUnauthenticated.WithError(err)
	}

	return nil
}
