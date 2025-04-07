package handler

import (
	"fmt"
	"html/template"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/c2pc/go-pkg/v2/auth/internal/service"
	"github.com/c2pc/go-pkg/v2/auth/internal/transport/api/middleware"
	"github.com/c2pc/go-pkg/v2/auth/internal/transport/api/request"
	"github.com/c2pc/go-pkg/v2/auth/internal/transport/api/templates"
	"github.com/c2pc/go-pkg/v2/auth/internal/transport/api/transformer"
	"github.com/c2pc/go-pkg/v2/auth/profile"
	"github.com/c2pc/go-pkg/v2/utils/apperr"
	"github.com/c2pc/go-pkg/v2/utils/logger"
	"github.com/c2pc/go-pkg/v2/utils/mcontext"
	"github.com/c2pc/go-pkg/v2/utils/mw"
	request2 "github.com/c2pc/go-pkg/v2/utils/request"
	response "github.com/c2pc/go-pkg/v2/utils/response/http"
	"github.com/c2pc/go-pkg/v2/utils/sso"
	"github.com/c2pc/go-pkg/v2/utils/sso/oidc"
	"github.com/c2pc/go-pkg/v2/utils/sso/saml"
	"github.com/gin-gonic/gin"
)

type AuthHandler[Model profile.IModel, CreateInput, UpdateInput, UpdateProfileInput any] struct {
	authService        service.IAuthService[Model, CreateInput, UpdateInput, UpdateProfileInput]
	tr                 mw.ITransaction
	tokenMiddleware    middleware.ITokenMiddleware
	profileTransformer profile.ITransformer[Model]
	profileRequest     profile.IRequest[CreateInput, UpdateInput, UpdateProfileInput]
	oidcAuth           oidc.AuthService
	samlAuth           saml.AuthService
}

func NewAuthHandlers[Model profile.IModel, CreateInput, UpdateInput, UpdateProfileInput any](
	authService service.IAuthService[Model, CreateInput, UpdateInput, UpdateProfileInput],
	tr mw.ITransaction,
	tokenMiddleware middleware.ITokenMiddleware,
	profileTransformer profile.ITransformer[Model],
	profileRequest profile.IRequest[CreateInput, UpdateInput, UpdateProfileInput],
	oidcAuth oidc.AuthService,
	samlAuth saml.AuthService,
) *AuthHandler[Model, CreateInput, UpdateInput, UpdateProfileInput] {
	return &AuthHandler[Model, CreateInput, UpdateInput, UpdateProfileInput]{
		authService,
		tr,
		tokenMiddleware,
		profileTransformer,
		profileRequest,
		oidcAuth,
		samlAuth,
	}
}

