package handler

import (
	"net/http"

	"github.com/c2pc/go-pkg/v2/auth"
	"github.com/c2pc/go-pkg/v2/example/internal/service"
	"github.com/c2pc/go-pkg/v2/task"
	"github.com/c2pc/go-pkg/v2/utils/apperr"
	"github.com/c2pc/go-pkg/v2/utils/level"
	"github.com/c2pc/go-pkg/v2/utils/mw"
	response "github.com/c2pc/go-pkg/v2/utils/response/http"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	authService auth.IAuth
	taskService task.Tasker
	services    service.Services
	trx         mw.ITransaction
}

func NewHandlers(authService auth.IAuth, services service.Services, trx mw.ITransaction, taskService task.Tasker) *Handler {
	return &Handler{
		authService: authService,
		services:    services,
		trx:         trx,
		taskService: taskService,
	}
}

func (h *Handler) Init(debug string) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	handler := gin.New()

	gin.DebugPrintRouteFunc = func(httpMethod, absolutePath, handlerName string, nuHandlers int) {}

	handler.Use(
		gin.Recovery(),
		mw.CorsHandler(),
		mw.GinParseOperationID(),
	)

	if level.Is(debug, level.DEVELOPMENT, level.TEST) {
		handler.Use(
			gin.LoggerWithConfig(mw.LogHandler("HTTP", debug)),
			mw.GinBodyLogMiddleware("HTTP", debug),
		)
	}

	// Init handler
	handler.NoRoute(func(c *gin.Context) {
		response.Response(c, apperr.ErrNotFound)
	})
	handler.POST("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})
	handler.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "")
	})

	h.initAPI(handler)

	return handler
}

func (h *Handler) initAPI(handler *gin.Engine) {
	api := handler.Group("api/v1")
	{
		h.authService.InitHandler(api)

		secure := api.Group("", h.authService.Authenticate, h.authService.CanPermission)
		{
			h.taskService.InitHandler(secure)
			NewNewsHandlers(h.services.NewsService, h.trx, h.taskService).Init(secure)
		}
	}
}