package handler

import (
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

type Handler struct {
	authService          service.IAuthService
	permissionService    service.IPermissionService
	roleService          service.IRoleService
	userService          service.IUserService
	tr                   mw.ITransaction
	tokenMiddleware      middleware.ITokenMiddleware
	permissionMiddleware middleware.IPermissionMiddleware
}

func NewHandlers(
	authService service.IAuthService,
	permissionService service.IPermissionService,
	roleService service.IRoleService,
	userService service.IUserService,
	tr mw.ITransaction,
	tokenMiddleware middleware.ITokenMiddleware,
	permissionMiddleware middleware.IPermissionMiddleware,
) *Handler {

	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		translator.SetValidateTranslators(v)

		_ = v.RegisterValidation("device_id", validator2.ValidateDeviceID, true)

		_ = v.RegisterTranslation(translator.RegisterValidatorTranslation(translator.RU, "device_id", "{0} неизвестное устройство", true))
		_ = v.RegisterTranslation(translator.RegisterValidatorTranslation(translator.EN, "device_id", "{0] unknown device", true))
	}

	return &Handler{
		authService,
		permissionService,
		roleService,
		userService,
		tr,
		tokenMiddleware,
		permissionMiddleware,
	}
}

func (h *Handler) Init(api *gin.RouterGroup) {
	authHandler := NewAuthHandlers(h.authService, h.tr, h.tokenMiddleware)
	permissionHandler := NewPermissionHandlers(h.permissionService)
	roleHandler := NewRoleHandlers(h.roleService)
	userHandler := NewUserHandlers(h.userService)

	handler := api.Group("/auth")
	{
		authHandler.Init(handler)
		//Authenticate
		auth := handler.Group("", h.tokenMiddleware.Authenticate)
		{
			permissionHandler.Init(auth)
			//Can
			perm := auth.Group("", h.permissionMiddleware.Can)
			{
				roleHandler.Init(perm)
				userHandler.Init(perm)
			}
		}
	}
}
