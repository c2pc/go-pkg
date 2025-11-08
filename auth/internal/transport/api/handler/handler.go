package handler

import (
	"github.com/c2pc/go-pkg/v2/auth/internal/fx"
	"github.com/c2pc/go-pkg/v2/auth/internal/service"
	"github.com/c2pc/go-pkg/v2/auth/internal/transport/api/middleware"
	customValidator "github.com/c2pc/go-pkg/v2/auth/internal/validator"
	"github.com/c2pc/go-pkg/v2/auth/profile"
	"github.com/c2pc/go-pkg/v2/task"
	"github.com/c2pc/go-pkg/v2/utils/mw"
	"github.com/c2pc/go-pkg/v2/websocket"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

type IHandler interface {
	Init(engine *gin.Engine, api *gin.RouterGroup, handlers ...gin.HandlerFunc)
}

type Handler struct {
	authService        service.IAuthService
	permissionService  service.IPermissionService
	roleService        service.IRoleService
	userService        service.IUserService
	settingService     service.ISettingService
	sessionService     service.ISessionService
	filterService      service.IFilterService
	configService      service.IConfigService
	analyticService    service.IAnalyticService
	userBlockedService service.IUserBlockedService
	versionService     service.VersionService

	tr                   mw.ITransaction
	tokenMiddleware      *middleware.TokenMiddleware
	permissionMiddleware middleware.IPermissionMiddleware
	analyticMiddleware   *middleware.AnalyticMiddleware
	profileTransformer   profile.ITransformer
	profileRequest       profile.IRequest
	oidcAuth             *fx.OIDCHolder
	samlAuth             *fx.SAMLHolder
	limiter              *fx.LimiterHolder

	ws     websocket.WebSocket
	tasker task.Tasker
}

func NewHandlers(
	authService service.IAuthService,
	permissionService service.IPermissionService,
	roleService service.IRoleService,
	userService service.IUserService,
	settingService service.ISettingService,
	filterService service.IFilterService,
	sessionService service.ISessionService,
	configService service.IConfigService,
	analyticService service.IAnalyticService,
	userBlockedService service.IUserBlockedService,

	tr mw.ITransaction,
	tokenMiddleware *middleware.TokenMiddleware,
	permissionMiddleware middleware.IPermissionMiddleware,
	analyticMiddleware *middleware.AnalyticMiddleware,
	profileTransformer profile.ITransformer,
	profileRequest profile.IRequest,
	oidcAuth *fx.OIDCHolder,
	samlAuth *fx.SAMLHolder,
	versionService service.VersionService,
	limiter *fx.LimiterHolder,
	ws websocket.WebSocket,
	tasker task.Tasker,
) *Handler {
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		customValidator.ExcludeSpaceValidation(v) //exclude_space
		customValidator.DeviceIDValidation(v)     //device_id
		customValidator.SpecCharsValidation(v)    //spec_chars
	}

	return &Handler{
		authService,
		permissionService,
		roleService,
		userService,
		settingService,
		sessionService,
		filterService,
		configService,
		analyticService,
		userBlockedService,
		versionService,
		tr,
		tokenMiddleware,
		permissionMiddleware,
		analyticMiddleware,
		profileTransformer,
		profileRequest,
		oidcAuth,
		samlAuth,
		limiter,
		ws,
		tasker,
	}
}

func (h *Handler) Init(engine *gin.Engine, api *gin.RouterGroup, handlers ...gin.HandlerFunc) {
	authHandler := NewAuthHandlers(h.authService, h.tr, h.tokenMiddleware, h.profileTransformer, h.profileRequest, h.oidcAuth, h.samlAuth, h.permissionMiddleware)
	userHandler := NewUserHandlers(h.userService, h.tr, h.profileTransformer, h.profileRequest)
	roleHandler := NewRoleHandlers(h.roleService, h.tr)
	permissionHandler := NewPermissionHandlers(h.permissionService)
	settingHandler := NewSettingHandlers(h.settingService, h.tr)
	filterHandler := NewFilterHandlers(h.filterService, h.tr)
	sessionHandler := NewSessionHandlers(h.sessionService, h.tr)
	versionHandler := NewVersionHandlers(h.versionService)
	configHandler := NewConfigHandlers(h.configService, h.tr)
	analyticHandler := NewAnalyticsHandler(h.analyticService)
	userBlockedHandler := NewUserBlockedHandlers(h.userBlockedService, h.tr, h.limiter)

	api.Any("ping", func(c *gin.Context) {
		version := h.versionService.Get(c.Request.Context())
		c.Data(200, "text/plain", []byte(version.AppName))
	})

	handler := api.Group("/auth")
	{
		authHandler.Init(engine, handler)
		secure := handler.Group("", h.tokenMiddleware.Authenticate)
		{
			settingHandler.Init(secure)
			filterHandler.Init(secure)
			versionHandler.Init(secure)
			permissionHandler.Init(secure)

			can := secure.Group("", h.permissionMiddleware.Can).Group("", handlers...)
			{
				roleHandler.Init(can)
				userHandler.Init(can)
				sessionHandler.Init(can)
				analyticHandler.Init(can)
				userBlockedHandler.Init(can)
			}
		}
	}

	secure := api.Group("", h.tokenMiddleware.Authenticate)
	{
		h.ws.InitHandler(secure)

		can := secure.Group("", h.permissionMiddleware.Can).Group("", handlers...)
		{
			configHandler.Init(can)
			h.tasker.InitHandler(can, api)
		}
	}
}
