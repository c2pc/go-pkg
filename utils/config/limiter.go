package config

const (
	DefaultTTL         = 15
	DefaultMaxAttempts = 3
)

type Limiter struct {
	MaxAttempts int `yaml:"max_attempts,omitempty"`
	TTL         int `yaml:"ttl,omitempty"`
}
