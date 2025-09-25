package handler

import (
	"github.com/c2pc/go-pkg/v2/auth/fx"
	"github.com/c2pc/go-pkg/v2/auth/internal/service"
	"github.com/c2pc/go-pkg/v2/auth/internal/transport/api/middleware"
	customValidator "github.com/c2pc/go-pkg/v2/auth/internal/validator"
	"github.com/c2pc/go-pkg/v2/auth/profile"
	"github.com/c2pc/go-pkg/v2/utils/mw"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

type IHandler interface {
	Init(engine *gin.Engine, api *gin.RouterGroup, handlers ...gin.HandlerFunc)
}

type Handler struct {
	authService          service.IAuthService
	permissionService    service.IPermissionService
	roleService          service.IRoleService
	userService          service.IUserService
	settingService       service.ISettingService
	sessionService       service.ISessionService
	filterService        service.IFilterService
	tr                   mw.ITransaction
	tokenMiddleware      *middleware.TokenMiddleware
	permissionMiddleware middleware.IPermissionMiddleware
	profileTransformer   profile.ITransformer
	profileRequest       profile.IRequest
	oidcAuth             *fx.OIDCHolder
	samlAuth             *fx.SAMLHolder
	versionService       service.VersionService
}

func NewHandlers(
	authService service.IAuthService,
	permissionService service.IPermissionService,
	roleService service.IRoleService,
	userService service.IUserService,
	settingService service.ISettingService,
	filterService service.IFilterService,
	sessionService service.ISessionService,
	tr mw.ITransaction,
	tokenMiddleware *middleware.TokenMiddleware,
	permissionMiddleware middleware.IPermissionMiddleware,
	profileTransformer profile.ITransformer,
	profileRequest profile.IRequest,
	oidcAuth *fx.OIDCHolder,
	samlAuth *fx.SAMLHolder,
	versionService service.VersionService,
) *Handler {

	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		customValidator.DotUnderscoreHyphenValidation(v)      //dot_underscore_hyphen
		customValidator.DotUnderscoreHyphenSpaceValidation(v) //dot_underscore_hyphen_space
		customValidator.DeviceIDValidation(v)                 //device_id
		customValidator.SpecCharsValidation(v)                //spec_chars
		customValidator.PhoneNumberValidation(v)              //phone_number
	}

	return &Handler{
		authService,
		permissionService,
		roleService,
		userService,
		settingService,
		sessionService,
		filterService,
		tr,
		tokenMiddleware,
		permissionMiddleware,
		profileTransformer,
		profileRequest,
		oidcAuth,
		samlAuth,
		versionService,
	}
}

func (h *Handler) Init(engine *gin.Engine, api *gin.RouterGroup, handlers ...gin.HandlerFunc) {
	authHandler := NewAuthHandlers(h.authService, h.tr, h.tokenMiddleware, h.profileTransformer, h.profileRequest, h.oidcAuth, h.samlAuth, h.permissionMiddleware)
	permissionHandler := NewPermissionHandlers(h.permissionService)
	roleHandler := NewRoleHandlers(h.roleService, h.tr)
	userHandler := NewUserHandlers(h.userService, h.tr, h.profileTransformer, h.profileRequest)
	settingHandler := NewSettingHandlers(h.settingService, h.tr)
	filterHandler := NewFilterHandlers(h.filterService, h.tr)
	sessionHandler := NewSessionHandlers(h.sessionService, h.tr)
	versionHandler := NewVersionHandlers(h.versionService)

	handler := api.Group("/auth")
	{
		authHandler.Init(engine, handler)
		//Authenticate
		auth := handler.Group("", h.tokenMiddleware.Authenticate)
		{
			settingHandler.Init(auth)
			filterHandler.Init(auth)
			versionHandler.Init(auth)
			permissionHandler.Init(auth)

			//Can
			perm := auth.Group("", h.permissionMiddleware.Can).Group("", handlers...)
			{
				roleHandler.Init(perm)
				userHandler.Init(perm)
				sessionHandler.Init(perm)
			}
		}
	}
}
