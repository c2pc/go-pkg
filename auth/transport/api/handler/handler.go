package handler

import (
	"github.com/c2pc/go-pkg/v2/auth/service"
	"github.com/c2pc/go-pkg/v2/auth/transport/api/middleware"
	validator2 "github.com/c2pc/go-pkg/v2/auth/validator"
	"github.com/c2pc/go-pkg/v2/utils/transaction"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

type IHandler interface {
	Init(api *gin.RouterGroup)
}

type Handler struct {
	authService     service.IAuthService
	tr              transaction.ITransaction
	tokenMiddleware middleware.ITokenMiddleware
}

func NewHandlers(
	authService service.IAuthService,
	tr transaction.ITransaction,
	tokenMiddleware middleware.ITokenMiddleware,
) *Handler {
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		_ = v.RegisterValidation("device_id", validator2.ValidateDeviceID, true)
	}

	return &Handler{
		authService,
		tr,
		tokenMiddleware,
	}
}

func (h *Handler) Init(api *gin.RouterGroup) {
	authHandler := NewAuthHandlers(h.authService, h.tr, h.tokenMiddleware)

	handler := api.Group("/auth")
	{
		authHandler.Init(handler)
	}
}
