package configurator

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"time"

	"github.com/c2pc/go-pkg/v2/auth/configurator"
	"github.com/c2pc/go-pkg/v2/auth/internal/model"
	"github.com/c2pc/go-pkg/v2/utils/apperr"
	"github.com/c2pc/go-pkg/v2/utils/apperr/code"
	"github.com/c2pc/go-pkg/v2/utils/cipher"
	"github.com/c2pc/go-pkg/v2/utils/datautil"
	"github.com/c2pc/go-pkg/v2/utils/sso/oidc"
	"github.com/c2pc/go-pkg/v2/utils/sso/saml"
	"github.com/c2pc/go-pkg/v2/utils/translator"
	"github.com/gin-gonic/gin/binding"
)

var (
	ErrInvalidOIDCSecret = apperr.New("invalid_oidc_secret", apperr.WithTextTranslate(translator.Translate{translator.RU: "Не задан OIDC секрет", translator.EN: "OIDC secret not set"}))
)

type ARequest struct {
	AccessKey       *string  `json:"access_key" binding:"omitempty,min=8,max=255"`
	AccessTokenTTL  *float64 `json:"access_token_ttl" binding:"omitempty,min=1,max=43200"`
	RefreshTokenTTL *float64 `json:"refresh_token_ttl" binding:"omitempty,min=1,gtefield=AccessTokenTTL,max=43200"`
	LDAP            []struct {
		Domain  string   `json:"domain" binding:"required,min=1,max=255"`
		Secured bool     `json:"secured"`
		Md5hash bool     `json:"md5hash"`
		Addrs   []string `json:"addrs" binding:"required,min=1,max=255,dive,min=1,max=255"`
	} `json:"ldap" binding:"omitempty,unique=Domain,dive"`
	SSO *struct {
		Enabled     string  `json:"enabled" binding:"required,oneof=oidc saml none"`
		Description *string `json:"description" binding:"omitempty,min=1,max=255"`
		OIDC        *struct {
			ConfigURL         string   `json:"config_url" binding:"required"`
			ClientID          string   `json:"client_id" binding:"required"`
			ClientSecret      *string  `json:"client_secret" binding:"omitempty,min=1"`
			RootURL           string   `json:"root_url" binding:"required,min=1,max=255"`
			LoginAttr         string   `json:"login_attr" binding:"required,min=1,max=255"`
			ValidRedirectURLs []string `json:"valid_redirect_urls" binding:"required,max=255,dive,min=1,max=255"`
		} `json:"oidc,omitempty" binding:"required_if=Enabled oidc"`
		SAML *struct {
			RootURL           string   `json:"root_url" binding:"required,min=1,max=255"`
			LoginAttr         string   `json:"login_attr" binding:"required,min=1,max=255"`
			ValidRedirectURLs []string `json:"valid_redirect_urls" binding:"required,max=255,dive,min=1,max=255"`
		} `json:"saml,omitempty" binding:"required_if=Enabled saml"`
	} `json:"sso" binding:"omitempty"`
	Limiter *struct {
		MaxAttempts int     `json:"max_attempts" binding:"required,min=1,max=999"`
		TTL         float64 `json:"ttl" binding:"required,min=1,max=999"`
		BlockingTTL float64 `json:"blocking_ttl" binding:"required,min=1,max=999"`
	} `json:"limiter" binding:"omitempty"`
}

type ATransformer struct {
	IsAccessKey     bool    `json:"is_access_key"`
	AccessTokenTTL  float64 `json:"access_token_ttl"`
	RefreshTokenTTL float64 `json:"refresh_token_ttl"`
	LDAP            []struct {
		Domain  string   `json:"domain"`
		Secured bool     `json:"secured"`
		Md5hash bool     `json:"md5hash"`
		Addrs   []string `json:"addrs"`
	} `json:"ldap"`
	SSO struct {
		Enabled     string `json:"enabled"`
		Description string `json:"description"`
		OIDC        *struct {
			ConfigURL         string   `json:"config_url"`
			ClientID          string   `json:"client_id"`
			IsClientSecret    bool     `json:"is_client_secret"`
			RootURL           string   `json:"root_url"`
			LoginAttr         string   `json:"login_attr"`
			ValidRedirectURLs []string `json:"valid_redirect_urls"`
		} `json:"oidc,omitempty"`
		SAML *struct {
			MetadataFile      string   `json:"metadata_file"`
			CertFile          string   `json:"cert_file"`
			KeyFile           string   `json:"key_file"`
			RootURL           string   `json:"root_url"`
			LoginAttr         string   `json:"login_attr"`
			ValidRedirectURLs []string `json:"valid_redirect_urls"`
		} `json:"saml,omitempty"`
	} `json:"sso"`
	Limiter struct {
		MaxAttempts int     `json:"max_attempts"`
		TTL         float64 `json:"ttl"`
		BlockingTTL float64 `json:"blocking_ttl"`
	} `json:"limiter"`
}

