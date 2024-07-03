package auth

import (
	"github.com/c2pc/go-pkg/v2/auth/cache"
	"github.com/c2pc/go-pkg/v2/auth/repository"
	"github.com/c2pc/go-pkg/v2/auth/service"
	"github.com/c2pc/go-pkg/v2/auth/transport/api/handler"
	"github.com/c2pc/go-pkg/v2/auth/transport/api/middleware"
	cache2 "github.com/c2pc/go-pkg/v2/utils/cache"
	"github.com/c2pc/go-pkg/v2/utils/secret"
	"github.com/c2pc/go-pkg/v2/utils/transaction"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	"time"
)

type IAuth interface {
	GetHandler() handler.IHandler
	GetTokenMiddleware() middleware.ITokenMiddleware
}

type Config struct {
	DB            *gorm.DB
	Rdb           redis.UniversalClient
	Transaction   transaction.ITransaction
	Hasher        secret.Hasher
	AccessExpire  time.Duration
	RefreshExpire time.Duration
	AccessSecret  string
}

func New(cfg Config) IAuth {
	userRepo := repository.NewUserRepository(cfg.DB)
	tokenRepo := repository.NewTokenRepository(cfg.DB)

	tokenCache := cache.NewTokenCache(cfg.Rdb, cfg.AccessExpire)
	userCache := cache.NewUserCache(cfg.Rdb, cfg.AccessExpire, cache2.GetRocksCacheOptions())

	authService := service.NewAuthService(userRepo, tokenRepo, tokenCache, userCache, cfg.Hasher, cfg.AccessExpire, cfg.RefreshExpire, cfg.AccessSecret)

	tokenMiddleware := middleware.NewTokenMiddleware(tokenCache, cfg.AccessSecret)
	handlers := handler.NewHandlers(authService, cfg.Transaction, tokenMiddleware)

	return Auth{
		Handler:         handlers,
		TokenMiddleware: tokenMiddleware,
	}
}

type Auth struct {
	Handler         handler.IHandler
	TokenMiddleware middleware.ITokenMiddleware
}

func (a Auth) GetHandler() handler.IHandler {
	return a.Handler
}

func (a Auth) GetTokenMiddleware() middleware.ITokenMiddleware {
	return a.TokenMiddleware
}
