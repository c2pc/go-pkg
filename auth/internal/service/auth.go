package service

import (
	"context"
	"crypto/sha256"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/c2pc/go-pkg/v2/auth/fx"
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
	"github.com/c2pc/go-pkg/v2/utils/tokenverify"
	"github.com/golang-jwt/jwt/v4"
	"github.com/rs/xid"
	"gorm.io/gorm"
)

var (
	ErrAuthNoAccess    = apperr.New("auth_no_access", apperr.WithTextTranslate(i18n.ErrAuthNoAccess), apperr.WithCode(code.PermissionDenied))
	ErrAuthBlocked     = apperr.New("auth_blocked", apperr.WithTextTranslate(i18n.ErrAuthNoAccess), apperr.WithCode(code.PermissionDenied))
	ErrSSONotSupported = apperr.New("sso_not_supported", apperr.WithTextTranslate(i18n.ErrSSONotSupported), apperr.WithCode(code.Unauthenticated))
)

type IAuthService interface {
	Trx(db *gorm.DB) IAuthService
	Login(ctx context.Context, input AuthLogin) (*model2.AuthToken, int, error)
	Refresh(ctx context.Context, input AuthRefresh) (*model2.AuthToken, int, error)
	Logout(ctx context.Context, input AuthLogout) (int, error)
	Account(ctx context.Context) (*model2.User, error)
	SSO(ctx context.Context, input SSO) (*model2.AuthToken, int, error)
}

type AuthService struct {
	profileService profile.IProfileService
	repositories   repository2.Repositories
	cache          *fx.CacheHolder
	cfg            *fx.AuthHolder
	ldapAuth       *fx.LDAPHolder
	oidcAuth       *fx.OIDCHolder
	samlAuth       *fx.SAMLHolder
}

func NewAuthService(
	profileService profile.IProfileService,
	repositories repository2.Repositories,
	cache *fx.CacheHolder,
	cfg *fx.AuthHolder,
	ldapAuth *fx.LDAPHolder,
	oidcAuth *fx.OIDCHolder,
	samlAuth *fx.SAMLHolder,
) AuthService {
	return AuthService{
		profileService: profileService,
		repositories:   repositories,
		cache:          cache,
		cfg:            cfg,
		ldapAuth:       ldapAuth,
		oidcAuth:       oidcAuth,
		samlAuth:       samlAuth,
	}
}

func (s AuthService) Trx(db *gorm.DB) IAuthService {
	s.repositories.UserRepository = s.repositories.UserRepository.Trx(db)
	s.repositories.TokenRepository = s.repositories.TokenRepository.Trx(db)

	if s.profileService != nil {
		s.profileService = s.profileService.Trx(db)
	}

	return s
}

type AuthLogin struct {
	Login    string
	Password string
	Secret   string
	DeviceID int
}

func (s AuthService) Login(ctx context.Context, input AuthLogin) (*model2.AuthToken, int, error) {
	user, err := s.repositories.UserRepository.With("roles").Find(ctx, "login = ?", input.Login)
	if err != nil {
		return nil, 0, apperr.ErrUnauthenticated.WithError(err)
	}

	var isBroker bool
	for _, role := range user.Roles {
		if role.Name == model2.Broker {
			isBroker = true
			break
		}
	}

	if isBroker {
		secr := sha256.Sum256([]byte(input.Login + input.Password + strconv.Itoa(input.DeviceID)))
		if input.Secret == "" || input.Secret != fmt.Sprintf("%x", secr) {
			return nil, 0, ErrAuthBlocked.WithErrorText("broker invalid")
		}
	}

	if user.Blocked {
		return nil, user.ID, ErrAuthBlocked.WithErrorText("user is blocked")
	}

	var domain string
	spl := strings.SplitN(user.Login, "@", 2)
	if len(spl) == 2 {
		domain = spl[1]
	}

	var provider, refreshToken string
	if user.IsDomain {
		if s.ldapAuth.Get() != nil && s.ldapAuth.Get().IsEnabled() {
			err = s.ldapAuth.Get().CheckAuth(domain, input.Login, input.Password)
			if err != nil {
				return nil, user.ID, apperr.ErrUnauthenticated.WithError(err)
			}

			provider = "ldap"
			refreshToken = xid.New().String()
		} else {
			return nil, user.ID, apperr.ErrUnauthenticated.WithErrorText("domain is empty")
		}
	} else {
		if user.Password != nil {
			if !secret.HasherSecret.HashMatchesString(*user.Password, input.Password) {
				return nil, user.ID, apperr.ErrUnauthenticated.WithErrorText("hash matches password error")
			}
			refreshToken = xid.New().String()
		} else {
			return nil, user.ID, apperr.ErrUnauthenticated
		}
	}

	data, err := s.createSession(ctx, createSessionInput{
		IsLogin:      true,
		UserID:       user.ID,
		DeviceID:     input.DeviceID,
		Provider:     provider,
		RefreshToken: refreshToken,
	})

	return data, user.ID, err
}