type ACfg struct {
	AccessKey       []byte      `json:"access_key,omitempty"`
	AccessTokenTTL  float64     `json:"access_token_ttl"`
	RefreshTokenTTL float64     `json:"refresh_token_ttl"`
	LDAP            []ACfgLDAP  `json:"ldap"`
	SSO             ACfgSSO     `json:"sso"`
	Limiter         ACfgLimiter `json:"limiter"`
}

type ACfgLimiter struct {
	MaxAttempts int     `json:"max_attempts"`
	TTL         float64 `json:"ttl"`
	BlockingTTL float64 `json:"blocking_ttl"`
}

type ACfgLDAP struct {
	Domain  string   `json:"domain"`
	Secured bool     `json:"secured"`
	Md5hash bool     `json:"md5hash"`
	Addrs   []string `json:"addrs"`
}

type ACfgSSO struct {
	Enabled     string    `json:"enabled"`
	Description string    `json:"description"`
	OIDC        *ACfgOIDC `json:"oidc,omitempty"`
	SAML        *ACfgSAML `json:"saml,omitempty"`
}

type ACfgOIDC struct {
	ConfigURL         string   `json:"config_url"`
	ClientID          string   `json:"client_id"`
	ClientSecret      []byte   `json:"client_secret,omitempty"`
	RootURL           string   `json:"root_url"`
	LoginAttr         string   `json:"login_attr"`
	ValidRedirectURLs []string `json:"valid_redirect_urls"`
}
type ACfgSAML struct {
	MetaDataFile      []byte   `json:"-"`
	CertFile          []byte   `json:"-"`
	KeyFile           []byte   `json:"-"`
	RootURL           string   `json:"root_url"`
	LoginAttr         string   `json:"login_attr"`
	ValidRedirectURLs []string `json:"valid_redirect_urls"`
}

type AConfigurator struct {
	ch chan ACfg
}

func NewAuthConfigurator() *AConfigurator {
	return &AConfigurator{
		ch: make(chan ACfg, 1),
	}
}

func (c *AConfigurator) Watch() <-chan ACfg {
	return c.ch
}

func (c *AConfigurator) Action() string {
	return "Изменение настроек аутентификации"
}

