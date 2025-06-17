package oidc

import (
	"context"
	"crypto/rand"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/c2pc/go-pkg/v2/utils/apperr"
	"github.com/c2pc/go-pkg/v2/utils/apperr/code"
	"github.com/c2pc/go-pkg/v2/utils/translator"
	"github.com/coreos/go-oidc"
	"golang.org/x/oauth2"
)

var (
	ErrServerIsNotUnavailable = apperr.New("oidc_server_is_not_unavailable", apperr.WithTextTranslate(translator.Translate{translator.RU: "Сервер недоступен", translator.EN: "Server is unavailable"}), apperr.WithCode(code.Unavailable))
	ErrInvalidState           = apperr.New("oidc_invalid_state", apperr.WithTextTranslate(translator.Translate{translator.RU: "Неизвестный источник", translator.EN: "Unknown source"}), apperr.WithCode(code.InvalidArgument))
	ErrInvalidCode            = apperr.New("oidc_invalid_code", apperr.WithTextTranslate(translator.Translate{translator.RU: "Неизвестный код", translator.EN: "Unknown code"}), apperr.WithCode(code.InvalidArgument))
	ErrInvalidToken           = apperr.New("oidc_invalid_token", apperr.WithTextTranslate(translator.Translate{translator.RU: "Неверный токен", translator.EN: "Invalid token"}), apperr.WithCode(code.InvalidArgument))
	ErrInvalidAttribute       = apperr.New("oidc_invalid_attribute", apperr.WithTextTranslate(translator.Translate{translator.RU: "Неверный аттрибут", translator.EN: "Invalid attribute"}), apperr.WithCode(code.InvalidArgument))
)

type AuthService interface {
	IsEnabled() bool
	SumState(redirectURL string, deviceID int) (string, string, error)
	SumStateAny(state any) (string, string, error)
	Verify(ctx context.Context, state string, code string) (*Token, error)
	VerifyAny(ctx context.Context, state string, code string, s any) (*Token, error)
	Refresh(ctx context.Context, token string) (*Token, error)
	CheckRedirectURLs(redirectURL string) bool
}

type Config struct {
	Enabled           bool
	ConfigURL         string
	ClientID          string
	ClientSecret      string
	RootURL           string
	LoginAttr         string
	ValidRedirectURLs []string
}

type Auth struct {
	enabled           bool
	loginAttr         string
	oauth2            *oauth2.Config
	verifier          *oidc.IDTokenVerifier
	validRedirectURLs []string
}

func NewAuthService(ctx context.Context, cfg Config) (*Auth, error) {
	auth := new(Auth)
	auth.enabled = cfg.Enabled
	auth.validRedirectURLs = cfg.ValidRedirectURLs

	if cfg.LoginAttr == "" {
		cfg.LoginAttr = "sub"
	}

	auth.loginAttr = cfg.LoginAttr

	if auth.enabled {
		if cfg.ConfigURL == "" {
			return nil, apperr.New("OIDC config url is required")
		}
		if cfg.ClientID == "" {
			return nil, apperr.New("OIDC client id is required")
		}
		if cfg.ClientSecret == "" {
			return nil, apperr.New("OIDC client secret is required")
		}

		if cfg.RootURL == "" {
			return nil, apperr.New("OIDC redirect url is required")
		}

		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		http.DefaultClient = &http.Client{Transport: tr}

		provider, err := oidc.NewProvider(ctx, cfg.ConfigURL)
		if err != nil {
			return nil, ErrServerIsNotUnavailable.WithError(err)
		}

		auth.oauth2 = &oauth2.Config{
			ClientID:     cfg.ClientID,
			ClientSecret: cfg.ClientSecret,
			RedirectURL:  cfg.RootURL,
			Endpoint:     provider.Endpoint(),
			Scopes:       []string{oidc.ScopeOpenID},
		}

		auth.verifier = provider.Verifier(&oidc.Config{
			ClientID: cfg.ClientID,
		})
	}

	return auth, nil
}

