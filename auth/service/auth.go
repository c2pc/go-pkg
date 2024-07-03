package service

import (
	"context"
	"errors"
	"github.com/c2pc/go-pkg/v2/auth/cache"
	"github.com/c2pc/go-pkg/v2/auth/model"
	"github.com/c2pc/go-pkg/v2/auth/repository"
	"github.com/c2pc/go-pkg/v2/utils/apperr"
	"github.com/c2pc/go-pkg/v2/utils/constant"
	"github.com/c2pc/go-pkg/v2/utils/mcontext"
	"github.com/c2pc/go-pkg/v2/utils/secret"
	"github.com/c2pc/go-pkg/v2/utils/tokenverify"
	"github.com/golang-jwt/jwt/v4"
	"github.com/redis/go-redis/v9"
	"github.com/rs/xid"
	"gorm.io/gorm"
	"time"
)

type IAuthService interface {
	Trx(db *gorm.DB) IAuthService
	Login(ctx context.Context, input AuthLogin) (*model.AuthToken, error)
	Refresh(ctx context.Context, input AuthRefresh) (*model.AuthToken, error)
	Logout(ctx context.Context, input AuthLogout) error
	Account(ctx context.Context) (*model.User, error)
}

type AuthService struct {
	userRepository  repository.IUserRepository
	tokenRepository repository.ITokenRepository
	tokenCache      cache.ITokenCache
	userCache       cache.IUserCache
	hasher          secret.Hasher
	accessExpire    time.Duration
	refreshExpire   time.Duration
	accessSecret    string
}

func NewAuthService(
	userRepository repository.IUserRepository,
	tokenRepository repository.ITokenRepository,
	tokenCache cache.ITokenCache,
	userCache cache.IUserCache,
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
	return s
}

type AuthLogin struct {
	Login    string
	Password string
	DeviceID int
}

func (s AuthService) Login(ctx context.Context, input AuthLogin) (*model.AuthToken, error) {
	user, err := s.userRepository.With("roles", "roles.role_permissions").Find(ctx, "users.login = ?", input.Login)
	if err != nil {
		return nil, apperr.ErrUnauthenticated.WithError(err)
	}

	if !s.hasher.HashMatchesPassword(user.Password, input.Password) {
		return nil, apperr.ErrUnauthenticated.WithErrorText("hash matches password error")
	}

	return s.createSession(ctx, user, input.DeviceID)
}

type AuthRefresh struct {
	Token    string
	DeviceID int
}

func (s AuthService) Refresh(ctx context.Context, input AuthRefresh) (*model.AuthToken, error) {
	token, err := s.tokenRepository.With("user", "user.roles", "user.roles.role_permissions").Find(ctx, "tokens.token = ? && tokens.device_id = ?", input.Token, input.DeviceID)
	if err != nil {
		return nil, apperr.ErrUnauthenticated.WithError(err)
	}

	if time.Now().After(token.ExpiresAt) {
		return nil, apperr.ErrUnauthenticated.WithErrorText("token is expired")
	}

	return s.createSession(ctx, token.User, token.DeviceID)
}

type AuthLogout struct {
	Token string
}

func (s AuthService) Logout(ctx context.Context, input AuthLogout) error {
	claims, err := tokenverify.GetClaimFromToken(input.Token, tokenverify.Secret(s.accessSecret))
	if err != nil {
		return apperr.ErrUnauthenticated.WithErrorText("invalid token")
	}

	return s.clearSession(ctx, claims.UserID, claims.DeviceID)
}

func (s AuthService) Account(ctx context.Context) (*model.User, error) {
	userID, ok := mcontext.GetOpUserID(ctx)
	if !ok {
		return nil, apperr.ErrUnauthenticated.WithErrorText("operation user id is empty")
	}

	user, err := s.userCache.GetUserInfo(ctx, userID, func(ctx context.Context) (*model.User, error) {
		return s.userRepository.With("roles", "roles.role_permissions").Find(ctx, "users.id = ?", userID)
	})
	if err != nil {
		return nil, apperr.ErrUnauthenticated.WithError(err)
	}

	return user, nil
}

func (s AuthService) createSession(ctx context.Context, user *model.User, deviceID int) (*model.AuthToken, error) {
	claims := tokenverify.BuildClaims(user.ID, deviceID, s.accessExpire)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.accessSecret))
	if err != nil {
		return nil, apperr.ErrUnauthenticated.WithError(err)
	}

	refreshToken := xid.New().String()
	expiresAt := time.Now().Add(s.refreshExpire)

	if _, err = s.tokenRepository.CreateOrUpdate(ctx, &model.RefreshToken{
		UserID:    user.ID,
		DeviceID:  deviceID,
		Token:     refreshToken,
		ExpiresAt: expiresAt,
	}, []interface{}{"user_id", "device_id"}, []interface{}{"refresh_token", "expires_at"}); err != nil {
		return nil, apperr.ErrUnauthenticated.WithError(err)
	}

	tokens, err := s.tokenCache.GetTokensWithoutError(ctx, user.ID, deviceID)
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
		err = s.tokenCache.DeleteTokenByUidPid(ctx, user.ID, deviceID, deleteTokenKey)
		if err != nil {
			return nil, apperr.ErrUnauthenticated.WithError(err)
		}
	}

	if err = s.tokenCache.SetTokenFlagEx(ctx, user.ID, deviceID, tokenString, constant.NormalToken); err != nil {
		return nil, apperr.ErrUnauthenticated.WithError(err)
	}

	if err := s.userCache.DelUsersInfo(user.ID).ChainExecDel(ctx); err != nil {
		return nil, apperr.ErrUnauthenticated.WithError(err)
	}

	return &model.AuthToken{
		Auth: model.Token{
			Token:        tokenString,
			RefreshToken: refreshToken,
			ExpiresAt:    s.accessExpire.Seconds(),
			TokenType:    "Bearer",
			UserID:       user.ID,
		},
		User: *user,
	}, nil
}

func (s AuthService) clearSession(ctx context.Context, userID, deviceID int) error {
	if err := s.tokenRepository.Delete(ctx, `tokens.user_id = ? AND tokens.device_id = ?`, userID, deviceID); err != nil {
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