func randomBase64(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func filesMapBytes(files []configurator.File) map[string][]byte {
	m := make(map[string][]byte, len(files))
	for _, f := range files {
		m[f.GetKey()] = f.GetData()
	}
	return m
}

func filesMapNames(files []configurator.File) map[string]string {
	m := make(map[string]string, len(files))
	for _, f := range files {
		m[f.GetKey()] = f.GetName()
	}
	return m
}

func (c *AConfigurator) Unmarshal(data configurator.Config) (any, error) {
	var cfg ACfg
	if err := json.Unmarshal(data.GetValue(), &cfg); err != nil {
		return ACfg{}, err
	}

	if cfg.AccessKey != nil {
		decr, err := cipher.AESCipher.Decrypt(cfg.AccessKey)
		if err != nil {
			return ACfg{}, apperr.New("decrypt_access_key_failed", apperr.WithText("failed to decrypt key: "+err.Error()))
		}
		cfg.AccessKey = decr
	}

	if cfg.SSO.OIDC != nil && cfg.SSO.OIDC.ClientSecret != nil {
		decr, err := cipher.AESCipher.Decrypt(cfg.SSO.OIDC.ClientSecret)
		if err != nil {
			return ACfg{}, apperr.New("decrypt_oidc_secret", apperr.WithText("failed to decrypt oidc client secret: "+err.Error()))
		}
		cfg.SSO.OIDC.ClientSecret = decr
	}

	if cfg.SSO.SAML != nil {
		fMap := filesMapBytes(data.GetFiles())
		if d, ok := fMap[model.SAMLMetadataConfigFileKey]; ok {
			cfg.SSO.SAML.MetaDataFile = d
		}
		if d, ok := fMap[model.SAMLKeyConfigFileKey]; ok {
			cfg.SSO.SAML.KeyFile = d
		}
		if d, ok := fMap[model.SAMLCertConfigFileKey]; ok {
			cfg.SSO.SAML.CertFile = d
		}
	}

	return cfg, nil
}

func (c *AConfigurator) Check(newData, lastData configurator.Config) ([]byte, error) {
	var newCfg ARequest
	if err := binding.JSON.BindBody(newData.GetValue(), &newCfg); err != nil {
		return nil, apperr.ErrValidation.WithError(err)
	}

	var lastCfg ACfg
	if err := json.Unmarshal(lastData.GetValue(), &lastCfg); err != nil {
		return nil, err
	}

	if newCfg.AccessKey != nil {
		enc, err := cipher.AESCipher.Encrypt([]byte(*newCfg.AccessKey))
		if err != nil {
			return nil, apperr.New("encrypt_key_failed", apperr.WithText("failed to encrypt key: "+err.Error()))
		}
		lastCfg.AccessKey = enc
	} else if lastCfg.AccessKey == nil {
		return nil, apperr.New("access_key_missing", apperr.WithTextTranslate(translator.Translate{translator.RU: "Ключ шифрования Access токена отсутствует", translator.EN: "Access token encryption key is missing"}), apperr.WithCode(code.InvalidArgument))
	}

	if newCfg.AccessTokenTTL != nil {
		lastCfg.AccessTokenTTL = *newCfg.AccessTokenTTL
	}
	if newCfg.RefreshTokenTTL != nil {
		lastCfg.RefreshTokenTTL = *newCfg.RefreshTokenTTL
	}

	if newCfg.LDAP != nil {
		ld := make([]ACfgLDAP, len(newCfg.LDAP))
		for i, ldap := range newCfg.LDAP {
			ld[i] = ACfgLDAP{
				Domain:  ldap.Domain,
				Secured: ldap.Secured,
				Md5hash: ldap.Md5hash,
				Addrs:   datautil.Distinct(ldap.Addrs),
			}
		}
		lastCfg.LDAP = ld
	}

	if newCfg.SSO != nil {
		var plainClientSecret []byte
		var fromNew bool

		if newCfg.SSO.OIDC != nil {
			if newCfg.SSO.OIDC.ClientSecret != nil {
				plainClientSecret = []byte(*newCfg.SSO.OIDC.ClientSecret)
				fromNew = true
			} else if lastCfg.SSO.OIDC != nil && lastCfg.SSO.OIDC.ClientSecret != nil {
				decr, err := cipher.AESCipher.Decrypt(lastCfg.SSO.OIDC.ClientSecret)
				if err != nil {
					return nil, apperr.New("decrypt_oidc_secret", apperr.WithText("failed to decrypt existing oidc client secret: "+err.Error()))
				}
				plainClientSecret = decr
				fromNew = false
			}
			if plainClientSecret == nil && lastCfg.SSO.Enabled == "oidc" && newCfg.SSO.Enabled == "oidc" {
				return nil, ErrInvalidOIDCSecret
			}
		}

		lastCfg.SSO.Enabled = newCfg.SSO.Enabled
		if newCfg.SSO.Description != nil {
			lastCfg.SSO.Description = *newCfg.SSO.Description
		}

		if newCfg.SSO.OIDC != nil {
			var encryptedSecret []byte
			if plainClientSecret != nil {
				enc, err := cipher.AESCipher.Encrypt(plainClientSecret)
				if err != nil {
					return nil, apperr.New("encrypt_oidc_secret", apperr.WithText("failed to encrypt oidc client secret: "+err.Error()))
				}
				encryptedSecret = enc
			} else if lastCfg.SSO.OIDC != nil {
				encryptedSecret = lastCfg.SSO.OIDC.ClientSecret
			}

			lastCfg.SSO.OIDC = &ACfgOIDC{
				ConfigURL:         newCfg.SSO.OIDC.ConfigURL,
				ClientID:          newCfg.SSO.OIDC.ClientID,
				ClientSecret:      encryptedSecret,
				RootURL:           newCfg.SSO.OIDC.RootURL,
				LoginAttr:         newCfg.SSO.OIDC.LoginAttr,
				ValidRedirectURLs: newCfg.SSO.OIDC.ValidRedirectURLs,
			}
		}

		if newCfg.SSO.SAML != nil {
			lastCfg.SSO.SAML = &ACfgSAML{
				RootURL:           newCfg.SSO.SAML.RootURL,
				LoginAttr:         newCfg.SSO.SAML.LoginAttr,
				ValidRedirectURLs: newCfg.SSO.SAML.ValidRedirectURLs,
			}
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if lastCfg.SSO.Enabled == "saml" {
			if lastCfg.SSO.SAML == nil {
				return nil, apperr.New("saml_config_missing", apperr.WithTextTranslate(translator.Translate{translator.RU: "Конфигурация SAML отсутствует", translator.EN: "SAML configuration is missing"}), apperr.WithCode(code.InvalidArgument))
			}
			fMap := filesMapBytes(lastData.GetFiles())

			if dataBytes, ok := fMap[model.SAMLMetadataConfigFileKey]; ok {
				lastCfg.SSO.SAML.MetaDataFile = dataBytes
			} else if lastCfg.SSO.SAML.MetaDataFile == nil {
				return nil, apperr.New("saml_metadata_not_found", apperr.WithTextTranslate(translator.Translate{translator.RU: "Файл saml.metadata не загружен", translator.EN: "The saml.metadata file was not loaded"}), apperr.WithCode(code.InvalidArgument))
			}

			if dataBytes, ok := fMap[model.SAMLKeyConfigFileKey]; ok {
				lastCfg.SSO.SAML.KeyFile = dataBytes
			} else if lastCfg.SSO.SAML.KeyFile == nil {
				return nil, apperr.New("saml_key_not_found", apperr.WithTextTranslate(translator.Translate{translator.RU: "Файл saml.key не загружен", translator.EN: "The saml.key file was not loaded"}), apperr.WithCode(code.InvalidArgument))
			}

			if dataBytes, ok := fMap[model.SAMLCertConfigFileKey]; ok {
				lastCfg.SSO.SAML.CertFile = dataBytes
			} else if lastCfg.SSO.SAML.CertFile == nil {
				return nil, apperr.New("saml_cert_not_found", apperr.WithTextTranslate(translator.Translate{translator.RU: "Файл saml.cert не загружен", translator.EN: "The saml.cert file was not loaded"}), apperr.WithCode(code.InvalidArgument))
			}

			samlCfg := saml.Config{
				Enabled:           true,
				MetaDataFile:      lastCfg.SSO.SAML.MetaDataFile,
				CertFile:          lastCfg.SSO.SAML.CertFile,
				KeyFile:           lastCfg.SSO.SAML.KeyFile,
				RootURL:           lastCfg.SSO.SAML.RootURL,
				LoginAttr:         lastCfg.SSO.SAML.LoginAttr,
				ValidRedirectURLs: lastCfg.SSO.SAML.ValidRedirectURLs,
			}

			if _, err := saml.NewAuthService(ctx, samlCfg); err != nil {
				return nil, apperr.New("saml_invalid", apperr.WithText("Ошибка проверки конфигурации SAML: "+err.Error()), apperr.WithCode(code.InvalidArgument))
			}

		} else if lastCfg.SSO.Enabled == "oidc" {
			if lastCfg.SSO.OIDC == nil {
				return nil, apperr.New("oidc_config_missing", apperr.WithTextTranslate(translator.Translate{translator.RU: "Конфигурация OIDC отсутствует", translator.EN: "OIDC configuration is missing"}), apperr.WithCode(code.InvalidArgument))
			}

			var plainForValidation []byte
			if fromNew && plainClientSecret != nil {
				plainForValidation = plainClientSecret
			} else if lastCfg.SSO.OIDC.ClientSecret != nil {
				decrypted, err := cipher.AESCipher.Decrypt(lastCfg.SSO.OIDC.ClientSecret)
				if err != nil {
					return nil, apperr.New("decrypt_oidc_secret", apperr.WithText("failed to decrypt oidc client secret: "+err.Error()))
				}
				plainForValidation = decrypted
			}

			if len(plainForValidation) == 0 {
				return nil, ErrInvalidOIDCSecret
			}

			if _, err := oidc.NewAuthService(ctx, oidc.Config{
				Enabled:           true,
				ConfigURL:         lastCfg.SSO.OIDC.ConfigURL,
				ClientID:          lastCfg.SSO.OIDC.ClientID,
				ClientSecret:      string(plainForValidation),
				RootURL:           lastCfg.SSO.OIDC.RootURL,
				LoginAttr:         lastCfg.SSO.OIDC.LoginAttr,
				ValidRedirectURLs: lastCfg.SSO.OIDC.ValidRedirectURLs,
			}); err != nil {
				return nil, apperr.New("oidc_invalid", apperr.WithText("Ошибка проверки конфигурации OIDC: "+err.Error()), apperr.WithCode(code.InvalidArgument))
			}
		}
	}

	if newCfg.Limiter != nil {
		lastCfg.Limiter = ACfgLimiter{
			MaxAttempts: newCfg.Limiter.MaxAttempts,
			TTL:         newCfg.Limiter.TTL,
			BlockingTTL: newCfg.Limiter.BlockingTTL,
		}
	}

	out, err := json.Marshal(lastCfg)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *AConfigurator) AfterUpdate(data configurator.Config) error {
	cfgUn, err := c.Unmarshal(data)
	if err != nil {
		return err
	}
	cfg, ok := cfgUn.(ACfg)
	if !ok {
		return apperr.New("invalid_unmarshal_result", apperr.WithText("unmarshal did not return ACfg"))
	}

	select {
	case c.ch <- cfg:
	default:
		select {
		case <-c.ch:
		default:
		}
		select {
		case c.ch <- cfg:
		default:
		}
	}

	return nil
}

func (c *AConfigurator) Init() ([]byte, error) {
	t, err := randomBase64(32)
	if err != nil {
		return nil, err
	}

	encr, err := cipher.AESCipher.Encrypt([]byte(t))
	if err != nil {
		return nil, apperr.New("encrypt_init_key_failed", apperr.WithText("failed to encrypt initial key: "+err.Error()))
	}

	cfg := ACfg{
		AccessKey:       encr,
		AccessTokenTTL:  15,
		RefreshTokenTTL: 43200,
		LDAP:            []ACfgLDAP{},
		SSO: ACfgSSO{
			Enabled:     "none",
			Description: "SSO",
			OIDC:        &ACfgOIDC{},
			SAML:        &ACfgSAML{},
		},
		Limiter: ACfgLimiter{
			MaxAttempts: 5,
			TTL:         1,
			BlockingTTL: 1,
		},
	}

	return json.Marshal(cfg)
}

func (c *AConfigurator) Transform(data configurator.Config) ([]byte, error) {
	var cfg ACfg
	if err := json.Unmarshal(data.GetValue(), &cfg); err != nil {
		return nil, err
	}

	var tr ATransformer
	if err := json.Unmarshal(data.GetValue(), &tr); err != nil {
		return nil, err
	}

	if cfg.AccessKey != nil {
		tr.IsAccessKey = true
	}

	fNameMap := filesMapNames(data.GetFiles())

	if tr.SSO.SAML != nil {
		if name, ok := fNameMap[model.SAMLMetadataConfigFileKey]; ok {
			tr.SSO.SAML.MetadataFile = name
		}
		if name, ok := fNameMap[model.SAMLKeyConfigFileKey]; ok {
			tr.SSO.SAML.KeyFile = name
		}
		if name, ok := fNameMap[model.SAMLCertConfigFileKey]; ok {
			tr.SSO.SAML.CertFile = name
		}
		if len(tr.SSO.SAML.ValidRedirectURLs) == 0 {
			tr.SSO.SAML.ValidRedirectURLs = make([]string, 0)
		}
	}

	if tr.SSO.OIDC != nil {
		if cfg.SSO.OIDC != nil && cfg.SSO.OIDC.ClientSecret != nil {
			tr.SSO.OIDC.IsClientSecret = true
		}
		if len(tr.SSO.OIDC.ValidRedirectURLs) == 0 {
			tr.SSO.OIDC.ValidRedirectURLs = make([]string, 0)
		}
	}

	return json.Marshal(tr)
}