func (auth *Auth) IsEnabled() bool {
	return auth.enabled
}

func (auth *Auth) CheckRedirectURLs(redirectURL string) bool {
	if len(auth.validRedirectURLs) == 0 {
		return true
	}

	for _, u := range auth.validRedirectURLs {
		if strings.Index(redirectURL, u) == 0 {
			return true
		}
	}

	return false
}

type StateCode struct {
	Rand        string `json:"rand"`
	RedirectURL string `json:"redirect_url"`
	DeviceID    int    `json:"device_id"`
}

func (auth *Auth) SumState(redirectURL string, deviceID int) (string, string, error) {
	b, err := randString(16)
	if err != nil {
		return "", "", err
	}

	return auth.SumStateAny(StateCode{
		Rand:        string(b),
		RedirectURL: redirectURL,
		DeviceID:    deviceID,
	})
}

func (auth *Auth) SumStateAny(state any) (string, string, error) {
	n, err := json.Marshal(state)
	if err != nil {
		return "", "", err
	}

	s := base64.RawURLEncoding.EncodeToString(n)

	return s, auth.oauth2.AuthCodeURL(s), nil
}

type Token struct {
	IDToken *oauth2.Token
	Login   *string
	State   *StateCode
}

func (auth *Auth) Verify(ctx context.Context, state string, code string) (*Token, error) {
	var s StateCode
	token, err := auth.VerifyAny(ctx, state, code, &s)
	if err != nil {
		return nil, err
	}
	token.State = &s
	return token, nil
}

func (auth *Auth) VerifyAny(ctx context.Context, state string, code string, s any) (*Token, error) {
	//Декодирование state
	d, err := base64.RawURLEncoding.DecodeString(state)
	if err != nil {
		return nil, err
	}

	//Получение отправленных раннее данных в state
	err = json.Unmarshal(d, s)
	if err != nil {
		return nil, ErrInvalidState.WithError(err)
	}

	//Конвертирование кода в токен
	oauth2Token, err := auth.oauth2.Exchange(ctx, code)
	if err != nil {
		return nil, ErrInvalidCode.WithError(err)
	}

	//Проверка типа данных токена
	rawIDToken, ok := oauth2Token.Extra("id_token").(string)
	if !ok {
		return nil, ErrInvalidToken
	}

	//Парсинг и проверка токена
	idToken, err := auth.verifier.Verify(ctx, rawIDToken)
	if err != nil {
		return nil, ErrInvalidToken.WithError(err)
	}

	//Получение всех аттрибутов из токена
	var idTokenClaims map[string]interface{}
	if err = idToken.Claims(&idTokenClaims); err != nil {
		return nil, ErrInvalidToken.WithError(err)
	}

	//Проверка логина в аттрибутах(аттрибут логина указан в конфиге)
	login, ok := idTokenClaims[auth.loginAttr]
	if !ok {
		return nil, ErrInvalidAttribute.WithErrorText(fmt.Sprintf("%s is missing or does not exist", auth.loginAttr))
	}

	//Проверка типа данных логина
	loginStr, ok := login.(string)
	if !ok {
		return nil, ErrInvalidAttribute.WithErrorText(fmt.Sprintf("%s is not a string", auth.loginAttr))
	}

	return &Token{
		IDToken: oauth2Token,
		Login:   &loginStr,
	}, nil
}

func (auth *Auth) Refresh(ctx context.Context, token string) (*Token, error) {
	source := auth.oauth2.TokenSource(ctx, &oauth2.Token{
		AccessToken:  "",
		TokenType:    "",
		RefreshToken: token,
		Expiry:       time.Time{},
	})

	oauth2Token, err := source.Token()
	if err != nil {
		return nil, ErrInvalidToken.WithError(err)
	}

	return &Token{IDToken: oauth2Token}, nil
}

func randString(nByte int) ([]byte, error) {
	b := make([]byte, nByte)
	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		return nil, err
	}
	return b, nil
}
