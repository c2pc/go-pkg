package auth

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"os"
	"time"

	"github.com/c2pc/go-pkg/v2/utils/apperr"
	"github.com/c2pc/go-pkg/v2/utils/apperr/code"
	"github.com/c2pc/go-pkg/v2/utils/cipher"
	"github.com/c2pc/go-pkg/v2/utils/sso/oidc"
	"github.com/c2pc/go-pkg/v2/utils/sso/saml"
	"github.com/c2pc/go-pkg/v2/utils/translator"
	"github.com/gin-gonic/gin/binding"
)

var (
	ErrInvalidOIDCSecret = apperr.New("invalid_oidc_secret", apperr.WithTextTranslate(translator.Translate{translator.RU: "Не задан OIDC секрет", translator.EN: "OIDC secret not set"}))
)

type Request struct {
	Key             *string  `json:"key" binding:"omitempty,min=8,max=256"`
	AccessTokenTTL  *float64 `json:"access_token_ttl" binding:"omitempty,min=1"`
	RefreshTokenTTL *float64 `json:"refresh_token_ttl" binding:"omitempty,min=1"`
	LDAP            []struct {
		Domain string `json:"domain" binding:"required,min=1,max=256"`
		Addrs  []struct {
			Addr    string `json:"addr" binding:"required,min=1,max=256"`
			Secured bool   `json:"secured"`
		} `json:"addrs" binding:"required,min=1,max=10,dive"`
	} `json:"ldap" binding:"omitempty,unique=Domain,dive"`
	SSO *struct {
		Enabled string `json:"enabled" binding:"required,oneof=oidc saml none"`
		OIDC    *struct {
			ConfigURL         string   `json:"config_url" binding:"required"`
			ClientID          string   `json:"client_id" binding:"required"`
			ClientSecret      *string  `json:"client_secret" binding:"omitempty"`
			RootURL           string   `json:"root_url" binding:"required,min=1,max=1024"`
			LoginAttr         string   `json:"login_attr" binding:"required"`
			ValidRedirectURLs []string `json:"valid_redirect_urls" binding:"required,max=10,dive,min=1,max=256"`
		} `json:"oidc,omitempty" binding:"required_if=Enabled oidc"`
		SAML *struct {
			RootURL           string   `json:"root_url" binding:"required,min=1,max=1024"`
			LoginAttr         string   `json:"login_attr" binding:"required"`
			ValidRedirectURLs []string `json:"valid_redirect_urls" binding:"required,max=10,dive,min=1,max=256"`
		} `json:"saml,omitempty" binding:"required_if=Enabled saml"`
	} `json:"sso" binding:"omitempty"`
	Limiter *struct {
		MaxAttempts int     `json:"max_attempts" binding:"required,min=1"`
		TTL         float64 `json:"ttl" binding:"required,min=1"`
	} `json:"limiter" binding:"omitempty"`
}

type Transformer struct {
	AccessTokenTTL  float64 `json:"access_token_ttl"`
	RefreshTokenTTL float64 `json:"refresh_token_ttl"`
	LDAP            []struct {
		Domain string `json:"key"`
		Addrs  []struct {
			Addr    string `json:"addr"`
			Secured bool   `json:"secured"`
		} `json:"addrs"`
	} `json:"ldap"`
	SSO *struct {
		Enabled string `json:"enabled"`
		OIDC    *struct {
			ConfigURL         string   `json:"config_url"`
			ClientID          string   `json:"client_id"`
			RootURL           string   `json:"root_url"`
			LoginAttr         string   `json:"login_attr"`
			ValidRedirectURLs []string `json:"valid_redirect_urls"`
		} `json:"oidc,omitempty"`
		SAML *struct {
			MetadataUploaded  bool     `json:"metadata_uploaded"`
			CertUploaded      bool     `json:"cert_uploaded"`
			KeyUploaded       bool     `json:"key_uploaded"`
			RootURL           string   `json:"root_url"`
			LoginAttr         string   `json:"login_attr"`
			ValidRedirectURLs []string `json:"valid_redirect_urls"`
		} `json:"saml,omitempty"`
	} `json:"sso"`
	Limiter *struct {
		MaxAttempts int     `json:"max_attempts"`
		TTL         float64 `json:"ttl"`
	} `json:"limiter"`
}

type Cfg struct {
	Key             []byte     `json:"key,omitempty"`
	AccessTokenTTL  float64    `json:"access_token_ttl"`
	RefreshTokenTTL float64    `json:"refresh_token_ttl"`
	LDAP            []CfgLDAP  `json:"ldap"`
	SSO             CfgSSO     `json:"sso"`
	Limiter         CfgLimiter `json:"limiter"`
}

type CfgLimiter struct {
	MaxAttempts int     `json:"max_attempts"`
	TTL         float64 `json:"ttl"`
}

type CfgLDAP struct {
	Domain string        `json:"key"`
	Addrs  []CfgLDAPAddr `json:"addrs"`
}

type CfgLDAPAddr struct {
	Addr    string `json:"addr"`
	Secured bool   `json:"secured"`
}

