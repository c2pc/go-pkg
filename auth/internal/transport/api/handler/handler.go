package handler

import (
	"github.com/c2pc/go-pkg/v2/auth/profile"
	"github.com/c2pc/go-pkg/v2/utils/sso/oidc"
	"github.com/c2pc/go-pkg/v2/utils/sso/saml"

	"github.com/c2pc/go-pkg/v2/auth/internal/service"
	"github.com/c2pc/go-pkg/v2/auth/internal/transport/api/middleware"
	customValidator "github.com/c2pc/go-pkg/v2/auth/internal/validator"
	"github.com/c2pc/go-pkg/v2/utils/mw"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

type IHandler interface {
	Init(engine *gin.Engine, api *gin.RouterGroup)
}

type Handler[Model profile.IModel, CreateInput, UpdateInput, UpdateProfileInput any] struct {
	authService          service.IAuthService[Model, CreateInput, UpdateInput, UpdateProfileInput]
	permissionService    service.IPermissionService
	roleService          service.IRoleService
	userService          service.IUserService[Model, CreateInput, UpdateInput, UpdateProfileInput]
	settingService       service.ISettingService
	sessionService       service.ISessionService
	filterService        service.IFilterService
	tr                   mw.ITransaction
	tokenMiddleware      middleware.ITokenMiddleware
	permissionMiddleware middleware.IPermissionMiddleware
	profileTransformer   profile.ITransformer[Model]
	profileRequest       profile.IRequest[CreateInput, UpdateInput, UpdateProfileInput]
	oidcAuth             oidc.AuthService
	samlAuth             saml.AuthService
	versionService       service.VersionService
}

func NewHandlers[Model profile.IModel, CreateInput, UpdateInput, UpdateProfileInput any](
	authService service.IAuthService[Model, CreateInput, UpdateInput, UpdateProfileInput],
	permissionService service.IPermissionService,
	roleService service.IRoleService,
	userService service.IUserService[Model, CreateInput, UpdateInput, UpdateProfileInput],
	settingService service.ISettingService,
	filterService service.IFilterService,
	sessionService service.ISessionService,
	tr mw.ITransaction,
	tokenMiddleware middleware.ITokenMiddleware,
	permissionMiddleware middleware.IPermissionMiddleware,
	profileTransformer profile.ITransformer[Model],
	profileRequest profile.IRequest[CreateInput, UpdateInput, UpdateProfileInput],
	oidcAuth oidc.AuthService,
	samlAuth saml.AuthService,
	versionService service.VersionService,
) *Handler[Model, CreateInput, UpdateInput, UpdateProfileInput] {

	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		customValidator.DotUnderscoreHyphenValidation(v)      //dot_underscore_hyphen
		customValidator.DotUnderscoreHyphenSpaceValidation(v) //dot_underscore_hyphen_space
		customValidator.DeviceIDValidation(v)                 //device_id
		customValidator.SpecCharsValidation(v)                //spec_chars
		customValidator.PhoneNumberValidation(v)              //phone_number
	}

	return &Handler[Model, CreateInput, UpdateInput, UpdateProfileInput]{
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

func (h *Handler[Model, CreateInput, UpdateInput, UpdateProfileInput]) Init(engine *gin.Engine, api *gin.RouterGroup) {
	authHandler := NewAuthHandlers(h.authService, h.tr, h.tokenMiddleware, h.profileTransformer, h.profileRequest, h.oidcAuth, h.samlAuth)
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

			//Can
			perm := auth.Group("", h.permissionMiddleware.Can)
			{
				permissionHandler.Init(perm)
				roleHandler.Init(perm)
				userHandler.Init(perm)
				sessionHandler.Init(perm)
			}
		}
	}
}
