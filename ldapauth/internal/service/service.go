package service

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/c2pc/go-pkg/apperr/utils/appErrors"
	"github.com/c2pc/go-pkg/ldapauth/internal/config"
	"github.com/c2pc/go-pkg/level"
	logger2 "github.com/c2pc/go-pkg/logger"
	"github.com/golang-jwt/jwt"
)

type LdapService struct {
	serverURL  string
	secretKey  []byte
	httpClient *http.Client
	debug      string
	serverId   int
}

func createCustomHTTPClient(timeout time.Duration) *http.Client {
	customTransport := http.DefaultTransport.(*http.Transport).Clone()

	customTransport.TLSClientConfig = &tls.Config{
		CipherSuites: []uint16{
			tls.TLS_RSA_WITH_RC4_128_SHA,
			tls.TLS_RSA_WITH_3DES_EDE_CBC_SHA,
			tls.TLS_RSA_WITH_AES_128_CBC_SHA,
			tls.TLS_RSA_WITH_AES_256_CBC_SHA,
			tls.TLS_RSA_WITH_AES_128_CBC_SHA256,
			tls.TLS_RSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_ECDSA_WITH_RC4_128_SHA,
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA,
			tls.TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA,
			tls.TLS_ECDHE_RSA_WITH_RC4_128_SHA,
			tls.TLS_ECDHE_RSA_WITH_3DES_EDE_CBC_SHA,
			tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA,
			tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA256,
			tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA256,
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256,
			tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305_SHA256,
			tls.TLS_AES_128_GCM_SHA256,
			tls.TLS_AES_256_GCM_SHA384,
			tls.TLS_CHACHA20_POLY1305_SHA256,
		},
		InsecureSkipVerify: true,
	}

	client := &http.Client{
		Timeout:   timeout,
		Transport: customTransport,
	}

	return client
}

func NewAuthService(cfg config.Config) *LdapService {
	client := createCustomHTTPClient(cfg.Timeout)

	return &LdapService{
		serverURL:  cfg.ServerURL,
		secretKey:  []byte(cfg.SecretKey),
		httpClient: client,
		serverId:   cfg.ServerID,
		debug:      cfg.Debug,
	}
}

type TokenResponse struct {
	Access   string `json:"access"`
	Refresh  string `json:"refresh"`
	UserData UserClaims
}

type UserClaims struct {
	UserID      int    `json:"user_id"`
	UserRoleID  int    `json:"user_role_id"`
	UserLogin   string `json:"user_login"`
	ServerId    int    `json:"server_id"`
	ServerAllow int    `json:"server_allow"`
	jwt.StandardClaims
}