type CfgSSO struct {
	Enabled string   `json:"enabled"`
	OIDC    *CfgOIDC `json:"oidc,omitempty"`
	SAML    *CfgSAML `json:"saml,omitempty"`
}

type CfgOIDC struct {
	ConfigURL         string   `json:"config_url"`
	ClientID          string   `json:"client_id"`
	ClientSecret      []byte   `json:"client_secret,omitempty"`
	RootURL           string   `json:"root_url"`
	LoginAttr         string   `json:"login_attr"`
	ValidRedirectURLs []string `json:"valid_redirect_urls"`
}
type CfgSAML struct {
	MetaDataFile      string   `json:"meta_data_file"`
	CertFile          string   `json:"cert_file"`
	KeyFile           string   `json:"key_file"`
	RootURL           string   `json:"root_url"`
	LoginAttr         string   `json:"login_attr"`
	ValidRedirectURLs []string `json:"valid_redirect_urls"`
}

type Configurator struct {
	ch chan Cfg
}

func NewConfigurator() *Configurator {
	return &Configurator{
		ch: make(chan Cfg, 1),
	}
}

func (c *Configurator) Watch() <-chan Cfg {
	return c.ch
}

func (c *Configurator) Action() string {
	return "Изменение настроек аутентификации"
}

func (c *Configurator) Unmarshal(data []byte) (any, error) {
	var cfg Cfg
	err := json.Unmarshal(data, &cfg)
	if err != nil {
		return Cfg{}, err
	}

	if cfg.Key != nil {
		decr, _ := cipher.AESCipher.Decrypt(cfg.Key)
		cfg.Key = decr
	}

	if cfg.SSO.OIDC != nil {
		if cfg.SSO.OIDC.ClientSecret != nil {
			decr, _ := cipher.AESCipher.Decrypt(cfg.SSO.OIDC.ClientSecret)
			cfg.SSO.OIDC.ClientSecret = decr
		}
	}

	return cfg, nil
}

