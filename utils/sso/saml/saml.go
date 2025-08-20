package saml

import (
	"context"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/c2pc/go-pkg/v2/utils/apperr"
	"github.com/c2pc/go-pkg/v2/utils/apperr/code"
	"github.com/c2pc/go-pkg/v2/utils/translator"
	"github.com/crewjam/saml"
	"github.com/crewjam/saml/samlsp"
)

var (
	ErrServerIsNotUnavailable = apperr.New("saml_server_is_not_unavailable", apperr.WithTextTranslate(translator.Translate{translator.RU: "Сервер недоступен", translator.EN: "Server is unavailable"}), apperr.WithCode(code.Unavailable))
	ErrInvalidState           = apperr.New("saml_invalid_state", apperr.WithTextTranslate(translator.Translate{translator.RU: "Неизвестный источник", translator.EN: "Unknown source"}), apperr.WithCode(code.InvalidArgument))
	ErrInvalidCode            = apperr.New("saml_invalid_code", apperr.WithTextTranslate(translator.Translate{translator.RU: "Неизвестный код", translator.EN: "Unknown code"}), apperr.WithCode(code.InvalidArgument))
	ErrInvalidToken           = apperr.New("saml_invalid_token", apperr.WithTextTranslate(translator.Translate{translator.RU: "Неверный токен", translator.EN: "Invalid token"}), apperr.WithCode(code.InvalidArgument))
	ErrInvalidAttribute       = apperr.New("saml_invalid_attribute", apperr.WithTextTranslate(translator.Translate{translator.RU: "Неверный аттрибут", translator.EN: "Invalid attribute"}), apperr.WithCode(code.InvalidArgument))
)

type AuthService interface {
	IsEnabled() bool
	GetLoginFromContext(ctx context.Context) string
	CheckRedirectURLs(redirectURL string) bool
	SamlSP() *Middleware
}

type Config struct {
	Enabled           bool
	MetaDataURL       string
	MetaDataPath      string
	CertFile          string
	KeyFile           string
	RootURL           string
	LoginAttr         string
	ValidRedirectURLs []string
}

type Auth struct {
	enabled           bool
	loginAttr         string
	validRedirectURLs []string
	samlSP            *Middleware
}

func NewAuthService(ctx context.Context, cfg Config) (*Auth, error) {
	auth := new(Auth)
	auth.enabled = cfg.Enabled
	auth.validRedirectURLs = cfg.ValidRedirectURLs

	if auth.enabled {
		if cfg.LoginAttr == "" {
			return nil, apperr.New("SAML login attribute is required")
		}
		auth.loginAttr = cfg.LoginAttr

		if cfg.CertFile == "" {
			return nil, apperr.New("SAML cert file is required")
		}
		if cfg.KeyFile == "" {
			return nil, apperr.New("SAML key file is required")
		}

		if cfg.MetaDataURL == "" && cfg.MetaDataPath == "" {
			return nil, apperr.New("SAML metadata url or metadata path are required")
		}

		rootURL, err := url.Parse(cfg.RootURL)
		if err != nil {
			return nil, fmt.Errorf(`SAML: invalid root url "%s"`, cfg.RootURL)
		}

		keyPair, err := tls.LoadX509KeyPair(cfg.CertFile, cfg.KeyFile)
		if err != nil {
			return nil, fmt.Errorf("SAML: failed to load certificate key pair: %s", err)
		}

		keyPair.Leaf, err = x509.ParseCertificate(keyPair.Certificate[0])
		if err != nil {
			return nil, fmt.Errorf("SAML: failed to parse certificate: %s", err)
		}

		var idpMetadata *saml.EntityDescriptor
		if cfg.MetaDataPath != "" {
			dat, err := os.ReadFile(cfg.MetaDataPath)
			if err != nil {
				return nil, fmt.Errorf("SAML: failed to read metadata file: %s", err)
			}
			idpMetadata, err = samlsp.ParseMetadata(dat)
			if err != nil {
				return nil, fmt.Errorf("SAML: failed to parse metadata file: %s", err)
			}
		} else {
			tr := &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			}
			httpClient := &http.Client{Transport: tr}

			idpMetadataURL, err := url.Parse(cfg.MetaDataURL)
			if err != nil {
				return nil, fmt.Errorf("SAML: failed to parse metadata url: %s", err)
			}

			idpMetadata, err = samlsp.FetchMetadata(ctx, httpClient, *idpMetadataURL)
			if err != nil {
				return nil, fmt.Errorf("SAML: failed to fetch metadata: %s", err)
			}
		}

		samlSP, err := newSamlSP(samlsp.Options{
			URL:         *rootURL,
			Key:         keyPair.PrivateKey.(*rsa.PrivateKey),
			Certificate: keyPair.Leaf,
			IDPMetadata: idpMetadata,
		})
		if err != nil {
			return nil, fmt.Errorf("SAML: failed to create SAML SP: %s", err)
		}

		samlSP.ServiceProvider.AuthnNameIDFormat = saml.UnspecifiedNameIDFormat

		auth.samlSP = samlSP
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

func (auth *Auth) SamlSP() *Middleware {
	return auth.samlSP
}

func (auth *Auth) GetLoginFromContext(ctx context.Context) string {
	fmt.Printf("%+v\n", samlsp.SessionFromContext(ctx))
	return samlsp.AttributeFromContext(ctx, auth.loginAttr)
}
