package config

const (
	DefaultTimeout = 10
)

type LDAPAuth struct {
	Enable    bool   `yaml:"enable"`
	ServerURL string `yaml:"server_url"`
	SecretKey string `yaml:"secret_key"`
	ServerID  int    `yaml:"server_id"`
	Timeout   int    `yaml:"timeout,omitempty"`
}
