package auth

import (
	"context"
	"github.com/c2pc/go-pkg/v2/auth/profile"
	"strconv"
	"time"

	cache3 "github.com/c2pc/go-pkg/v2/auth/internal/cache"
	"github.com/c2pc/go-pkg/v2/auth/internal/database"
	model2 "github.com/c2pc/go-pkg/v2/auth/internal/model"
	"github.com/c2pc/go-pkg/v2/auth/internal/repository"
	service2 "github.com/c2pc/go-pkg/v2/auth/internal/service"
	"github.com/c2pc/go-pkg/v2/auth/internal/transport/api/handler"
	middleware2 "github.com/c2pc/go-pkg/v2/auth/internal/transport/api/middleware"
	cache2 "github.com/c2pc/go-pkg/v2/utils/cache"
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
	InitHandler(api *gin.RouterGroup)
	Authenticate(c *gin.Context)
	CanPermission(c *gin.Context)
	GetAdminID() int
	LimiterMiddleware(c *gin.Context)
}

type Config struct {
	Debug         string
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
}

func New[Model, CreateInput, UpdateInput, UpdateProfileInput any](cfg Config, prof *profile.Profile[Model, CreateInput, UpdateInput, UpdateProfileInput]) (IAuth, error) {
	model2.SetPermissions(cfg.Permissions)
	ctx := mcontext.WithOperationIDContext(context.Background(), strconv.Itoa(int(time.Now().UTC().Unix())))

	repositories := repository.NewRepositories(cfg.DB)
	admin, err := database.SeedersRun(ctx, cfg.DB, repositories, cfg.Hasher, model2.GetPermissionsKeys())
	if err != nil {
		return nil, err
	}

	rcClient := rockscache.NewClient(cfg.Rdb, cache2.GetRocksCacheOptions())
	batchHandler := cache2.NewBatchDeleterRedis(cfg.Rdb, cache2.GetRocksCacheOptions())

	tokenCache := cache3.NewTokenCache(cfg.Rdb, cfg.AccessExpire)
	userCache := cache3.NewUserCache(cfg.Rdb, rcClient, batchHandler, cfg.AccessExpire)
	permissionCache := cache3.NewPermissionCache(cfg.Rdb, rcClient, batchHandler)
	limiterCache := cache3.NewLimiterCache(cfg.Rdb)

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

	authService := service2.NewAuthService(profileService, repositories.UserRepository, repositories.TokenRepository, tokenCache, userCache, cfg.Hasher, cfg.AccessExpire, cfg.RefreshExpire, cfg.AccessSecret)
	permissionService := service2.NewPermissionService(repositories.PermissionRepository, permissionCache)
	roleService := service2.NewRoleService(repositories.RoleRepository, repositories.PermissionRepository, repositories.RolePermissionRepository, repositories.UserRoleRepository, userCache, tokenCache)
	userService := service2.NewUserService(profileService, repositories.UserRepository, repositories.RoleRepository, repositories.UserRoleRepository, userCache, tokenCache, cfg.Hasher)
	settingService := service2.NewSettingService(repositories.SettingRepository)
	sessionService := service2.NewSessionService(repositories.TokenRepository, tokenCache, userCache, cfg.RefreshExpire)

	tokenMiddleware := middleware2.NewTokenMiddleware(tokenCache, cfg.AccessSecret)
	permissionMiddleware := middleware2.NewPermissionMiddleware(userCache, permissionCache, repositories.UserRepository, repositories.PermissionRepository, cfg.Debug)
	authLimiterMiddleware := middleware2.NewAuthLimiterMiddleware(middleware2.ConfigLimiter{
		MaxAttempts: cfg.MaxAttempts,
		TTL:         cfg.TTL,
	}, limiterCache, cfg.Debug)
	handlers := handler.NewHandlers[Model, CreateInput, UpdateInput, UpdateProfileInput](authService, permissionService, roleService, userService, settingService, sessionService, cfg.Transaction, tokenMiddleware, permissionMiddleware, profileTransformer, profileRequest)

	return Auth{
		handler:              handlers,
		tokenMiddleware:      tokenMiddleware,
		permissionMiddleware: permissionMiddleware,
		adminID:              admin.ID,
		limiterMiddleware:    authLimiterMiddleware,
	}, nil
}

type Auth struct {
	handler              handler.IHandler
	tokenMiddleware      middleware2.ITokenMiddleware
	permissionMiddleware middleware2.IPermissionMiddleware
	limiterMiddleware    middleware2.AuthMiddleware
	adminID              int
}

func (a Auth) InitHandler(api *gin.RouterGroup) {
	a.handler.Init(api)
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
