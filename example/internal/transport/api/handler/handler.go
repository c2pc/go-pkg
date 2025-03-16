package handler

import (
	"net/http"

	"github.com/c2pc/go-pkg/v2/analytics"
	"github.com/c2pc/go-pkg/v2/auth"
	"github.com/c2pc/go-pkg/v2/example/internal/service"
	"github.com/c2pc/go-pkg/v2/task"
	"github.com/c2pc/go-pkg/v2/utils/apperr"
	"github.com/c2pc/go-pkg/v2/utils/level"
	"github.com/c2pc/go-pkg/v2/utils/mw"
	response "github.com/c2pc/go-pkg/v2/utils/response/http"
	"github.com/c2pc/go-pkg/v2/websocket"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	authService     auth.IAuth
	taskService     task.Tasker
	analyticService analytics.Analytics
	services        service.Services
	trx             mw.ITransaction
	ws              websocket.WebSocket
}

func NewHandlers(authService auth.IAuth,
	services service.Services,
	trx mw.ITransaction,
	taskService task.Tasker,
	analyticService analytics.Analytics,
	ws websocket.WebSocket,
) *Handler {
	return &Handler{
		authService:     authService,
		services:        services,
		trx:             trx,
		taskService:     taskService,
		analyticService: analyticService,
		ws:              ws,
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
	api := handler.Group("api/v1", h.authService.LimiterMiddleware)
	{
		unsecured := api.Group("", h.analyticService.CollectAnalytic)
		{
			h.authService.InitHandler(unsecured)
		}

		secure := api.Group("", h.authService.Authenticate, h.authService.CanPermission)
		{

			h.analyticService.InitHandler(secure)
			h.ws.InitHandler(secure)
			withLimiter := secure.Group("", h.analyticService.CollectAnalytic)
			{
				h.taskService.InitHandler(withLimiter, unsecured)
				NewNewsHandlers(h.services.News, h.trx, h.taskService).Init(withLimiter)
			}

		}
	}
}
