package handler

import (
	"github.com/c2pc/go-pkg/v2/auth/profile"
	"net/http"

	"github.com/c2pc/go-pkg/v2/auth/internal/service"
	"github.com/c2pc/go-pkg/v2/auth/internal/transport/api/middleware"
	"github.com/c2pc/go-pkg/v2/auth/internal/transport/api/request"
	"github.com/c2pc/go-pkg/v2/auth/internal/transport/api/transformer"
	"github.com/c2pc/go-pkg/v2/utils/mcontext"
	"github.com/c2pc/go-pkg/v2/utils/mw"
	request2 "github.com/c2pc/go-pkg/v2/utils/request"
	response "github.com/c2pc/go-pkg/v2/utils/response/http"
	"github.com/gin-gonic/gin"
)

type AuthHandler[Model, CreateInput, UpdateInput, UpdateProfileInput any] struct {
	authService        service.IAuthService[Model, CreateInput, UpdateInput, UpdateProfileInput]
	tr                 mw.ITransaction
	tokenMiddleware    middleware.ITokenMiddleware
	profileTransformer profile.ITransformer[Model]
	profileRequest     profile.IRequest[CreateInput, UpdateInput, UpdateProfileInput]
}

func NewAuthHandlers[Model, CreateInput, UpdateInput, UpdateProfileInput any](
	authService service.IAuthService[Model, CreateInput, UpdateInput, UpdateProfileInput],
	tr mw.ITransaction,
	tokenMiddleware middleware.ITokenMiddleware,
	profileTransformer profile.ITransformer[Model],
	profileRequest profile.IRequest[CreateInput, UpdateInput, UpdateProfileInput],
) *AuthHandler[Model, CreateInput, UpdateInput, UpdateProfileInput] {
	return &AuthHandler[Model, CreateInput, UpdateInput, UpdateProfileInput]{
		authService,
		tr,
		tokenMiddleware,
		profileTransformer,
		profileRequest,
	}
}

func (h *AuthHandler[Model, CreateInput, UpdateInput, UpdateProfileInput]) Init(api *gin.RouterGroup) {
	auth := api.Group("")
	{
		auth.POST("/login", h.tr.DBTransaction, h.login)
		auth.POST("/refresh", h.tr.DBTransaction, h.refresh)
		auth.POST("/logout", h.tr.DBTransaction, h.logout)
		auth.POST("/account", h.tokenMiddleware.Authenticate, h.account)
		auth.PATCH("/account", h.tokenMiddleware.Authenticate, h.tr.DBTransaction, h.updateAccountData)
	}
}

func (h *AuthHandler[Model, CreateInput, UpdateInput, UpdateProfileInput]) login(c *gin.Context) {
	cred, err := request2.BindJSON[request.AuthLoginRequest](c)
	if err != nil {
		response.Response(c, err)
		return
	}

	data, userID, err := h.authService.Trx(request2.TxHandle(c)).Login(c.Request.Context(), service.AuthLogin{
		Login:    cred.Login,
		Password: cred.Password,
		DeviceID: cred.DeviceID,
	})
	if userID != 0 {
		c.Request = c.Request.WithContext(mcontext.WithOpUserIDContext(c.Request.Context(), userID))
	}
	if err != nil {
		response.Response(c, err)
		return
	}

	c.JSON(http.StatusOK, transformer.AuthTokenTransform(data, h.profileTransformer))
}

func (h *AuthHandler[Model, CreateInput, UpdateInput, UpdateProfileInput]) refresh(c *gin.Context) {
	cred, err := request2.BindJSON[request.AuthRefreshRequest](c)
	if err != nil {
		response.Response(c, err)
		return
	}

	data, userID, err := h.authService.Trx(request2.TxHandle(c)).Refresh(c.Request.Context(), service.AuthRefresh{
		Token:    cred.Token,
		DeviceID: cred.DeviceID,
	})
	if userID != 0 {
		c.Request = c.Request.WithContext(mcontext.WithOpUserIDContext(c.Request.Context(), userID))
	}
	if err != nil {
		response.Response(c, err)
		return
	}

	c.Request = c.Request.WithContext(mcontext.WithOpUserIDContext(c.Request.Context(), data.Auth.UserID))

	c.JSON(http.StatusOK, transformer.AuthTokenTransform(data, h.profileTransformer))
}

func (h *AuthHandler[Model, CreateInput, UpdateInput, UpdateProfileInput]) logout(c *gin.Context) {
	cred, err := request2.BindJSON[request.AuthLogoutRequest](c)
	if err != nil {
		response.Response(c, err)
		return
	}

	userID, err := h.authService.Trx(request2.TxHandle(c)).Logout(c.Request.Context(), service.AuthLogout{
		Token: cred.Token,
	})
	if userID != 0 {
		c.Request = c.Request.WithContext(mcontext.WithOpUserIDContext(c.Request.Context(), userID))
	}
	if err != nil {
		response.Response(c, err)
		return
	}

	c.Status(http.StatusOK)
}

func (h *AuthHandler[Model, CreateInput, UpdateInput, UpdateProfileInput]) account(c *gin.Context) {
	data, err := h.authService.Account(c.Request.Context())
	if err != nil {
		response.Response(c, err)
		return
	}

	c.JSON(http.StatusOK, transformer.AuthAccountTransform(data, h.profileTransformer))
}

func (h *AuthHandler[Model, CreateInput, UpdateInput, UpdateProfileInput]) updateAccountData(c *gin.Context) {
	cred, err := request2.BindJSON[request.AuthUpdateAccountDataRequest](c)
	if err != nil {
		response.Response(c, err)
		return
	}

	var profileCred *UpdateProfileInput
	if h.profileRequest != nil {
		profileCred, err = h.profileRequest.UpdateProfileRequest(c)
		if err != nil {
			response.Response(c, err)
			return
		}
	}

	if err := h.authService.Trx(request2.TxHandle(c)).UpdateAccountData(c.Request.Context(), service.AuthUpdateAccountData{
		Login:      cred.Login,
		FirstName:  cred.FirstName,
		SecondName: cred.SecondName,
		LastName:   cred.LastName,
		Password:   cred.Password,
		Email:      cred.Email,
		Phone:      cred.Phone,
	}, profileCred); err != nil {
		response.Response(c, err)
		return
	}

	c.Status(http.StatusOK)
}
