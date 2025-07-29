package auth

import (
	"context"
	"errors"
	"strings"

	"strconv"
	"time"

	"github.com/c2pc/go-pkg/v2/auth/internal/cache/cachekey"
	"github.com/c2pc/go-pkg/v2/auth/profile"
	"github.com/c2pc/go-pkg/v2/utils/sso/ldap"
	"github.com/c2pc/go-pkg/v2/utils/sso/oidc"
	"github.com/c2pc/go-pkg/v2/utils/sso/saml"

	cache2 "github.com/c2pc/go-pkg/v2/auth/internal/cache"
	"github.com/c2pc/go-pkg/v2/auth/internal/database"
	model2 "github.com/c2pc/go-pkg/v2/auth/internal/model"
	"github.com/c2pc/go-pkg/v2/auth/internal/repository"
	"github.com/c2pc/go-pkg/v2/auth/internal/service"
	"github.com/c2pc/go-pkg/v2/auth/internal/transport/api/handler"
	"github.com/c2pc/go-pkg/v2/auth/internal/transport/api/middleware"
	"github.com/c2pc/go-pkg/v2/utils/cache"
	"github.com/c2pc/go-pkg/v2/utils/mcontext"
	"github.com/c2pc/go-pkg/v2/utils/model"
	"github.com/c2pc/go-pkg/v2/utils/mw"
	"github.com/c2pc/go-pkg/v2/utils/secret"
	"github.com/dtm-labs/rockscache"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type IAuth interface {
	InitHandler(engine *gin.Engine, api *gin.RouterGroup, handlers ...gin.HandlerFunc)
	Authenticate(c *gin.Context)
	CanPermission(c *gin.Context)
	GetAdminID() int
	LimiterMiddleware(c *gin.Context)
}

type SSO struct {
	LDAP ldap.Config
	OIDC oidc.Config
	SAML saml.Config
}

type Config struct {
	DB            *gorm.DB
	Rdb           redis.UniversalClient
	Transaction   mw.ITransaction
	Hasher        secret.Hasher
	AccessExpire  time.Duration
	RefreshExpire time.Duration
	AccessSecret  string
	Permissions   []model.Permission
	TTL           time.Duration
	MaxAttempts   int
	SSO           SSO
}

func New[Model profile.IModel, CreateInput, UpdateInput, UpdateProfileInput any](
	ctx context.Context,
	serviceName string,
	version string,
	cfg Config,
	prof *profile.Profile[Model, CreateInput, UpdateInput, UpdateProfileInput],
) (IAuth, error) {
	if serviceName == "" {
		return nil, errors.New("service name is required")
	}

	cachekey.SetServiceName(serviceName)

	model2.SetPermissions(cfg.Permissions)
	ctx = mcontext.WithOperationIDContext(ctx, strconv.Itoa(int(time.Now().UTC().Unix())))

	repositories := repository.NewRepositories(cfg.DB)
	admin, err := database.SeedersRun(ctx, cfg.DB, repositories, cfg.Hasher, model2.GetPermissionsKeys())
	if err != nil {
		return nil, err
	}

	rcClient := rockscache.NewClient(cfg.Rdb, cache.GetRocksCacheOptions())
	batchHandler := cache.NewBatchDeleterRedis(cfg.Rdb, cache.GetRocksCacheOptions())

	tokenCache := cache2.NewTokenCache(cfg.Rdb, cfg.AccessExpire)
	userCache := cache2.NewUserCache(cfg.Rdb, rcClient, batchHandler, cfg.AccessExpire)
	permissionCache := cache2.NewPermissionCache(cfg.Rdb, rcClient, batchHandler)
	limiterCache := cache2.NewLimiterCache(cfg.Rdb)

	var profileService profile.IProfileService[Model, CreateInput, UpdateInput, UpdateProfileInput]
	if prof != nil {
		profileService = prof.Service
	} else {
		profileService = nil
	}

	var profileTransformer profile.ITransformer[Model]
	if prof != nil {
		profileTransformer = prof.Transformer
	} else {
		profileTransformer = nil
	}

	var profileRequest profile.IRequest[CreateInput, UpdateInput, UpdateProfileInput]
	if prof != nil {
		profileRequest = prof.Request
	} else {
		profileRequest = nil
	}

	if profileService == nil || profileTransformer == nil || profileRequest == nil {
		profileService = nil
		profileTransformer = nil
		profileRequest = nil
	}

	ldapAuthService, err := ldap.NewAuthService(cfg.SSO.LDAP)
	if err != nil {
		return nil, err
	}

	if cfg.SSO.OIDC.Enabled {
		cfg.SSO.SAML.Enabled = false
	} else if cfg.SSO.SAML.Enabled {
		//TODO
	}

	cfg.SSO.OIDC.RootURL = strings.TrimRight(cfg.SSO.OIDC.RootURL, "/") + "/api/v1/auth/sso/callback"
	oidcAuthService, err := oidc.NewAuthService(ctx, cfg.SSO.OIDC)
	if err != nil {
		return nil, err
	}

	cfg.SSO.SAML.RootURL = strings.TrimRight(cfg.SSO.SAML.RootURL, "/") + "/api/v1/auth/sso/login"
	samlAuthService, err := saml.NewAuthService(ctx, cfg.SSO.SAML)
	if err != nil {
		return nil, err
	}

	authService := service.NewAuthService(profileService, repositories.UserRepository, repositories.TokenRepository,
		tokenCache, userCache, cfg.Hasher, cfg.AccessExpire, cfg.RefreshExpire, cfg.AccessSecret, ldapAuthService, oidcAuthService, samlAuthService)
	permissionService := service.NewPermissionService(repositories.PermissionRepository, permissionCache)
	roleService := service.NewRoleService(repositories.RoleRepository, repositories.PermissionRepository,
		repositories.RolePermissionRepository, repositories.UserRoleRepository, userCache, tokenCache)
	userService := service.NewUserService(profileService, repositories.UserRepository, repositories.RoleRepository,
		repositories.UserRoleRepository, userCache, tokenCache, cfg.Hasher)
	settingService := service.NewSettingService(repositories.SettingRepository)
	sessionService := service.NewSessionService(repositories.TokenRepository, tokenCache, userCache, cfg.RefreshExpire)
	filterService := service.NewFilterService(repositories.FilterRepository)
	versionService := service.NewVersionService(version, repositories.MigrationRepository)

	tokenMiddleware := middleware.NewTokenMiddleware(tokenCache, cfg.AccessSecret)
	permissionMiddleware := middleware.NewPermissionMiddleware(userCache, permissionCache, repositories.UserRepository,
		repositories.PermissionRepository)
	authLimiterMiddleware := middleware.NewAuthLimiterMiddleware(middleware.ConfigLimiter{
		MaxAttempts: cfg.MaxAttempts,
		TTL:         cfg.TTL,
	}, limiterCache)

	handlers := handler.NewHandlers[Model, CreateInput, UpdateInput, UpdateProfileInput](
		authService,
		permissionService,
		roleService,
		userService,
		settingService,
		filterService,
		sessionService,
		cfg.Transaction,
		tokenMiddleware,
		permissionMiddleware,
		profileTransformer,
		profileRequest,
		oidcAuthService,
		samlAuthService,
		versionService,
	)

	auth := Auth{
		handler:              handlers,
		tokenMiddleware:      tokenMiddleware,
		permissionMiddleware: permissionMiddleware,
		adminID:              admin.ID,
		limiterMiddleware:    authLimiterMiddleware,
	}

	go auth.startSessionCleaner(ctx, cfg.DB)

	return auth, nil
}

type Auth struct {
	handler              handler.IHandler
	tokenMiddleware      middleware.ITokenMiddleware
	permissionMiddleware middleware.IPermissionMiddleware
	limiterMiddleware    middleware.AuthMiddleware
	adminID              int
}

func (a Auth) InitHandler(engine *gin.Engine, api *gin.RouterGroup, handlers ...gin.HandlerFunc) {
	a.handler.Init(engine, api, handlers...)
}

func (a Auth) Authenticate(c *gin.Context) {
	a.tokenMiddleware.Authenticate(c)
}

func (a Auth) CanPermission(c *gin.Context) {
	a.permissionMiddleware.Can(c)
}

func (a Auth) GetAdminID() int {
	return a.adminID
}

func (a Auth) LimiterMiddleware(c *gin.Context) {
	a.limiterMiddleware.LimiterMiddleware(c)
}

func (a Auth) startSessionCleaner(ctx context.Context, db *gorm.DB) {
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
