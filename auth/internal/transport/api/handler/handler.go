package handler

import (
	"regexp"

	"github.com/c2pc/go-pkg/v2/auth/profile"

	service2 "github.com/c2pc/go-pkg/v2/auth/internal/service"
	middleware2 "github.com/c2pc/go-pkg/v2/auth/internal/transport/api/middleware"
	validator2 "github.com/c2pc/go-pkg/v2/auth/internal/validator"
	"github.com/c2pc/go-pkg/v2/utils/mw"
	"github.com/c2pc/go-pkg/v2/utils/translator"
	validator3 "github.com/c2pc/go-pkg/v2/utils/validator"
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
		translator.SetValidateTranslators(v)

		_ = v.RegisterValidation("device_id", validator2.ValidateDeviceID, true)
		_ = v.RegisterValidation("dot_underscore_hyphen", validator2.DotUnderscoreHyphen, false)
		_ = v.RegisterValidation("dot_underscore_hyphen_space", validator2.DotUnderscoreHyphenSpace, true)
		_ = v.RegisterValidation("spec_chars", validator3.ValidateRegex(regexp.MustCompile("^[a-zA-Z0-9а-яА-Я`~!@#$%^&*()_+={}\\[\\]\\\\|:;\"/'<>,.?-]*$")), false)

		_ = v.RegisterTranslation(translator.RegisterValidatorTranslation(translator.RU, "device_id", "{0} неизвестное устройство", true))
		_ = v.RegisterTranslation(translator.RegisterValidatorTranslation(translator.EN, "device_id", "{0] unknown device", true))
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
