package service

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/c2pc/go-pkg/v2/auth/internal/cache/cachekey"
	"github.com/c2pc/go-pkg/v2/auth/internal/fx"
	"github.com/c2pc/go-pkg/v2/auth/internal/i18n"
	"github.com/c2pc/go-pkg/v2/auth/internal/model"
	"github.com/c2pc/go-pkg/v2/auth/internal/repository"
	"github.com/c2pc/go-pkg/v2/auth/profile"
	"github.com/c2pc/go-pkg/v2/utils/apperr"
	"github.com/c2pc/go-pkg/v2/utils/apperr/code"
	"github.com/c2pc/go-pkg/v2/utils/constant"
	i18n2 "github.com/c2pc/go-pkg/v2/utils/i18n"
	"github.com/c2pc/go-pkg/v2/utils/mcontext"
	"github.com/c2pc/go-pkg/v2/utils/secret"
	"github.com/c2pc/go-pkg/v2/utils/sso"
	"github.com/c2pc/go-pkg/v2/utils/syslog"
	"github.com/c2pc/go-pkg/v2/utils/tokenverify"
	"github.com/golang-jwt/jwt/v4"
	"github.com/rs/xid"
	"gorm.io/gorm"
)

var brSec = []byte("*&*&UYV8fu98agV*Y(&*")

var (
	ErrAuthNoAccess    = apperr.New("auth_no_access", apperr.WithTextTranslate(i18n2.ErrUnauthenticated), apperr.WithCode(code.Unauthenticated))
	ErrAuthBlocked     = apperr.New("auth_blocked", apperr.WithTextTranslate(i18n2.ErrUnauthenticated), apperr.WithCode(code.Unauthenticated))
	ErrSSONotSupported = apperr.New("sso_not_supported", apperr.WithTextTranslate(i18n.ErrSSONotSupported), apperr.WithCode(code.Unauthenticated))
)

type IAuthService interface {
	Trx(db *gorm.DB) IAuthService
	Login(ctx context.Context, input AuthLogin) (*model.AuthToken, int64, error)
	Refresh(ctx context.Context, input AuthRefresh) (*model.AuthToken, int64, error)
	Logout(ctx context.Context, input AuthLogout) (int64, error)
	Account(ctx context.Context) (*model.User, error)
	SSO(ctx context.Context, input SSO) (*model.AuthToken, int64, error)
	GetOptions(ctx context.Context) map[string]interface{}
}

type AuthService struct {
	profileService profile.IProfileService
	repositories   repository.Repositories
	cache          *fx.CacheHolder
	cfg            *fx.AuthHolder
	ldapAuth       *fx.LDAPHolder
	oidcAuth       *fx.OIDCHolder
	samlAuth       *fx.SAMLHolder
	limiter        *fx.LimiterHolder
}

