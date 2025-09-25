package handler

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"html/template"
	"net/http"
	"strings"
	"time"

	"github.com/c2pc/go-pkg/v2/auth/fx"
	"github.com/c2pc/go-pkg/v2/auth/internal/service"
	"github.com/c2pc/go-pkg/v2/auth/internal/transport/api/middleware"
	"github.com/c2pc/go-pkg/v2/auth/internal/transport/api/request"
	"github.com/c2pc/go-pkg/v2/auth/internal/transport/api/templates"
	"github.com/c2pc/go-pkg/v2/auth/internal/transport/api/transformer"
	"github.com/c2pc/go-pkg/v2/auth/profile"
	"github.com/c2pc/go-pkg/v2/utils/apperr"
	"github.com/c2pc/go-pkg/v2/utils/apperr/code"
	"github.com/c2pc/go-pkg/v2/utils/mcontext"
	"github.com/c2pc/go-pkg/v2/utils/mw"
	request2 "github.com/c2pc/go-pkg/v2/utils/request"
	response "github.com/c2pc/go-pkg/v2/utils/response/http"
	"github.com/c2pc/go-pkg/v2/utils/sso"
	"github.com/crewjam/saml/samlsp"
	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authService          service.IAuthService
	tr                   mw.ITransaction
	tokenMiddleware      *middleware.TokenMiddleware
	profileTransformer   profile.ITransformer
	profileRequest       profile.IRequest
	oidcAuth             *fx.OIDCHolder
	samlAuth             *fx.SAMLHolder
	permissionMiddleware middleware.IPermissionMiddleware
}

func NewAuthHandlers(
	authService service.IAuthService,
	tr mw.ITransaction,
	tokenMiddleware *middleware.TokenMiddleware,
	profileTransformer profile.ITransformer,
	profileRequest profile.IRequest,
	oidcAuth *fx.OIDCHolder,
	samlAuth *fx.SAMLHolder,
	permissionMiddleware middleware.IPermissionMiddleware,
) *AuthHandler {
	return &AuthHandler{
		authService,
		tr,
		tokenMiddleware,
		profileTransformer,
		profileRequest,
		oidcAuth,
		samlAuth,
		permissionMiddleware,
	}
}

func (h *AuthHandler) Init(engine *gin.Engine, api *gin.RouterGroup) {
	auth := api.Group("")
	{
		auth.POST("/login", h.tr.DBTransaction, h.login)
		auth.POST("/refresh", h.tr.DBTransaction, h.refresh)
		auth.POST("/logout", h.tr.DBTransaction, h.logout)
		auth.GET("/account", h.tokenMiddleware.Authenticate, h.account)
		auth.GET("/sso/login", h.tr.DBTransaction, h.ssoLogin)
		auth.GET("/sso/callback", h.tr.DBTransaction, h.ssoCallback)
		auth.POST("/configs/saml.metadata", h.tokenMiddleware.Authenticate, h.permissionMiddleware.Can, h.uploadSamlMetadata)
		auth.POST("/configs/saml.cert", h.tokenMiddleware.Authenticate, h.permissionMiddleware.Can, h.uploadSamlCert)
	}
	engine.Any("/saml/:key", func(c *gin.Context) {
		if h.samlAuth.Get().IsEnabled() {
			h.samlAuth.Get().SamlSP().ServeHTTP(c)
		}
	})
}

func (h *AuthHandler) ssoLogin(c *gin.Context) {
	if h.oidcAuth.Get().IsEnabled() {
		h.oidcLogin(c)
		return
	}

	if h.samlAuth.Get().IsEnabled() {
		handlers := []gin.HandlerFunc{
			h.samlAuth.Get().SamlSP().RequireAccount,
			h.samlLogin,
		}
		runHandlers(c, handlers)
		return
	}

	executeTemplate(c, "sso_not_supported.html")
}

func (h *AuthHandler) ssoCallback(c *gin.Context) {
	if h.oidcAuth.Get().IsEnabled() {
		handlers := []gin.HandlerFunc{
			h.oidcVerify,
		}
		runHandlers(c, handlers)
		return
	}

	c.Request = c.Request.WithContext(mcontext.WithOpActionContext(c.Request.Context(), "Неуспешный вход в систему"))
	executeTemplate(c, "sso_not_supported.html")
}

func runHandlers(c *gin.Context, handlers []gin.HandlerFunc) {
	for _, h := range handlers {
		if c.IsAborted() {
			return
		}
		h(c)
	}
}

