package auth

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/c2pc/go-pkg/v2/auth/fx"
	"github.com/c2pc/go-pkg/v2/auth/internal/cache/cachekey"
	authConf "github.com/c2pc/go-pkg/v2/auth/internal/configurator"
	"github.com/c2pc/go-pkg/v2/auth/profile"
	"github.com/c2pc/go-pkg/v2/auth_config"
	"github.com/c2pc/go-pkg/v2/utils/model"
	"github.com/c2pc/go-pkg/v2/utils/mw"
	"github.com/c2pc/go-pkg/v2/utils/sso/ldap"
	"github.com/c2pc/go-pkg/v2/utils/sso/oidc"
	"github.com/c2pc/go-pkg/v2/utils/sso/saml"
	"github.com/redis/go-redis/v9"

	"github.com/c2pc/go-pkg/v2/auth/internal/database"
	model2 "github.com/c2pc/go-pkg/v2/auth/internal/model"
	"github.com/c2pc/go-pkg/v2/auth/internal/repository"
	"github.com/c2pc/go-pkg/v2/auth/internal/service"
	"github.com/c2pc/go-pkg/v2/auth/internal/transport/api/handler"
	"github.com/c2pc/go-pkg/v2/auth/internal/transport/api/middleware"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type IAuth interface {
	InitHandler(engine *gin.Engine, api *gin.RouterGroup, handlers ...gin.HandlerFunc)
	Authenticate(c *gin.Context)
	CanPermission(c *gin.Context)
	GetAdminID() int
	LimiterMiddleware(c *gin.Context)
}

type Auth struct {
	handler              handler.IHandler
	adminID              int
	permissionMiddleware *middleware.PermissionMiddleware
	cfg                  authConf.Cfg
	rdb                  redis.UniversalClient
	tokenMW              *middleware.TokenMiddleware
	limiterMW            *middleware.AuthMiddleware

	cacheHolder   *fx.CacheHolder
	ldapHolder    *fx.LDAPHolder
	oidcHolder    *fx.OIDCHolder
	samlHolder    *fx.SAMLHolder
	limiterHolder *fx.LimiterHolder
	authHolder    *fx.AuthHolder
}

type Input struct {
	DB           *gorm.DB
	Rdb          redis.UniversalClient
	Transaction  mw.ITransaction
	Permissions  []model.Permission
	Configurator auth_config.Config
}

func New(
	ctx context.Context,
	serviceName string,
	version string,
	input Input,
	prof *profile.Profile,
) (IAuth, error) {
	if serviceName == "" {
		return nil, errors.New("service name is required")
	}
	cachekey.SetServiceName(serviceName)

	authConfig := authConf.NewConfigurator()

	err := input.Configurator.SetConfig(ctx, "auth", authConfig)
	if err != nil {
		return nil, err
	}

	cfgByte, err := input.Configurator.GetService().GetWithoutTransform(ctx, "auth")
	if err != nil {
		return nil, err
	}

	cfgW, err := authConfig.Unmarshal(cfgByte.Value)
	if err != nil {
		return nil, err
	}

	cfg := cfgW.(authConf.Cfg)

	model2.SetPermissions(input.Permissions)

	repositories := repository.NewRepositories(input.DB)
	admin, err := database.SeedersRun(ctx, input.DB, repositories, model2.GetPermissionsKeys())
	if err != nil {
		return nil, err
	}

	var profileService profile.IProfileService
	var profileTransformer profile.ITransformer
	var profileRequest profile.IRequest
	if prof != nil {
		profileService = prof.Service
		profileTransformer = prof.Transformer
		profileRequest = prof.Request
	}

	accessTokenTTL := time.Duration(cfg.AccessTokenTTL) * time.Minute
	refreshExpire := time.Duration(cfg.AccessTokenTTL) * time.Minute
	accessSecret := cfg.Key

	cacheHolder := fx.NewCacheHolder(input.Rdb, accessTokenTTL)
	ldapHolder := fx.NewLDAPHolder(len(cfg.LDAP) != 0, getLdapConfig(cfg.LDAP))
	oidcHolder, err := fx.NewOIDCHolder(ctx, getOIDCConfig(cfg.SSO.Enabled == "oidc", cfg.SSO.OIDC))
	if err != nil {
		return nil, err
	}
	samlHolder, err := fx.NewSAMLHolder(ctx, getSAMLConfig(cfg.SSO.Enabled == "saml", cfg.SSO.SAML))
	if err != nil {
		return nil, err
	}

	limiterHolder := fx.NewLimiterHolder(getLimiterConfig(cfg.Limiter))
	authHolder := fx.NewAuthHolder(accessTokenTTL, refreshExpire, string(accessSecret))
	tokenMW := middleware.NewTokenMiddleware(cacheHolder, repositories, authHolder)
	limiterMW := middleware.NewAuthLimiterMiddleware(cacheHolder, limiterHolder)

	authService := service.NewAuthService(profileService, repositories, cacheHolder, authHolder, ldapHolder, oidcHolder, samlHolder)
	permissionService := service.NewPermissionService(repositories, cacheHolder)
	roleService := service.NewRoleService(repositories, cacheHolder)
	userService := service.NewUserService(profileService, repositories, cacheHolder)
	settingService := service.NewSettingService(repositories)
	sessionService := service.NewSessionService(repositories, cacheHolder)
	filterService := service.NewFilterService(repositories)
	versionService := service.NewVersionService(version, repositories)

	permissionMiddleware := middleware.NewPermissionMiddleware(cacheHolder.Get(), repositories)

	handlers := handler.NewHandlers(
		authService,
		permissionService,
		roleService,
		userService,
		settingService,
		filterService,
		sessionService,
		input.Transaction,
		tokenMW,
		permissionMiddleware,
		profileTransformer,
		profileRequest,
		oidcHolder,
		samlHolder,
		versionService,
	)

	auth := &Auth{
		handler:              handlers,
		adminID:              admin.ID,
		cacheHolder:          cacheHolder,
		ldapHolder:           ldapHolder,
		oidcHolder:           oidcHolder,
		samlHolder:           samlHolder,
		limiterMW:            limiterMW,
		tokenMW:              tokenMW,
		authHolder:           authHolder,
		limiterHolder:        limiterHolder,
		permissionMiddleware: permissionMiddleware,
		cfg:                  cfg,
		rdb:                  input.Rdb,
	}

	go auth.startSessionCleaner(ctx, input.DB)
	go auth.configWatcher(ctx, authConfig)

	return auth, nil
}