func NewAuthService(
	profileService profile.IProfileService,
	repositories repository.Repositories,
	cache *fx.CacheHolder,
	cfg *fx.AuthHolder,
	ldapAuth *fx.LDAPHolder,
	oidcAuth *fx.OIDCHolder,
	samlAuth *fx.SAMLHolder,
	limiter *fx.LimiterHolder,
) AuthService {
	return AuthService{
		profileService: profileService,
		repositories:   repositories,
		cache:          cache,
		cfg:            cfg,
		ldapAuth:       ldapAuth,
		oidcAuth:       oidcAuth,
		samlAuth:       samlAuth,
		limiter:        limiter,
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
	DeviceID int
}

func (s AuthService) Login(ctx context.Context, input AuthLogin) (authToken *model.AuthToken, userID int64, err error) {
	defer func() {
		success := err == nil
		eventName := "Вход пользователя в систему"
		msg := fmt.Sprintf("%s: %s", eventName, input.Login)
		if err != nil {
			eventName = "Неуспешный вход пользователя в систему"
			msg = fmt.Sprintf("%s: %s: %s", eventName, input.Login, err.Error())
		}

		syslog.Write(ctx, syslog.Record{EventID: "signin-admin", EventName: eventName, Severity: syslog.SeverityLow, Success: success}, msg)
	}()

	user, err := s.repositories.UserRepository.With("roles").Find(ctx, "login = ?", input.Login)
	if err != nil {
		if apperr.Is(err, apperr.ErrDBRecordNotFound) {
			return nil, 0, apperr.ErrUnauthenticated.WithErrorText("Администратор с таким логином не найден")
		}
		return nil, 0, apperr.ErrUnauthenticated.WithError(err)
	}

	_, err = s.repositories.UserBlockedRepository.Find(ctx, `user_id = ?`, user.ID)
	if err != nil {
		if !apperr.Is(err, apperr.ErrDBRecordNotFound) {
			return nil, 0, apperr.ErrUnauthenticated.WithError(err)
		}
	} else {
		return nil, user.ID, ErrAuthBlocked.WithErrorText("Администратор временно заблокирован")
	}

	if user.Blocked {
		return nil, user.ID, ErrAuthBlocked.WithErrorText("Администратор заблокирован")
	}

	var domain string
	spl := strings.SplitN(user.Login, "@", 2)
	if len(spl) == 2 {
		domain = spl[1]
	}

	var provider, refreshToken string
	if user.IsDomain {
		if domain == "" {
			return nil, user.ID, apperr.ErrUnauthenticated.WithErrorText("Домен не содержится в логине")
		}

		if s.ldapAuth.Get() != nil && s.ldapAuth.Get().IsEnabled() {
			err = s.ldapAuth.Get().CheckAuth(domain, input.Login, input.Password)
			if err != nil {
				if apperr.Is(err, apperr.ErrUnauthenticated) {
					_ = s.checkBlock(ctx, user.ID)
				}
				return nil, user.ID, apperr.ErrUnauthenticated.WithError(err)
			}

			provider = "ldap"
			refreshToken = xid.New().String()
		} else {
			return nil, user.ID, apperr.ErrUnauthenticated.WithErrorText("Доменная аутентификация не настроена")
		}
	} else {
		if user.Password != nil {
			if !secret.HasherSecret.HashMatchesString(*user.Password, model.GeneratePassword(input.Password, user.ID)) {
				if !secret.HasherSecret.HashMatchesString(*user.Password, input.Password) {
					_ = s.checkBlock(ctx, user.ID)
					return nil, user.ID, apperr.ErrUnauthenticated.WithErrorText("Пароль не совпадает")
				}
			}
			refreshToken = xid.New().String()
		} else {
			return nil, user.ID, apperr.ErrUnauthenticated.WithErrorText("Пустой пароль")
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

func (s AuthService) SSO(ctx context.Context, input SSO) (authToken *model.AuthToken, userID int64, err error) {
	defer func() {
		success := err == nil
		eventName := "Вход пользователя в систему"
		msg := fmt.Sprintf("%s: %s", eventName, input.Login)
		if err != nil {
			eventName = "Неуспешный вход пользователя в систему"
			msg = fmt.Sprintf("%s: %s: %s", eventName, input.Login, err.Error())
		}

		syslog.Write(ctx, syslog.Record{EventID: "signin-sso-admin", EventName: eventName, Severity: syslog.SeverityLow, Success: success}, msg)
	}()

	user, err := s.repositories.UserRepository.Find(ctx, "login = ?", input.Login)
	if err != nil {
		if apperr.Is(err, apperr.ErrDBRecordNotFound) {
			return nil, 0, ErrAuthNoAccess.WithErrorText("Администратор с таким логином не найден")
		}
		return nil, 0, ErrAuthNoAccess.WithError(err)
	}

	if user.Blocked {
		return nil, user.ID, ErrAuthBlocked.WithErrorText("Администратор заблокирован")
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

	if input.Provider == sso.OIDC && input.RefreshToken == "" {
		input.Provider = ""
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

func (s AuthService) Refresh(ctx context.Context, input AuthRefresh) (*model.AuthToken, int64, error) {
	token, err := s.repositories.TokenRepository.With("user").Find(ctx, "token = ? AND device_id = ?", input.Token, input.DeviceID)
	if err != nil {
		if apperr.Is(err, apperr.ErrDBRecordNotFound) {
			return nil, 0, ErrAuthNoAccess.WithErrorText("Токен не найден")
		}
		return nil, 0, apperr.ErrUnauthenticated.WithError(err)
	}

	if token.User.Blocked {
		err := s.clearSession(ctx, token.UserID, token.DeviceID, true)
		if err != nil {
			return nil, token.User.ID, ErrAuthBlocked.WithError(err)
		}

		return nil, token.User.ID, ErrAuthBlocked.WithErrorText("Администратор заблокирован")
	}

	var provider, refreshToken string
	err = func() error {
		if token.Provider == nil || (token.Provider != nil && *token.Provider == "ldap") {
			if time.Now().UTC().After(token.ExpiresAt) {
				return apperr.ErrUnauthenticated.WithErrorText("Время жизни токена истекло")
			}

			if token.Provider != nil {
				provider = *token.Provider
			}
			refreshToken = xid.New().String()
		} else if token.Provider != nil {
			if *token.Provider == sso.OIDC && s.oidcAuth.Get().IsEnabled() {
				oidcToken, err := s.oidcAuth.Get().Refresh(ctx, input.Token)
				if err != nil {
					_ = s.repositories.TokenRepository.Delete(ctx, "token = ? ", input.Token)
					return apperr.ErrUnauthenticated.WithError(err)
				}
				provider = sso.OIDC
				refreshToken = oidcToken.IDToken.RefreshToken
			} else if *token.Provider == sso.SAML && s.samlAuth.Get().IsEnabled() {
				if time.Now().UTC().After(token.ExpiresAt) {
					return apperr.ErrUnauthenticated.WithErrorText("Время жизни токена истекло")
				}

				provider = sso.SAML
				refreshToken = xid.New().String()
			} else {
				return apperr.ErrUnauthenticated.WithError(ErrSSONotSupported)
			}
		} else {
			return apperr.ErrUnauthenticated.WithErrorText("Неизвестный провайдер")
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

func (s AuthService) Logout(ctx context.Context, input AuthLogout) (userID int64, err error) {
	var login string
	defer func() {
		success := err == nil
		eventName := "Выход пользователя из системы"
		msg := fmt.Sprintf("%s: %s", eventName, login)
		if err != nil {
			eventName = "Неуспешный выход пользователя из системы"
			msg = fmt.Sprintf("%s: %s: %s", eventName, login, err.Error())
		}

		syslog.Write(ctx, syslog.Record{EventID: "signout-admin", EventName: eventName, Severity: syslog.SeverityLow, Success: success}, msg)
	}()

	claims, err := tokenverify.GetClaimFromToken(input.Token, tokenverify.Secret(s.cfg.Get().AccessSecret))
	if err != nil {
		return 0, apperr.ErrUnauthenticated.WithErrorText("Неправильный токен")
	}

	user, err := s.repositories.UserRepository.Find(ctx, "id = ?", claims.UserID)
	if err != nil {
		if apperr.Is(err, apperr.ErrDBRecordNotFound) {
			return claims.UserID, apperr.ErrUnauthenticated.WithErrorText("Администратор с таким ID не найден")
		}
		return claims.UserID, apperr.ErrUnauthenticated.WithError(err)
	}
	login = user.Login

	return claims.UserID, s.clearSession(ctx, claims.UserID, claims.DeviceID, true)
}

func (s AuthService) Account(ctx context.Context) (*model.User, error) {
	userID, ok := mcontext.GetOpUserID(ctx)
	if !ok {
		return nil, apperr.ErrUnauthenticated.WithErrorText("operation user id is empty")
	}

	user, err := s.repositories.UserRepository.GetUserWithPermissions(ctx, "id = ?", userID)
	if err != nil {
		return nil, apperr.ErrUnauthenticated.WithError(err)
	}

	var prof profile.IModel
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

func (s AuthService) GetOptions(ctx context.Context) map[string]interface{} {
	options := make(map[string]interface{})
	domains := make([]string, 0)
	ssoDesc := ""

	if s.ldapAuth.Get().IsEnabled() {
		cfg := s.ldapAuth.Get().GetConfigs()
		for domain := range cfg {
			domains = append(domains, domain)
		}
	}
	options["domains"] = domains

	if s.oidcAuth.Get().IsEnabled() {
		ssoDesc = s.oidcAuth.Get().GetDescription()
	} else if s.samlAuth.Get().IsEnabled() {
		ssoDesc = s.samlAuth.Get().GetDescription()
	}

	if ssoDesc != "" {
		options["sso"] = ssoDesc
	} else {
		options["sso"] = nil
	}

	return options
}

type createSessionInput struct {
	IsLogin      bool
	UserID       int64
	DeviceID     int
	Provider     string
	RefreshToken string
}

func (s AuthService) createSession(ctx context.Context, input createSessionInput) (*model.AuthToken, error) {
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

	if _, err := s.repositories.TokenRepository.CreateOrUpdate(ctx, &model.RefreshToken{
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

	user, err := s.cache.Get().UserCache.GetUserInfo(ctx, input.UserID, func(ctx context.Context) (*model.User, error) {
		user, err := s.repositories.UserRepository.GetUserWithPermissions(ctx, "id = ?", input.UserID)
		if err != nil {
			return nil, err
		}

		return user, nil
	})
	if err != nil {
		return nil, apperr.ErrUnauthenticated.WithError(err)
	}

	var prof profile.IModel
	if s.profileService != nil {
		prof, err = s.profileService.GetById(ctx, input.UserID)
		if err != nil {
			if !(apperr.Is(err, profile.ErrNotFound) || apperr.Is(err, apperr.ErrDBRecordNotFound)) {
				return nil, apperr.ErrUnauthenticated.WithError(err)
			}
		}
	}

	user.Profile = prof

	return &model.AuthToken{
		Auth: model.Token{
			Token:        tokenString,
			RefreshToken: input.RefreshToken,
			ExpiresAt:    s.cfg.Get().RefreshExpire.Seconds(),
			TokenType:    "Bearer",
			UserID:       input.UserID,
		},
		User: *user,
	}, nil
}

func (s AuthService) clearSession(ctx context.Context, userID int64, deviceID int, clearRefresh bool) error {
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

func (s AuthService) checkBlock(ctx context.Context, userID int64) error {
	key := cachekey.GetUsernameKey() + strconv.FormatInt(userID, 10)

	attempts, err := s.cache.Get().LimiterCache.GetAttempts(ctx, key)
	if err != nil {
		return err
	}

	if attempts >= s.limiter.Get().MaxAttempts {
		_, err = s.repositories.UserBlockedRepository.CreateOrUpdate(ctx,
			&model.UserBlocked{BlockedAt: time.Now().UTC(), UserID: userID},
			[]interface{}{"user_id"}, []interface{}{"blocked_at"}, []interface{}{})
		if err != nil {
			return err
		}

		_ = s.cache.Get().LimiterCache.ResetAttempts(ctx, key)
		return nil
	}

	_, err = s.cache.Get().LimiterCache.IncrAttempts(ctx, key, s.limiter.Get().TTL)
	return err
}
