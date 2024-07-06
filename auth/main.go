package auth

import (
	"context"
	"github.com/c2pc/go-pkg/v2/auth/cache"
	"github.com/c2pc/go-pkg/v2/auth/database"
	model2 "github.com/c2pc/go-pkg/v2/auth/model"
	"github.com/c2pc/go-pkg/v2/auth/repository"
	"github.com/c2pc/go-pkg/v2/auth/service"
	"github.com/c2pc/go-pkg/v2/auth/transport/api/handler"
	"github.com/c2pc/go-pkg/v2/auth/transport/api/middleware"
	cache2 "github.com/c2pc/go-pkg/v2/utils/cache"
	"github.com/c2pc/go-pkg/v2/utils/mcontext"
	"github.com/c2pc/go-pkg/v2/utils/model"
	"github.com/c2pc/go-pkg/v2/utils/mw"
	"github.com/c2pc/go-pkg/v2/utils/secret"
	"github.com/dtm-labs/rockscache"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	"strconv"
	"time"
)

type IAuth interface {
	InitHandler(api *gin.RouterGroup)
	Authenticate(c *gin.Context)
	CanPermission(c *gin.Context)
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
}

func New(cfg Config) (IAuth, error) {
	model2.SetPermissions(cfg.Permissions)
	ctx := mcontext.WithOperationIDContext(context.Background(), strconv.Itoa(int(time.Now().Unix())))

	repositories := repository.NewRepositories(cfg.DB)
	err := database.SeedersRun(ctx, cfg.DB, repositories, cfg.Hasher, model2.GetPermissionsKeys())
	if err != nil {
		return nil, err
	}

	rcClient := rockscache.NewClient(cfg.Rdb, cache2.GetRocksCacheOptions())
	batchHandler := cache2.NewBatchDeleterRedis(cfg.Rdb, cache2.GetRocksCacheOptions())

	tokenCache := cache.NewTokenCache(cfg.Rdb, cfg.AccessExpire)
	userCache := cache.NewUserCache(cfg.Rdb, rcClient, batchHandler, cfg.AccessExpire)
	permissionCache := cache.NewPermissionCache(cfg.Rdb, rcClient, batchHandler)

	authService := service.NewAuthService(repositories.UserRepository, repositories.TokenRepository, tokenCache, userCache, cfg.Hasher, cfg.AccessExpire, cfg.RefreshExpire, cfg.AccessSecret)
	permissionService := service.NewPermissionService(repositories.PermissionRepository, permissionCache)
	roleService := service.NewRoleService(repositories.RoleRepository, repositories.PermissionRepository, repositories.RolePermissionRepository, repositories.UserRoleRepository, userCache)
	userService := service.NewUserService(repositories.UserRepository, repositories.RoleRepository, repositories.UserRoleRepository, userCache, tokenCache, cfg.Hasher)

	tokenMiddleware := middleware.NewTokenMiddleware(tokenCache, cfg.AccessSecret)
	permissionMiddleware := middleware.NewPermissionMiddleware(userCache, repositories.UserRepository)
	handlers := handler.NewHandlers(authService, permissionService, roleService, userService, cfg.Transaction, tokenMiddleware, permissionMiddleware)

	return Auth{
		handler:              handlers,
		tokenMiddleware:      tokenMiddleware,
		permissionMiddleware: permissionMiddleware,
	}, nil
}

type Auth struct {
	handler              handler.IHandler
	tokenMiddleware      middleware.ITokenMiddleware
	permissionMiddleware middleware.IPermissionMiddleware
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
