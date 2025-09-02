package auth_config

import (
	"context"
	"errors"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/c2pc/go-pkg/v2/auth_config/internal/repository"
	"github.com/c2pc/go-pkg/v2/auth_config/internal/service"
	"github.com/c2pc/go-pkg/v2/auth_config/internal/transport/api/handler"
	"github.com/c2pc/go-pkg/v2/auth_config/transformer"
	"github.com/c2pc/go-pkg/v2/utils/mw"
)

type IAuthConfigHandler interface {
	InitHandler(secured *gin.RouterGroup, unsecured *gin.RouterGroup, handlers ...gin.HandlerFunc)
	GetService() service.IAuthConfigService
}

type AuthConfigHandler struct {
	db      *gorm.DB
	handler *handler.AuthConfigHandler
}

func NewAuthConfig(ctx context.Context, db *gorm.DB, transformers transformer.AuthConfigTransformers, tr mw.ITransaction) (IAuthConfigHandler, error) {
	if transformers == nil {
		return nil, errors.New("transformers is empty")
	}

	repositories := repository.NewRepositories(db)

	authConfigService := service.NewAuthConfigService(repositories.AuthConfigRepository, transformers)

	authConfigHandler := handler.NewAuthConfigHandlers(authConfigService, tr, transformers)

	exporter := &AuthConfigHandler{
		handler: authConfigHandler,
		db:      db,
	}

	return exporter, nil
}

func (e *AuthConfigHandler) InitHandler(secured *gin.RouterGroup, unsecured *gin.RouterGroup, handlers ...gin.HandlerFunc) {
	e.handler.Init(secured, unsecured)
}

func (e *AuthConfigHandler) GetService() service.IAuthConfigService {
	return e.handler.GetService()
}