func (h *AuthHandler) login(c *gin.Context) {
	cred, err := request2.BindJSON[request.AuthLoginRequest](c)
	if err != nil {
		c.Request = c.Request.WithContext(mcontext.WithOpActionContext(c.Request.Context(), "Неуспешный вход в систему"))
		response.Response(c, err)
		return
	}

	data, userID, err := h.authService.Trx(request2.TxHandle(c)).Login(c.Request.Context(), service.AuthLogin{
		Login:    cred.Login,
		Password: cred.Password,
		DeviceID: cred.DeviceID,
		Secret:   c.GetHeader("X-Broker"),
	})
	if userID != 0 {
		c.Request = c.Request.WithContext(mcontext.WithOpUserIDContext(c.Request.Context(), userID))
	}
	if err != nil {
		c.Request = c.Request.WithContext(mcontext.WithOpActionContext(c.Request.Context(), "Неуспешный вход в систему"))
		response.Response(c, err)
		return
	}
	c.Request = c.Request.WithContext(mcontext.WithOpActionContext(c.Request.Context(), "Вход пользователя в систему"))

	c.JSON(http.StatusOK, transformer.AuthTokenTransform(data, h.profileTransformer))
}

func (h *AuthHandler) refresh(c *gin.Context) {
	cred, err := request2.BindJSON[request.AuthRefreshRequest](c)
	if err != nil {
		c.Request = c.Request.WithContext(mcontext.WithOpActionContext(c.Request.Context(), "Неуспешное обновление токена"))
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
		c.Request = c.Request.WithContext(mcontext.WithOpActionContext(c.Request.Context(), "Неуспешное обновление токена"))
		response.Response(c, err)
		return
	}

	c.Request = c.Request.WithContext(mcontext.WithOpActionContext(c.Request.Context(), "Обновление токена"))
	c.Request = c.Request.WithContext(mcontext.WithOpUserIDContext(c.Request.Context(), data.Auth.UserID))

	c.JSON(http.StatusOK, transformer.AuthTokenTransform(data, h.profileTransformer))
}

func (h *AuthHandler) logout(c *gin.Context) {
	cred, err := request2.BindJSON[request.AuthLogoutRequest](c)
	if err != nil {
		c.Request = c.Request.WithContext(mcontext.WithOpActionContext(c.Request.Context(), "Неуспешный выход из системы"))
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
		c.Request = c.Request.WithContext(mcontext.WithOpActionContext(c.Request.Context(), "Неуспешный выход из системы"))
		response.Response(c, err)
		return
	}

	c.Request = c.Request.WithContext(mcontext.WithOpActionContext(c.Request.Context(), "Выход пользователя из системы"))
	c.Status(http.StatusOK)
}

func (h *AuthHandler) account(c *gin.Context) {
	data, err := h.authService.Account(c.Request.Context())
	if err != nil {
		response.Response(c, err)
		return
	}

	c.JSON(http.StatusOK, transformer.AuthAccountTransform(data, h.profileTransformer))
}

