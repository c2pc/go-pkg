package handler

import (
	"github.com/c2pc/go-pkg/v2/auth"
	"github.com/c2pc/go-pkg/v2/example/internal/service"
	"github.com/c2pc/go-pkg/v2/utils/mw"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	authService auth.IAuth
	services    service.Services
	trx         mw.ITransaction
}

func NewHandlers(authService auth.IAuth,
	services service.Services,
	trx mw.ITransaction,
) *Handler {
	return &Handler{
		authService: authService,
		services:    services,
		trx:         trx,
	}
}

func (h *Handler) Init() *gin.Engine {
	handler := h.authService.NewHandlerEngine()

	h.initAPI(handler)

	return handler
}

func (h *Handler) initAPI(handler *gin.Engine) {
	api := handler.Group("api/v1")
	{
		h.authService.InitHandler(handler, api)

		secure := api.Group("", h.authService.AuthenticateMW, h.authService.CanPermissionMW)
		{
			NewNewsHandlers(h.services.News, h.trx, h.authService.Task()).Init(secure)
		}
	}
}