func (a *LdapService) Login(username, password string) (*TokenResponse, error) {
	url := fmt.Sprintf("%s/api/token/", a.serverURL)

	reqData := map[string]interface{}{
		"username":  username,
		"password":  password,
		"is_domain": 1,
		"server_id": a.serverId,
	}

	reqBody, err := json.Marshal(reqData)
	if err != nil {
		return nil, appErrors.ErrInternal.WithError(err)
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, appErrors.ErrInternal.WithError(err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := a.httpClient.Do(req)
	if err != nil {
		if level.Is(a.debug, level.TEST) {
			logger2.WarningLog("[LDAP AUTH]", true, fmt.Sprintf("making request to LDAP server: %v", err))
		}
		return nil, appErrors.ErrInternal.WithError(err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			if level.Is(a.debug, level.TEST) {
				logger2.WarningLog("[LDAP AUTH]", true, fmt.Sprintf("error - %v", err))
			}
		}
	}()

	if resp.StatusCode == http.StatusForbidden {
		if level.Is(a.debug, level.TEST) {
			logger2.WarningLog("[LDAP AUTH]", true, fmt.Sprintf("Unauthorized access for user: %s", username))
		}
		return nil, appErrors.ErrForbidden
	} else if resp.StatusCode == http.StatusUnauthorized {
		if level.Is(a.debug, level.TEST) {
			logger2.WarningLog("[LDAP AUTH]", true, fmt.Sprintf("user unauthorized: %d", resp.StatusCode))
		}
		return nil, appErrors.ErrUnauthenticated
	} else if resp.StatusCode != http.StatusOK {
		if level.Is(a.debug, level.TEST) {
			logger2.WarningLog("[LDAP AUTH]", true, fmt.Sprintf("Unexpected status code: %d", resp.StatusCode))
		}
		return nil, appErrors.ErrInternal
	}

	var tokenResp TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, appErrors.ErrInternal.WithError(err)
	}

	userClaims, err := a.validateToken(tokenResp.Access)
	if err != nil {
		if level.Is(a.debug, level.TEST) {
			logger2.WarningLog("[LDAP AUTH]", true, fmt.Sprintf("Token validation failed: %v", err))
		}
		return nil, appErrors.ErrForbidden.WithError(err)
	}

	if userClaims == nil {
		if level.Is(a.debug, level.TEST) {
			logger2.WarningLog("[LDAP AUTH]", true, "user claims is nil")
		}
		return nil, appErrors.ErrForbidden
	}

	tokenResp.UserData = *userClaims

	return &tokenResp, nil
}

func (a *LdapService) Refresh(refreshToken string) (*TokenResponse, error) {
	url := fmt.Sprintf("%s/api/token/refresh/", a.serverURL)

	reqData := map[string]string{
		"token": refreshToken,
	}

	reqBody, err := json.Marshal(reqData)
	if err != nil {
		return nil, appErrors.ErrInternal.WithError(err)
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, appErrors.ErrInternal.WithError(err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := a.httpClient.Do(req)
	if err != nil {
		if level.Is(a.debug, level.TEST) {
			logger2.WarningLog("[LDAP AUTH]", true, fmt.Sprintf("Error making refresh request: %v", err))
		}
		return nil, appErrors.ErrInternal.WithError(err)
	}

	defer func() {
		if err := resp.Body.Close(); err != nil {
			if level.Is(a.debug, level.TEST) {
				logger2.WarningLog("[LDAP AUTH]", true, fmt.Sprintf("Error closing response body: %v", err))
			}
		}
	}()

	if resp.StatusCode == http.StatusNotFound {
		if level.Is(a.debug, level.TEST) {
			logger2.WarningLog("[LDAP AUTH]", true, fmt.Sprintf("Refresh token not found"))
		}
		return nil, appErrors.ErrNotFound
	} else if resp.StatusCode != http.StatusOK {
		if level.Is(a.debug, level.TEST) {
			logger2.WarningLog("[LDAP AUTH]", true, fmt.Sprintf("Unexpected status code during refresh: %d", resp.StatusCode))
		}
		return nil, appErrors.ErrInternal
	}

	var tokenResp TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, appErrors.ErrInternal.WithError(err)
	}

	userClaims, err := a.validateToken(tokenResp.Access)
	if err != nil {
		if level.Is(a.debug, level.TEST) {
			logger2.WarningLog("[LDAP AUTH]", true, fmt.Sprintf("Token validation failed: %v", err))
		}
		return nil, appErrors.ErrForbidden.WithError(err)
	}
	tokenResp.UserData = *userClaims

	return &tokenResp, nil
}

func (a *LdapService) validateToken(tokenString string) (*UserClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &UserClaims{}, func(token *jwt.Token) (interface{}, error) {

		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			if level.Is(a.debug, level.TEST) {
				logger2.ErrorLog("[LDAP AUTH]", true, fmt.Sprintf("Unexpected signing method: %v", token.Header["alg"]))
			}
			return nil, appErrors.ErrInternal
		}
		return a.secretKey, nil
	})

	if err != nil {
		return nil, appErrors.ErrInternal.WithError(err)
	}

	if claims, ok := token.Claims.(*UserClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, appErrors.ErrInternal
}
