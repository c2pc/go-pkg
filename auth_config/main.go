package auth_config

import (
	"context"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/c2pc/go-pkg/v2/auth_config/configurator"
	"github.com/c2pc/go-pkg/v2/auth_config/internal/repository"
	"github.com/c2pc/go-pkg/v2/auth_config/internal/service"
	"github.com/c2pc/go-pkg/v2/auth_config/internal/transport/api/handler"
	"github.com/c2pc/go-pkg/v2/utils/mw"
)

type Configurator interface {
	service.IAuthConfigService
}

type Config interface {
	InitHandler(secured *gin.RouterGroup)
	GetService() service.IAuthConfigService
	SetConfig(ctx context.Context, key string, value configurator.Configurator) error
	Init(ctx context.Context) error
}

type AuthConfig struct {
	handler *handler.AuthConfigHandler
	service service.AuthConfigService
	db      *gorm.DB
}

func NewAuthConfig(db *gorm.DB, tr mw.ITransaction) Config {
	authConfigRepository := repository.NewAuthConfigRepository(db)
	authConfigService := service.NewAuthConfigService(authConfigRepository)
	authConfigHandler := handler.NewAuthConfigHandlers(authConfigService, tr)

	authConfig := &AuthConfig{
		handler: authConfigHandler,
		service: authConfigService,
		db:      db,
	}

	return authConfig
}

func (e *AuthConfig) InitHandler(secured *gin.RouterGroup) {
	e.handler.Init(secured)
}

func (e *AuthConfig) GetService() service.IAuthConfigService {
	return e.service
}

func (e *AuthConfig) SetConfig(ctx context.Context, key string, value configurator.Configurator) error {
	return e.db.Transaction(func(tx *gorm.DB) error {
		return e.service.Trx(e.db).SetConfig(ctx, key, value)
	})
}

func (e *AuthConfig) Init(ctx context.Context) error {
	return e.db.Transaction(func(tx *gorm.DB) error {
		return e.service.Trx(tx).Init(ctx)
	})
}
