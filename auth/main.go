package auth

import (
	"context"
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
}

func New(cfg Config) (IAuth, error) {
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

	authService := service2.NewAuthService(repositories.UserRepository, repositories.TokenRepository, tokenCache, userCache, cfg.Hasher, cfg.AccessExpire, cfg.RefreshExpire, cfg.AccessSecret)
	permissionService := service2.NewPermissionService(repositories.PermissionRepository, permissionCache)
	roleService := service2.NewRoleService(repositories.RoleRepository, repositories.PermissionRepository, repositories.RolePermissionRepository, repositories.UserRoleRepository, userCache, tokenCache)
	userService := service2.NewUserService(repositories.UserRepository, repositories.RoleRepository, repositories.UserRoleRepository, userCache, tokenCache, cfg.Hasher)
	settingService := service2.NewSettingService(repositories.SettingRepository)
	sessionService := service2.NewSessionService(repositories.TokenRepository, tokenCache, userCache, cfg.RefreshExpire)

	tokenMiddleware := middleware2.NewTokenMiddleware(tokenCache, cfg.AccessSecret)
	permissionMiddleware := middleware2.NewPermissionMiddleware(userCache, permissionCache, repositories.UserRepository, repositories.PermissionRepository, cfg.Debug)
	handlers := handler.NewHandlers(authService, permissionService, roleService, userService, settingService, sessionService, cfg.Transaction, tokenMiddleware, permissionMiddleware)

	return Auth{
		handler:              handlers,
		tokenMiddleware:      tokenMiddleware,
		permissionMiddleware: permissionMiddleware,
		adminID:              admin.ID,
	}, nil
}

type Auth struct {
	handler              handler.IHandler
	tokenMiddleware      middleware2.ITokenMiddleware
	permissionMiddleware middleware2.IPermissionMiddleware
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