func (h *AuthHandler) oidcLogin(c *gin.Context) {
	cred, err := request2.BindQuery[request.AuthSSOLoginRequest](c)
	if err != nil {
		c.Request = c.Request.WithContext(mcontext.WithOpActionContext(c.Request.Context(), "Неуспешный вход в систему"))
		executeTemplate(c, "bad_request.html", http.StatusBadRequest)
		return
	}

	origin := c.GetHeader("Referer")

	if strings.Index(cred.RedirectURL, origin) != 0 {
		c.Request = c.Request.WithContext(mcontext.WithOpActionContext(c.Request.Context(), "Неуспешный вход в систему"))
		executeTemplate(c, "bad_redirect_url2.html")
		return
	}

	if h.oidcAuth.Get().IsEnabled() {
		if ok := h.oidcAuth.Get().CheckRedirectURLs(cred.RedirectURL); !ok {
			c.Request = c.Request.WithContext(mcontext.WithOpActionContext(c.Request.Context(), "Неуспешный вход в систему"))
			executeTemplate(c, "bad_redirect_url.html")
			return
		}

		state, code, err := h.oidcAuth.Get().SumState(cred.RedirectURL, cred.DeviceID)
		if err != nil {
			c.Request = c.Request.WithContext(mcontext.WithOpActionContext(c.Request.Context(), "Неуспешный вход в систему"))
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

func (h *AuthHandler) oidcVerify(c *gin.Context) {
	if h.oidcAuth.Get().IsEnabled() {
		state, err := c.Request.Cookie("state")
		if err != nil {
			c.Request = c.Request.WithContext(mcontext.WithOpActionContext(c.Request.Context(), "Неуспешный вход в систему"))
			executeTemplate(c, "sso_invalid_state.html")
			return
		}

		if c.Request.URL.Query().Get("state") != state.Value {
			c.Request = c.Request.WithContext(mcontext.WithOpActionContext(c.Request.Context(), "Неуспешный вход в систему"))
			executeTemplate(c, "sso_invalid_state.html")
			return
		}

		token, err := h.oidcAuth.Get().Verify(c.Request.Context(), c.Request.URL.Query().Get("state"), c.Request.URL.Query().Get("code"))
		if err != nil {
			c.Request = c.Request.WithContext(mcontext.WithOpActionContext(c.Request.Context(), "Неуспешный вход в систему"))
			executeTemplate(c, "sso_invalid_state.html", http.StatusUnauthorized)
			return
		}

		authToken, userID, err := h.authService.Trx(request2.TxHandle(c)).SSO(c.Request.Context(), service.SSO{
			Provider:     sso.OIDC,
			RefreshToken: token.IDToken.RefreshToken,
			Login:        *token.Login,
			DeviceID:     token.State.DeviceID,
		})
		if userID != 0 {
			c.Request = c.Request.WithContext(mcontext.WithOpUserIDContext(c.Request.Context(), userID))
		}
		if err != nil {
			c.Request = c.Request.WithContext(mcontext.WithOpActionContext(c.Request.Context(), "Неуспешный вход в систему"))
			if apperr.Is(err, service.ErrAuthNoAccess) {
				executeTemplate(c, "sso_no_access.html", http.StatusForbidden)
			} else if apperr.Is(err, service.ErrAuthBlocked) {
				executeTemplate(c, "sso_no_access.html", http.StatusForbidden)
			} else if apperr.Is(err, service.ErrSSONotSupported) {
				executeTemplate(c, "sso_not_supported.html")
			} else {
				executeTemplate(c, "sso_unauthenticated.html", http.StatusUnauthorized)
			}
			return
		}

		c.Request = c.Request.WithContext(mcontext.WithOpActionContext(c.Request.Context(), "Вход пользователя в систему"))
		http.Redirect(c.Writer, c.Request,
			fmt.Sprintf("%s?accessToken=%s&refreshToken=%s&expires=%d",
				token.State.RedirectURL, authToken.Auth.Token, authToken.Auth.RefreshToken, int(authToken.Auth.ExpiresAt)),
			http.StatusFound)
	} else {
		c.Request = c.Request.WithContext(mcontext.WithOpActionContext(c.Request.Context(), "Неуспешный вход в систему"))
		executeTemplate(c, "sso_not_supported.html")
		return
	}
}

func (h *AuthHandler) samlLogin(c *gin.Context) {
	cred, err := request2.BindQuery[request.AuthSSOLoginRequest](c)
	if err != nil {
		c.Request = c.Request.WithContext(mcontext.WithOpActionContext(c.Request.Context(), "Неуспешный вход в систему"))
		executeTemplate(c, "bad_request.html", http.StatusBadRequest)
		return
	}

	origin := c.GetHeader("Origin")

	if !strings.Contains(cred.RedirectURL, origin) {
		c.Request = c.Request.WithContext(mcontext.WithOpActionContext(c.Request.Context(), "Неуспешный вход в систему"))
		executeTemplate(c, "bad_redirect_url.html")
		return
	}

	if h.samlAuth.Get().IsEnabled() {
		if ok := h.samlAuth.Get().CheckRedirectURLs(cred.RedirectURL); !ok {
			c.Request = c.Request.WithContext(mcontext.WithOpActionContext(c.Request.Context(), "Неуспешный вход в систему"))
			executeTemplate(c, "bad_redirect_url.html")
			return
		}

		login := h.samlAuth.Get().GetLoginFromContext(c.Request.Context())
		if login == "" {
			c.Request = c.Request.WithContext(mcontext.WithOpActionContext(c.Request.Context(), "Неуспешный вход в систему"))
			executeTemplate(c, "sso_invalid_state.html", http.StatusUnauthorized)
			return
		}

		authToken, userID, err := h.authService.Trx(request2.TxHandle(c)).SSO(c.Request.Context(), service.SSO{
			Provider: sso.SAML,
			Login:    login,
			DeviceID: cred.DeviceID,
		})
		if userID != 0 {
			c.Request = c.Request.WithContext(mcontext.WithOpUserIDContext(c.Request.Context(), userID))
		}
		if err != nil {
			c.Request = c.Request.WithContext(mcontext.WithOpActionContext(c.Request.Context(), "Неуспешный вход в систему"))
			if apperr.Is(err, service.ErrAuthNoAccess) {
				executeTemplate(c, "sso_no_access.html", http.StatusForbidden)
			} else if apperr.Is(err, service.ErrAuthBlocked) {
				executeTemplate(c, "sso_no_access.html", http.StatusForbidden)
			} else if apperr.Is(err, service.ErrSSONotSupported) {
				executeTemplate(c, "sso_not_supported.html")
			} else {
				executeTemplate(c, "sso_unauthenticated.html", http.StatusUnauthorized)
			}
			return
		}

		c.Request = c.Request.WithContext(mcontext.WithOpActionContext(c.Request.Context(), "Вход пользователя в систему"))
		http.Redirect(c.Writer, c.Request,
			fmt.Sprintf("%s?accessToken=%s&refreshToken=%s&expires=%d",
				cred.RedirectURL, authToken.Auth.Token, authToken.Auth.RefreshToken, int(authToken.Auth.ExpiresAt)),
			http.StatusFound)
	} else {
		c.Request = c.Request.WithContext(mcontext.WithOpActionContext(c.Request.Context(), "Неуспешный вход в систему"))
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
		c.Status(http.StatusUnauthorized)
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

func (h *AuthHandler) uploadSamlMetadata(c *gin.Context) {
	file, err := c.FormFile("metadata")
	if err != nil {
		response.Response(c, apperr.ErrBadRequest.WithError(err))
		return
	}

	dat, err := file.Open()
	if err != nil {
		response.Response(c, apperr.ErrInternal.WithError(err))
		return
	}
	defer dat.Close()

	var buf bytes.Buffer
	if _, err := buf.ReadFrom(dat); err != nil {
		response.Response(c, apperr.ErrInternal.WithError(err))
		return
	}

	_, err = samlsp.ParseMetadata(buf.Bytes())
	if err != nil {
		response.Response(c, apperr.New("invalid_file", apperr.WithText(err.Error()), apperr.WithCode(code.InvalidArgument)))
		return
	}

	err = c.SaveUploadedFile(file, "saml.metadata")
	if err != nil {
		response.Response(c, apperr.ErrInternal.WithError(err))
		return
	}

	c.Status(http.StatusOK)
}

func (h *AuthHandler) uploadSamlCert(c *gin.Context) {
	cert, err := c.FormFile("cert")
	if err != nil {
		response.Response(c, apperr.ErrBadRequest.WithError(err))
		return
	}

	key, err := c.FormFile("key")
	if err != nil {
		response.Response(c, apperr.ErrBadRequest.WithError(err))
		return
	}

	certFile, err := cert.Open()
	if err != nil {
		response.Response(c, apperr.ErrInternal.WithError(err))
		return
	}
	defer certFile.Close()

	keyFile, err := key.Open()
	if err != nil {
		response.Response(c, apperr.ErrInternal.WithError(err))
		return
	}
	defer keyFile.Close()

	var certBuf, keyBuf bytes.Buffer
	if _, err := certBuf.ReadFrom(certFile); err != nil {
		response.Response(c, apperr.ErrInternal.WithError(err))
		return
	}

	if _, err := keyBuf.ReadFrom(keyFile); err != nil {
		response.Response(c, apperr.ErrInternal.WithError(err))
		return
	}

	keyPair, err := tls.X509KeyPair(certBuf.Bytes(), keyBuf.Bytes())
	if err != nil {
		response.Response(c, apperr.New("invalid_cert", apperr.WithText(err.Error()), apperr.WithCode(code.InvalidArgument)))
		return
	}

	keyPair.Leaf, err = x509.ParseCertificate(keyPair.Certificate[0])
	if err != nil {
		response.Response(c, apperr.New("invalid_key", apperr.WithText(err.Error()), apperr.WithCode(code.InvalidArgument)))
		return
	}

	err = c.SaveUploadedFile(cert, "saml.cert")
	if err != nil {
		response.Response(c, apperr.ErrInternal.WithError(err))
		return
	}

	err = c.SaveUploadedFile(key, "saml.key")
	if err != nil {
		response.Response(c, apperr.ErrInternal.WithError(err))
		return
	}

	c.Status(http.StatusOK)
}