func (c *Configurator) Check(newData, lastData []byte) ([]byte, error) {
	var newCfg Request
	if err := binding.JSON.BindBody(newData, &newCfg); err != nil {
		return nil, apperr.ErrValidation.WithError(err)
	}

	var lastCfg Cfg
	err := json.Unmarshal(lastData, &lastCfg)
	if err != nil {
		return nil, err
	}

	if newCfg.Key != nil {
		encr, _ := cipher.AESCipher.Encrypt([]byte(*newCfg.Key))
		lastCfg.Key = encr
	}

	if newCfg.AccessTokenTTL != nil {
		lastCfg.AccessTokenTTL = *newCfg.AccessTokenTTL
	}

	if newCfg.RefreshTokenTTL != nil {
		lastCfg.RefreshTokenTTL = *newCfg.RefreshTokenTTL
	}

	if newCfg.LDAP != nil {
		ld := make([]CfgLDAP, len(newCfg.LDAP))
		for i, ldap := range newCfg.LDAP {
			addrs := make([]CfgLDAPAddr, len(ldap.Addrs))
			for j, addr := range ldap.Addrs {
				addrs[j] = CfgLDAPAddr{
					Addr:    addr.Addr,
					Secured: addr.Secured,
				}
			}
			ld[i] = CfgLDAP{
				Domain: ldap.Domain,
				Addrs:  addrs,
			}
		}
		lastCfg.LDAP = ld
	}

	if newCfg.SSO != nil {
		lastCfg.SSO = CfgSSO{
			Enabled: newCfg.SSO.Enabled,
		}

		if newCfg.SSO.OIDC != nil {
			if lastCfg.SSO.OIDC == nil && newCfg.SSO.OIDC.ClientSecret == nil {
				return nil, ErrInvalidOIDCSecret
			}

			var clientSecret []byte
			if newCfg.SSO.OIDC.ClientSecret != nil {
				clientSecret = []byte(*newCfg.SSO.OIDC.ClientSecret)
			} else if lastCfg.SSO.OIDC != nil {
				clientSecret = lastCfg.SSO.OIDC.ClientSecret
			}

			if clientSecret != nil {
				encr, _ := cipher.AESCipher.Encrypt(clientSecret)
				clientSecret = encr
			}

			lastCfg.SSO.OIDC = &CfgOIDC{
				ConfigURL:         newCfg.SSO.OIDC.ConfigURL,
				ClientID:          newCfg.SSO.OIDC.ClientID,
				ClientSecret:      clientSecret,
				RootURL:           newCfg.SSO.OIDC.RootURL,
				LoginAttr:         newCfg.SSO.OIDC.LoginAttr,
				ValidRedirectURLs: newCfg.SSO.OIDC.ValidRedirectURLs,
			}
		}

		if newCfg.SSO.SAML != nil {
			lastCfg.SSO.SAML = &CfgSAML{
				MetaDataFile:      "saml.key",
				CertFile:          "saml.metadata",
				KeyFile:           "saml.cert",
				RootURL:           newCfg.SSO.SAML.RootURL,
				LoginAttr:         newCfg.SSO.SAML.LoginAttr,
				ValidRedirectURLs: newCfg.SSO.SAML.ValidRedirectURLs,
			}
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if lastCfg.SSO.Enabled == "saml" {
			_, err := os.Stat("saml.key")
			if err != nil {
				return nil, apperr.New("saml_key_not_found", apperr.WithTextTranslate(translator.Translate{translator.RU: "Файл saml.key не найден", translator.EN: "File saml.key not found"}), apperr.WithCode(code.InvalidArgument))
			}
			_, err = os.Stat("saml.metadata")
			if err != nil {
				return nil, apperr.New("saml_metadata_not_found", apperr.WithTextTranslate(translator.Translate{translator.RU: "Файл saml.metadata не найден", translator.EN: "File saml.metadata not found"}), apperr.WithCode(code.InvalidArgument))
			}
			_, err = os.Stat("saml.cert")
			if err != nil {
				return nil, apperr.New("saml_cert_not_found", apperr.WithTextTranslate(translator.Translate{translator.RU: "Файл saml.cert не найден", translator.EN: "File saml.cert not found"}), apperr.WithCode(code.InvalidArgument))
			}

			_, err = saml.NewAuthService(ctx, saml.Config{
				Enabled:           true,
				MetaDataURL:       lastCfg.SSO.SAML.MetaDataFile,
				CertFile:          lastCfg.SSO.SAML.CertFile,
				KeyFile:           lastCfg.SSO.SAML.KeyFile,
				RootURL:           lastCfg.SSO.SAML.RootURL,
				LoginAttr:         lastCfg.SSO.SAML.LoginAttr,
				ValidRedirectURLs: lastCfg.SSO.SAML.ValidRedirectURLs,
			})
			if err != nil {
				return nil, apperr.New("saml_invalid", apperr.WithText("SAML: "+err.Error()), apperr.WithCode(code.InvalidArgument))
			}

		} else if lastCfg.SSO.Enabled == "oidc" {
			_, err = oidc.NewAuthService(ctx, oidc.Config{
				Enabled:           true,
				ConfigURL:         lastCfg.SSO.OIDC.ConfigURL,
				ClientID:          lastCfg.SSO.OIDC.ClientID,
				ClientSecret:      string(lastCfg.SSO.OIDC.ClientSecret),
				RootURL:           lastCfg.SSO.OIDC.RootURL,
				LoginAttr:         lastCfg.SSO.OIDC.LoginAttr,
				ValidRedirectURLs: lastCfg.SSO.OIDC.ValidRedirectURLs,
			})
			if err != nil {
				return nil, apperr.New("oidc_invalid", apperr.WithText("OIDC: "+err.Error()), apperr.WithCode(code.InvalidArgument))
			}
		}
	}

	if newCfg.Limiter != nil {
		lastCfg.Limiter = CfgLimiter{
			MaxAttempts: newCfg.Limiter.MaxAttempts,
			TTL:         newCfg.Limiter.TTL,
		}
	}

	return json.Marshal(lastCfg)
}

func (c *Configurator) AfterUpdate(data []byte) error {
	cfgUn, err := c.Unmarshal(data)
	if err != nil {
		return err
	}

	cfg := cfgUn.(Cfg)

	select {
	case c.ch <- cfg:
	default:
		select {
		case <-c.ch:
		default:
		}
		c.ch <- cfg
	}

	return nil
}

func (c *Configurator) Init() ([]byte, error) {
	t := rand.Text()
	encr, _ := cipher.AESCipher.Encrypt([]byte(t))

	cfg := Cfg{
		Key:             encr,
		AccessTokenTTL:  15,
		RefreshTokenTTL: 60,
		LDAP:            []CfgLDAP{},
		SSO: CfgSSO{
			Enabled: "none",
			OIDC:    &CfgOIDC{},
			SAML:    &CfgSAML{},
		},
		Limiter: CfgLimiter{
			MaxAttempts: 10,
			TTL:         60,
		},
	}

	return json.Marshal(cfg)
}

func (c *Configurator) Transform(data []byte) ([]byte, error) {
	var cfg Transformer
	err := json.Unmarshal(data, &cfg)
	if err != nil {
		return nil, err
	}

	if cfg.SSO.SAML != nil {
		_, err = os.Stat("saml.key")
		cfg.SSO.SAML.KeyUploaded = err == nil
		_, err = os.Stat("saml.metadata")
		cfg.SSO.SAML.MetadataUploaded = err == nil
		_, err = os.Stat("saml.cert")
		cfg.SSO.SAML.CertUploaded = err == nil

		if len(cfg.SSO.SAML.ValidRedirectURLs) == 0 {
			cfg.SSO.SAML.ValidRedirectURLs = make([]string, 0)
		}
	}
	if cfg.SSO.OIDC != nil {
		if len(cfg.SSO.OIDC.ValidRedirectURLs) == 0 {
			cfg.SSO.OIDC.ValidRedirectURLs = make([]string, 0)
		}
	}

	return json.Marshal(cfg)
}
