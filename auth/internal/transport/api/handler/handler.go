package handler

import (
	"github.com/c2pc/go-pkg/v2/auth/profile"

	service2 "github.com/c2pc/go-pkg/v2/auth/internal/service"
	middleware2 "github.com/c2pc/go-pkg/v2/auth/internal/transport/api/middleware"
	customValidator "github.com/c2pc/go-pkg/v2/auth/internal/validator"
	"github.com/c2pc/go-pkg/v2/utils/mw"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

type IHandler interface {
	Init(api *gin.RouterGroup)
}

type Handler[Model profile.IModel, CreateInput, UpdateInput, UpdateProfileInput any] struct {
	authService          service2.IAuthService[Model, CreateInput, UpdateInput, UpdateProfileInput]
	permissionService    service2.IPermissionService
	roleService          service2.IRoleService
	userService          service2.IUserService[Model, CreateInput, UpdateInput, UpdateProfileInput]
	settingService       service2.ISettingService
	sessionService       service2.ISessionService
	tr                   mw.ITransaction
	tokenMiddleware      middleware2.ITokenMiddleware
	permissionMiddleware middleware2.IPermissionMiddleware
	profileTransformer   profile.ITransformer[Model]
	profileRequest       profile.IRequest[CreateInput, UpdateInput, UpdateProfileInput]
}

func NewHandlers[Model profile.IModel, CreateInput, UpdateInput, UpdateProfileInput any](
	authService service2.IAuthService[Model, CreateInput, UpdateInput, UpdateProfileInput],
	permissionService service2.IPermissionService,
	roleService service2.IRoleService,
	userService service2.IUserService[Model, CreateInput, UpdateInput, UpdateProfileInput],
	settingService service2.ISettingService,
	sessionService service2.ISessionService,
	tr mw.ITransaction,
	tokenMiddleware middleware2.ITokenMiddleware,
	permissionMiddleware middleware2.IPermissionMiddleware,
	profileTransformer profile.ITransformer[Model],
	profileRequest profile.IRequest[CreateInput, UpdateInput, UpdateProfileInput],
) *Handler[Model, CreateInput, UpdateInput, UpdateProfileInput] {

	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		customValidator.DotUnderscoreHyphenValidation(v)      //dot_underscore_hyphen
		customValidator.DotUnderscoreHyphenSpaceValidation(v) //dot_underscore_hyphen_space
		customValidator.DeviceIDValidation(v)                 //device_id
		customValidator.SpecCharsValidation(v)                //spec_chars
	}

	return &Handler[Model, CreateInput, UpdateInput, UpdateProfileInput]{
		authService,
		permissionService,
		roleService,
		userService,
		settingService,
		sessionService,
		tr,
		tokenMiddleware,
		permissionMiddleware,
		profileTransformer,
		profileRequest,
	}
}

func (h *Handler[Model, CreateInput, UpdateInput, UpdateProfileInput]) Init(api *gin.RouterGroup) {
	authHandler := NewAuthHandlers(h.authService, h.tr, h.tokenMiddleware, h.profileTransformer, h.profileRequest)
	permissionHandler := NewPermissionHandlers(h.permissionService)
	roleHandler := NewRoleHandlers(h.roleService, h.tr)
	userHandler := NewUserHandlers(h.userService, h.tr, h.profileTransformer, h.profileRequest)
	settingHandler := NewSettingHandlers(h.settingService, h.tr)
	sessionHandler := NewSessionHandlers(h.sessionService, h.tr)

	handler := api.Group("/auth")
	{
		authHandler.Init(handler)
		//Authenticate
		auth := handler.Group("", h.tokenMiddleware.Authenticate)
		{
			settingHandler.Init(auth)
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
