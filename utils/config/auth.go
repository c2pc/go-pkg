package config

const (
	DefaultAccessTokenTTL  = 15           // Значение по умолчанию для TTL Access Token в минутах
	DefaultRefreshTokenTTL = 24 * 60 * 30 // Значение по умолчанию для TTL Refresh Token в минутах (30 дней)
	DefaultTimeout         = 10
)

type AUTH struct {
	Key             string  `yaml:"key"`                         // Ключ для аутентификации
	AccessTokenTTL  float64 `yaml:"access_token_ttl,omitempty"`  // TTL Access Token в минутах (опционально)
	RefreshTokenTTL float64 `yaml:"refresh_token_ttl,omitempty"` // TTL Refresh Token в минутах (опционально)
	LDAP            LDAP    `yaml:"ldap"`
	SSO             SSO     `yaml:"sso"`
}

type LDAP struct {
	Enabled    bool   `yaml:"enabled"`
	Addr       string `yaml:"addr"`
	BaseDN     string `yaml:"base_dn"`
	BaseFilter string `yaml:"base_filter"`
	LoginAttr  string `yaml:"login_attr"`
	Domain     string `yaml:"domain"`
}

type SSO struct {
	OIDC OIDC `yaml:"oidc"`
	SAML SAML `yaml:"saml"`
}

type OIDC struct {
	Enabled           bool     `yaml:"enabled"`
	ConfigURL         string   `yaml:"config_url"`
	ClientID          string   `yaml:"client_id"`
	ClientSecret      string   `yaml:"client_secret"`
	RootURL           string   `yaml:"root_url"`
	LoginAttr         string   `yaml:"login_attr"`
	ValidRedirectURLs []string `yaml:"valid_redirect_urls"`
}

type SAML struct {
	Enabled           bool     `yaml:"enabled"`
	MetaDataURL       string   `yaml:"meta_data_url"`
	MetaDataPath      string   `yaml:"meta_data_path"`
	CertFile          string   `yaml:"cert_file"`
	KeyFile           string   `yaml:"key_file"`
	RootURL           string   `yaml:"root_url"`
	LoginAttr         string   `yaml:"login_attr"`
	ValidRedirectURLs []string `yaml:"valid_redirect_urls"`
}
