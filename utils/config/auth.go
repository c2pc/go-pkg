package config

const (
	DefaultAccessTokenTTL  = 15
	DefaultRefreshTokenTTL = 24 * 60 * 30
)

type AUTH struct {
	Key             string  `yaml:"key"`
	AccessTokenTTL  float64 `yaml:"access_token_ttl,omitempty"`
	RefreshTokenTTL float64 `yaml:"refresh_token_ttl,omitempty"`
}