type SSO struct {
	Provider     string
	RefreshToken string
	Login        string
	DeviceID     int
}

func (s AuthService) SSO(ctx context.Context, input SSO) (*model2.AuthToken, int, error) {
	user, err := s.repositories.UserRepository.Find(ctx, "login = ?", input.Login)
	if err != nil {
		return nil, 0, ErrAuthNoAccess.WithError(err)
	}

	if user.Blocked {
		return nil, user.ID, ErrAuthBlocked.WithErrorText("user is blocked")
	}

	if input.Provider == sso.OIDC && !s.oidcAuth.Get().IsEnabled() {
		return nil, user.ID, ErrSSONotSupported
	}

	if input.Provider == sso.SAML {
		if !s.samlAuth.Get().IsEnabled() {
			return nil, user.ID, ErrSSONotSupported
		}
		input.RefreshToken = xid.New().String()
	}

	data, err := s.createSession(ctx, createSessionInput{
		IsLogin:      true,
		UserID:       user.ID,
		DeviceID:     input.DeviceID,
		Provider:     input.Provider,
		RefreshToken: input.RefreshToken,
	})

	return data, user.ID, err
}

type AuthRefresh struct {
	Token    string
	DeviceID int
}

func (s AuthService) Refresh(ctx context.Context, input AuthRefresh) (*model2.AuthToken, int, error) {
	token, err := s.repositories.TokenRepository.With("user").Find(ctx, "token = ? AND device_id = ?", input.Token, input.DeviceID)
	if err != nil {
		return nil, 0, apperr.ErrUnauthenticated.WithError(err)
	}

	if token.User.Blocked {
		err := s.clearSession(ctx, token.UserID, token.DeviceID, true)
		if err != nil {
			return nil, token.User.ID, ErrAuthBlocked.WithError(err)
		}

		return nil, token.User.ID, ErrAuthBlocked.WithErrorText("user is blocked")
	}

	var provider, refreshToken string
	err = func() error {
		if time.Now().UTC().After(token.ExpiresAt) {
			return apperr.ErrUnauthenticated.WithErrorText("token is expired")
		}

		if token.Provider == nil || (token.Provider != nil && *token.Provider == "ldap") {
			if token.Provider != nil {
				provider = *token.Provider
			}
			refreshToken = xid.New().String()
		} else if token.Provider != nil {
			if *token.Provider == sso.OIDC && s.oidcAuth.Get().IsEnabled() {
				oidcToken, err := s.oidcAuth.Get().Refresh(ctx, input.Token)
				if err != nil {
					_ = s.repositories.TokenRepository.Delete(ctx, "token = ? ", input.Token)
					return apperr.ErrUnauthenticated.WithErrorText("error to refresh token")
				}
				provider = sso.OIDC
				refreshToken = oidcToken.IDToken.RefreshToken
			} else if *token.Provider == sso.SAML && s.samlAuth.Get().IsEnabled() {
				provider = sso.SAML
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
		_ = s.repositories.TokenRepository.Delete(ctx, "token = ? ", input.Token)
		return nil, token.User.ID, err
	}

	data, err := s.createSession(ctx, createSessionInput{
		IsLogin:      false,
		UserID:       token.UserID,
		DeviceID:     token.DeviceID,
		Provider:     provider,
		RefreshToken: refreshToken,
	})

	return data, token.UserID, err
}

type AuthLogout struct {
	Token string
}

func (s AuthService) Logout(ctx context.Context, input AuthLogout) (int, error) {
	claims, err := tokenverify.GetClaimFromToken(input.Token, tokenverify.Secret(s.cfg.Get().AccessSecret))
	if err != nil {
		return 0, apperr.ErrUnauthenticated.WithErrorText("invalid token")
	}

	return claims.UserID, s.clearSession(ctx, claims.UserID, claims.DeviceID, true)
}

func (s AuthService) Account(ctx context.Context) (*model2.User, error) {
	userID, ok := mcontext.GetOpUserID(ctx)
	if !ok {
		return nil, apperr.ErrUnauthenticated.WithErrorText("operation user id is empty")
	}

	user, err := s.repositories.UserRepository.GetUserWithPermissions(ctx, "id = ?", userID)
	if err != nil {
		return nil, apperr.ErrUnauthenticated.WithError(err)
	}

	var prof *profile.IModel
	if s.profileService != nil {
		prof, err = s.profileService.GetById(ctx, userID)
		if err != nil {
			if !(apperr.Is(err, profile.ErrNotFound) || apperr.Is(err, apperr.ErrDBRecordNotFound)) {
				return nil, apperr.ErrUnauthenticated.WithError(err)
			}
		}
	}

	user.Profile = prof

	return user, nil
}

type createSessionInput struct {
	IsLogin      bool
	UserID       int
	DeviceID     int
	Provider     string
	RefreshToken string
}

func (s AuthService) createSession(ctx context.Context, input createSessionInput) (*model2.AuthToken, error) {
	claims := tokenverify.BuildClaims(input.UserID, input.DeviceID, s.cfg.Get().AccessExpire)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.cfg.Get().AccessSecret))
	if err != nil {
		return nil, apperr.ErrUnauthenticated.WithError(err)
	}

	doUpdate := []interface{}{"token", "expires_at", "updated_at", "provider"}
	doCreate := []interface{}{"logged_at"}
	if input.IsLogin {
		doUpdate = append(doUpdate, []string{"logged_at"})
	}

	var provider *string
	if input.Provider != "" {
		provider = &input.Provider
	}

	if _, err := s.repositories.TokenRepository.CreateOrUpdate(ctx, &model2.RefreshToken{
		UserID:    input.UserID,
		DeviceID:  input.DeviceID,
		Token:     input.RefreshToken,
		LoggedAt:  time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		ExpiresAt: time.Now().UTC().Add(s.cfg.Get().RefreshExpire),
		Provider:  provider,
	}, []interface{}{"user_id", "device_id"}, doUpdate, doCreate); err != nil {
		return nil, apperr.ErrUnauthenticated.WithError(err)
	}

	err = s.clearSession(ctx, input.UserID, input.DeviceID, false)
	if err != nil {
		return nil, apperr.ErrUnauthenticated.WithError(err)
	}

	if err = s.cache.Get().TokenCache.SetTokenFlagEx(ctx, input.UserID, input.DeviceID, tokenString, constant.NormalToken); err != nil {
		return nil, apperr.ErrUnauthenticated.WithError(err)
	}

	user, err := s.cache.Get().UserCache.GetUserInfo(ctx, input.UserID, func(ctx context.Context) (*model2.User, error) {
		user, err := s.repositories.UserRepository.GetUserWithPermissions(ctx, "id = ?", input.UserID)
		if err != nil {
			return nil, err
		}

		return user, nil
	})
	if err != nil {
		return nil, apperr.ErrUnauthenticated.WithError(err)
	}

	var prof *profile.IModel
	if s.profileService != nil {
		prof, err = s.profileService.GetById(ctx, input.UserID)
		if err != nil {
			if !(apperr.Is(err, profile.ErrNotFound) || apperr.Is(err, apperr.ErrDBRecordNotFound)) {
				return nil, apperr.ErrUnauthenticated.WithError(err)
			}
		}
	}

	user.Profile = prof

	return &model2.AuthToken{
		Auth: model2.Token{
			Token:        tokenString,
			RefreshToken: input.RefreshToken,
			ExpiresAt:    s.cfg.Get().RefreshExpire.Seconds(),
			TokenType:    "Bearer",
			UserID:       input.UserID,
		},
		User: *user,
	}, nil
}

func (s AuthService) clearSession(ctx context.Context, userID, deviceID int, clearRefresh bool) error {
	if clearRefresh {
		if err := s.repositories.TokenRepository.Delete(ctx, `user_id = ? AND device_id = ?`, userID, deviceID); err != nil {
			if !apperr.Is(err, apperr.ErrDBRecordNotFound) {
				return apperr.ErrUnauthenticated.WithError(err)
			}
		}
	}

	tokens, err := s.cache.Get().TokenCache.GetTokensWithoutError(ctx, userID, deviceID)
	if err != nil {
		return apperr.ErrUnauthenticated.WithError(err)
	}

	var deleteTokenKey []string
	for k := range tokens {
		deleteTokenKey = append(deleteTokenKey, k)
	}

	if len(deleteTokenKey) != 0 {
		err = s.cache.Get().TokenCache.DeleteTokenByUidPid(ctx, userID, deviceID, deleteTokenKey)
		if err != nil {
			return apperr.ErrUnauthenticated.WithError(err)
		}
	}

	if err := s.cache.Get().UserCache.DelUsersInfo(userID).ChainExecDel(ctx); err != nil {
		return apperr.ErrUnauthenticated.WithError(err)
	}

	return nil
}