func (h *AuthHandler[Model, CreateInput, UpdateInput, UpdateProfileInput]) Init(engine *gin.Engine, api *gin.RouterGroup) {
	auth := api.Group("")
	{
		auth.POST("/login", h.tr.DBTransaction, h.login)
		auth.POST("/refresh", h.tr.DBTransaction, h.refresh)
		auth.POST("/logout", h.tr.DBTransaction, h.logout)
		auth.GET("/account", h.tokenMiddleware.Authenticate, h.account)

		if h.oidcAuth.IsEnabled() {
			auth.GET("/sso/login", h.oidcLogin)
			auth.GET("/sso/callback", h.tr.DBTransaction, h.oidcVerify)
		} else if h.samlAuth.IsEnabled() {
			auth.GET("/sso/login", h.samlAuth.SamlSP().RequireAccount, h.tr.DBTransaction, h.samlLogin)
		} else {
			auth.GET("/sso/login", func(c *gin.Context) {
				executeTemplate(c, "sso_not_supported.html")
			})
			auth.GET("/sso/callback", func(c *gin.Context) {
				executeTemplate(c, "sso_not_supported.html")
			})
		}
	}

	if h.samlAuth.IsEnabled() {
		engine.Any("/saml/:key", h.samlAuth.SamlSP().ServeHTTP)
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
		IsDomain: cred.DomainAuth,
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

func (h *AuthHandler[Model, CreateInput, UpdateInput, UpdateProfileInput]) oidcLogin(c *gin.Context) {
	cred, err := request2.BindQuery[request.AuthSSOLoginRequest](c)
	if err != nil {
		logger.WarningfLog(c.Request.Context(), "AUTH", fmt.Sprintf("sso.BindQuery error: %v", err))
		executeTemplate(c, "bad_request.html")
		return
	}

	origin := c.GetHeader("Referer")

	if strings.Index(cred.RedirectURL, origin) != 0 {
		logger.WarningfLog(c.Request.Context(), "AUTH", fmt.Sprintf("RedirectURL and origin is not the same"))
		executeTemplate(c, "bad_redirect_url2.html")
		return
	}

	if h.oidcAuth.IsEnabled() {
		if ok := h.oidcAuth.CheckRedirectURLs(cred.RedirectURL); !ok {
			logger.WarningfLog(c.Request.Context(), "AUTH", fmt.Sprintf("oidcAuth.CheckRedirectURLs redirect url not valid"))
			executeTemplate(c, "bad_redirect_url.html")
			return
		}

		state, code, err := h.oidcAuth.SumState(cred.RedirectURL, cred.DeviceID)
		if err != nil {
			logger.WarningfLog(c.Request.Context(), "AUTH", fmt.Sprintf("oidcAuth.SumState error: %v", err))
			executeTemplate(c, "bad_request.html")
			return
		}

		setCallbackCookie(c.Writer, c.Request, "state", state)

		http.Redirect(c.Writer, c.Request, code, http.StatusFound)
	} else {
		executeTemplate(c, "sso_not_supported.html")
		return
	}
}

func (h *AuthHandler[Model, CreateInput, UpdateInput, UpdateProfileInput]) oidcVerify(c *gin.Context) {
	ctx := mcontext.WithOperationIDContext(c.Request.Context(), strconv.Itoa(int(time.Now().UTC().Unix())))

	if h.oidcAuth.IsEnabled() {
		state, err := c.Request.Cookie("state")
		if err != nil {
			logger.WarningfLog(c.Request.Context(), "AUTH", fmt.Sprintf("cookie state error: %v", err))
			executeTemplate(c, "sso_invalid_state.html")
			return
		}

		if c.Request.URL.Query().Get("state") != state.Value {
			logger.WarningfLog(c.Request.Context(), "AUTH", fmt.Sprintf("cookie state error: %v", err))
			executeTemplate(c, "sso_invalid_state.html")
			return
		}

		token, err := h.oidcAuth.Verify(ctx, c.Request.URL.Query().Get("state"), c.Request.URL.Query().Get("code"))
		if err != nil {
			logger.WarningfLog(c.Request.Context(), "AUTH", fmt.Sprintf("oidcAuth.Verify error: %v", err))
			executeTemplate(c, "sso_invalid_state.html", http.StatusUnauthorized)
			return
		}

		accessExpire := math.Ceil(token.IDToken.Expiry.UTC().Sub(time.Now().UTC()).Minutes())

		authToken, userID, err := h.authService.Trx(request2.TxHandle(c)).SSO(ctx, service.SSO{
			Provider:     sso.OIDC,
			RefreshToken: token.IDToken.RefreshToken,
			AccessExpire: time.Duration(accessExpire) * time.Minute,
			Login:        *token.Login,
			DeviceID:     token.State.DeviceID,
		})
		if userID != 0 {
			c.Request = c.Request.WithContext(mcontext.WithOpUserIDContext(ctx, userID))
		}
		if err != nil {
			logger.WarningfLog(c.Request.Context(), "AUTH", fmt.Sprintf("authService.SSO error: %v", err))
			if apperr.Is(err, service.ErrAuthNoAccess) {
				executeTemplate(c, "sso_no_access.html", http.StatusForbidden)
			} else if apperr.Is(err, service.ErrSSONotSupported) {
				executeTemplate(c, "sso_not_supported.html")
			} else {
				executeTemplate(c, "sso_unauthenticated.html", http.StatusUnauthorized)
			}
			return
		}

		http.Redirect(c.Writer, c.Request,
			fmt.Sprintf("%s?accessToken=%s&refreshToken=%s&expires=%d",
				token.State.RedirectURL, authToken.Auth.Token, authToken.Auth.RefreshToken, int(authToken.Auth.ExpiresAt)),
			http.StatusFound)
	} else {
		executeTemplate(c, "sso_not_supported.html")
		return
	}
}

func (h *AuthHandler[Model, CreateInput, UpdateInput, UpdateProfileInput]) samlLogin(c *gin.Context) {
	ctx := mcontext.WithOperationIDContext(c.Request.Context(), strconv.Itoa(int(time.Now().UTC().Unix())))

	cred, err := request2.BindQuery[request.AuthSSOLoginRequest](c)
	if err != nil {
		logger.WarningfLog(c.Request.Context(), "AUTH", fmt.Sprintf("sso.BindQuery error: %v", err))
		executeTemplate(c, "bad_request.html")
		return
	}

	origin := c.GetHeader("Origin")

	if !strings.Contains(cred.RedirectURL, origin) {
		logger.WarningfLog(c.Request.Context(), "AUTH", fmt.Sprintf("RedirectURL and origin is not the same"))
		executeTemplate(c, "bad_redirect_url.html")
		return
	}

	if h.samlAuth.IsEnabled() {
		if ok := h.samlAuth.CheckRedirectURLs(cred.RedirectURL); !ok {
			logger.WarningfLog(c.Request.Context(), "AUTH", fmt.Sprintf("samlAuth.CheckRedirectURLs redirect url not valid"))
			executeTemplate(c, "bad_redirect_url.html")
			return
		}

		login := h.samlAuth.GetLoginFromContext(c.Request.Context())
		if login == "" {
			executeTemplate(c, "sso_invalid_state.html", http.StatusUnauthorized)
			return
		}

		authToken, userID, err := h.authService.Trx(request2.TxHandle(c)).SSO(ctx, service.SSO{
			Provider: sso.SAML,
			Login:    login,
			DeviceID: cred.DeviceID,
		})
		if userID != 0 {
			c.Request = c.Request.WithContext(mcontext.WithOpUserIDContext(ctx, userID))
		}
		if err != nil {
			logger.WarningfLog(c.Request.Context(), "AUTH", fmt.Sprintf("authService.SSO error: %v", err))
			if apperr.Is(err, service.ErrAuthNoAccess) {
				executeTemplate(c, "sso_no_access.html", http.StatusForbidden)
			} else if apperr.Is(err, service.ErrSSONotSupported) {
				executeTemplate(c, "sso_not_supported.html")
			} else {
				executeTemplate(c, "sso_unauthenticated.html", http.StatusUnauthorized)
			}
			return
		}

		http.Redirect(c.Writer, c.Request,
			fmt.Sprintf("%s?accessToken=%s&refreshToken=%s&expires=%d",
				cred.RedirectURL, authToken.Auth.Token, authToken.Auth.RefreshToken, int(authToken.Auth.ExpiresAt)),
			http.StatusFound)
	} else {
		executeTemplate(c, "sso_not_supported.html")
		return
	}
}

func executeTemplate(c *gin.Context, name string, code ...int) {
	tmpl, err := template.New(name).ParseFS(templates.Templates, name)
	if err != nil {
		return
	}
	if len(code) > 0 {
		c.Status(code[0])
	} else {
		c.Status(http.StatusBadRequest)
	}
	c.Header("Content-Type", "text/html; charset=utf-8")
	_ = tmpl.Execute(c.Writer, nil)
}

func setCallbackCookie(w http.ResponseWriter, r *http.Request, name, value string) {
	c := &http.Cookie{
		Name:     name,
		Value:    value,
		MaxAge:   int(time.Hour.Seconds()),
		Secure:   r.TLS != nil,
		HttpOnly: true,
	}
	http.SetCookie(w, c)
}