func (a *Auth) InitHandler(engine *gin.Engine, api *gin.RouterGroup, handlers ...gin.HandlerFunc) {
	a.handler.Init(engine, api, handlers...)
}
func (a *Auth) Authenticate(c *gin.Context)  { a.tokenMW.Authenticate(c) }
func (a *Auth) CanPermission(c *gin.Context) { a.permissionMiddleware.Can(c) }
func (a *Auth) GetAdminID() int              { return a.adminID }
func (a *Auth) LimiterMiddleware(c *gin.Context) {
	a.limiterMW.LimiterMiddleware(c)
}
func (a *Auth) startSessionCleaner(ctx context.Context, db *gorm.DB) {
	tm := time.NewTicker(10 * time.Minute)
	defer tm.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-tm.C:
			db.WithContext(ctx).Where("expires_at < ?", time.Now().UTC()).Delete(&model2.RefreshToken{})
		}
	}
}

func getLdapConfig(c []authConf.CfgLDAP) map[string]ldap.Config {
	cfg := make(map[string]ldap.Config)
	for _, conf := range c {
		addrs := make([]string, len(conf.Addrs))
		for i, addr := range conf.Addrs {
			scheme := "ldap://"
			if addr.Secured {
				scheme = "ldaps://"
			}
			addrs[i] = scheme + addr.Addr
		}
		cfg[conf.Domain] = ldap.Config{
			Addrs:  addrs,
			Domain: conf.Domain,
		}
	}
	return cfg
}

func getOIDCConfig(enabled bool, cfg *authConf.CfgOIDC) oidc.Config {
	if cfg == nil {
		return oidc.Config{
			Enabled: false,
		}
	}

	return oidc.Config{
		Enabled:           enabled,
		ConfigURL:         cfg.ConfigURL,
		ClientID:          cfg.ClientID,
		ClientSecret:      string(cfg.ClientSecret),
		RootURL:           cfg.RootURL,
		LoginAttr:         cfg.LoginAttr,
		ValidRedirectURLs: cfg.ValidRedirectURLs,
	}
}

func getSAMLConfig(enabled bool, cfg *authConf.CfgSAML) saml.Config {
	if cfg == nil {
		return saml.Config{
			Enabled: false,
		}
	}

	return saml.Config{
		Enabled:           enabled,
		MetaDataURL:       cfg.MetaDataFile,
		CertFile:          cfg.CertFile,
		KeyFile:           cfg.KeyFile,
		RootURL:           cfg.RootURL,
		LoginAttr:         cfg.LoginAttr,
		ValidRedirectURLs: cfg.ValidRedirectURLs,
	}
}

func getLimiterConfig(cfg authConf.CfgLimiter) fx.ConfigLimiter {
	return fx.ConfigLimiter{
		MaxAttempts: cfg.MaxAttempts,
		TTL:         time.Duration(cfg.TTL) * time.Second,
	}
}

func (a *Auth) configWatcher(ctx context.Context, authConfig *authConf.Configurator) {
	for {
		select {
		case <-ctx.Done():
			return
		case newCfg := <-authConfig.Watch():
			oldCfg := a.cfg

			accessTokenTTL := time.Duration(newCfg.AccessTokenTTL) * time.Minute
			refreshExpire := time.Duration(newCfg.AccessTokenTTL) * time.Minute
			accessSecret := newCfg.Key

			if newCfg.AccessTokenTTL != oldCfg.AccessTokenTTL || newCfg.RefreshTokenTTL != oldCfg.RefreshTokenTTL || string(newCfg.Key) != string(oldCfg.Key) {
				a.cacheHolder.Reload(a.rdb, accessTokenTTL)
				a.authHolder.Reload(accessTokenTTL, refreshExpire, string(accessSecret))
			}

			if newCfg.Limiter.MaxAttempts != oldCfg.Limiter.MaxAttempts || newCfg.Limiter.TTL != oldCfg.Limiter.TTL {
				a.limiterHolder.Reload(getLimiterConfig(newCfg.Limiter))
			}

			m1, _ := json.Marshal(newCfg.LDAP)
			m2, _ := json.Marshal(oldCfg.LDAP)

			if string(m1) != string(m2) || len(newCfg.LDAP) != len(oldCfg.LDAP) {
				a.ldapHolder.Reload(len(newCfg.LDAP) != 0, getLdapConfig(newCfg.LDAP))
			}

			m3, _ := json.Marshal(newCfg.SSO)
			m4, _ := json.Marshal(oldCfg.SSO)
			if string(m3) != string(m4) {
				_ = a.oidcHolder.Reload(ctx, getOIDCConfig(newCfg.SSO.Enabled == "oidc", newCfg.SSO.OIDC))
				_ = a.samlHolder.Reload(ctx, getSAMLConfig(newCfg.SSO.Enabled == "saml", newCfg.SSO.SAML))
			}

			a.cfg = newCfg
		}
	}

}
