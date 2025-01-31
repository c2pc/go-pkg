package config

const (
	DefaultAccessTokenTTL  = 15           // Значение по умолчанию для TTL Access Token в минутах
	DefaultRefreshTokenTTL = 24 * 60 * 30 // Значение по умолчанию для TTL Refresh Token в минутах (30 дней)
	DefaultTimeout         = 10
)

type AUTH struct {
	Key             string   `yaml:"key"`                         // Ключ для аутентификации
	AccessTokenTTL  float64  `yaml:"access_token_ttl,omitempty"`  // TTL Access Token в минутах (опционально)
	RefreshTokenTTL float64  `yaml:"refresh_token_ttl,omitempty"` // TTL Refresh Token в минутах (опционально)
	LDAPConfig      LDAPAuth `yaml:"ldap"`
}

type LDAPAuth struct {
	Enable    bool   `yaml:"enable"`
	ServerURL string `yaml:"server_url"`
	SecretKey string `yaml:"secret_key"`
	ServerID  int    `yaml:"server_id"`
	Timeout   int    `yaml:"timeout,omitempty"`
}
