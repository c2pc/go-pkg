package handler

import (
	"github.com/c2pc/go-pkg/v2/auth/service"
	"github.com/c2pc/go-pkg/v2/auth/transport/api/middleware"
	"github.com/c2pc/go-pkg/v2/auth/transport/api/request"
	"github.com/c2pc/go-pkg/v2/auth/transport/api/transformer"
	"github.com/c2pc/go-pkg/v2/utils/mw"
	request2 "github.com/c2pc/go-pkg/v2/utils/request"
	response "github.com/c2pc/go-pkg/v2/utils/response/http"
	"github.com/gin-gonic/gin"
	"net/http"
)

type AuthHandler struct {
	authService     service.IAuthService
	tr              mw.ITransaction
	tokenMiddleware middleware.ITokenMiddleware
}

func NewAuthHandlers(
	authService service.IAuthService,
	tr mw.ITransaction,
	tokenMiddleware middleware.ITokenMiddleware,
) *AuthHandler {
	return &AuthHandler{
		authService,
		tr,
		tokenMiddleware,
	}
}

func (h *AuthHandler) Init(api *gin.RouterGroup) {
	auth := api.Group("")
	{
		auth.POST("/login", h.tr.DBTransaction, h.login)
		auth.POST("/refresh", h.tr.DBTransaction, h.refresh)
		auth.POST("/logout", h.tr.DBTransaction, h.logout)
		auth.POST("/account", h.tokenMiddleware.Authenticate, h.account)
		auth.PATCH("/account", h.tokenMiddleware.Authenticate, h.tr.DBTransaction, h.updateAccountData)
	}
}

func (h *AuthHandler) login(c *gin.Context) {
	cred, err := request2.BindJSON[request.AuthLoginRequest](c)
	if err != nil {
		response.Response(c, err)
		return
	}

	data, err := h.authService.Trx(request2.TxHandle(c)).Login(c.Request.Context(), service.AuthLogin{
		Login:    cred.Login,
		Password: cred.Password,
		DeviceID: cred.DeviceID,
	})
	if err != nil {
		response.Response(c, err)
		return
	}

	c.JSON(http.StatusOK, transformer.AuthTokenTransform(data))
}

func (h *AuthHandler) refresh(c *gin.Context) {
	cred, err := request2.BindJSON[request.AuthRefreshRequest](c)
	if err != nil {
		response.Response(c, err)
		return
	}

	data, err := h.authService.Trx(request2.TxHandle(c)).Refresh(c.Request.Context(), service.AuthRefresh{
		Token:    cred.Token,
		DeviceID: cred.DeviceID,
	})
	if err != nil {
		response.Response(c, err)
		return
	}

	c.JSON(http.StatusOK, transformer.AuthTokenTransform(data))
}

func (h *AuthHandler) logout(c *gin.Context) {
	cred, err := request2.BindJSON[request.AuthLogoutRequest](c)
	if err != nil {
		response.Response(c, err)
		return
	}

	err = h.authService.Trx(request2.TxHandle(c)).Logout(c.Request.Context(), service.AuthLogout{
		Token: cred.Token,
	})
	if err != nil {
		response.Response(c, err)
		return
	}

	c.Status(http.StatusOK)
}

func (h *AuthHandler) account(c *gin.Context) {
	data, err := h.authService.Account(c.Request.Context())
	if err != nil {
		response.Response(c, err)
		return
	}

	c.JSON(http.StatusOK, transformer.AuthAccountTransform(data))
}

func (h *AuthHandler) updateAccountData(c *gin.Context) {
	cred, err := request2.BindJSON[request.AuthUpdateAccountDataRequest](c)
	if err != nil {
		response.Response(c, err)
		return
	}

	if err := h.authService.Trx(request2.TxHandle(c)).UpdateAccountData(c.Request.Context(), service.AuthUpdateAccountData{
		Login:      cred.Login,
		FirstName:  cred.FirstName,
		SecondName: cred.SecondName,
		LastName:   cred.LastName,
		Password:   cred.Password,
		Email:      cred.Email,
		Phone:      cred.Phone,
	}); err != nil {
		response.Response(c, err)
		return
	}

	c.Status(http.StatusOK)
}
