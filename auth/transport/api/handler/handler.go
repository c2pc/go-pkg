package handler

import (
	"github.com/c2pc/go-pkg/v2/auth/profile"
	"github.com/c2pc/go-pkg/v2/auth/service"
	"github.com/c2pc/go-pkg/v2/auth/transport/api/middleware"
	validator2 "github.com/c2pc/go-pkg/v2/auth/validator"
	"github.com/c2pc/go-pkg/v2/utils/mw"
	"github.com/c2pc/go-pkg/v2/utils/translator"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

type IHandler interface {
	Init(api *gin.RouterGroup)
}

type Handler[Model, CreateInput, UpdateInput, UpdateProfileInput any] struct {
	authService          service.IAuthService[Model, CreateInput, UpdateInput, UpdateProfileInput]
	permissionService    service.IPermissionService
	roleService          service.IRoleService
	userService          service.IUserService[Model, CreateInput, UpdateInput, UpdateProfileInput]
	settingService       service.ISettingService
	tr                   mw.ITransaction
	tokenMiddleware      middleware.ITokenMiddleware
	permissionMiddleware middleware.IPermissionMiddleware
	profileTransformer   profile.ITransformer[Model]
	profileRequest       profile.IRequest[CreateInput, UpdateInput, UpdateProfileInput]
}

func NewHandlers[Model, CreateInput, UpdateInput, UpdateProfileInput any](
	authService service.IAuthService[Model, CreateInput, UpdateInput, UpdateProfileInput],
	permissionService service.IPermissionService,
	roleService service.IRoleService,
	userService service.IUserService[Model, CreateInput, UpdateInput, UpdateProfileInput],
	settingService service.ISettingService,
	tr mw.ITransaction,
	tokenMiddleware middleware.ITokenMiddleware,
	permissionMiddleware middleware.IPermissionMiddleware,
	profileTransformer profile.ITransformer[Model],
	profileRequest profile.IRequest[CreateInput, UpdateInput, UpdateProfileInput],
) *Handler[Model, CreateInput, UpdateInput, UpdateProfileInput] {

	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		translator.SetValidateTranslators(v)

		_ = v.RegisterValidation("device_id", validator2.ValidateDeviceID, true)

		_ = v.RegisterTranslation(translator.RegisterValidatorTranslation(translator.RU, "device_id", "{0} неизвестное устройство", true))
		_ = v.RegisterTranslation(translator.RegisterValidatorTranslation(translator.EN, "device_id", "{0] unknown device", true))
	}

	return &Handler[Model, CreateInput, UpdateInput, UpdateProfileInput]{
		authService,
		permissionService,
		roleService,
		userService,
		settingService,
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
			}
		}
	}
}
