package config

import (
	"github.com/c2pc/go-pkg/v2/utils/config"
	"os"
	"strings"
)

type Config struct {
	PasswordSalt string       `yaml:"password_salt"`
	PostgresUrl  string       `yaml:"postgres_url"`
	APP          config.APP   `yaml:"app"`
	HTTP         config.HTTP  `yaml:"http"`
	AUTH         config.AUTH  `yaml:"auth"`
	Redis        config.Redis `yaml:"redis"`
}

func NewConfig(configPath string) (*Config, error) {
	cfg, err := config.NewConfig[Config](configPath)
	if err != nil {
		return nil, err
	}

	if cfg.AUTH.AccessTokenTTL == 0 {
		cfg.AUTH.AccessTokenTTL = config.DefaultAccessTokenTTL
	}

	if cfg.AUTH.RefreshTokenTTL == 0 {
		cfg.AUTH.RefreshTokenTTL = config.DefaultRefreshTokenTTL
	}

	for _, e := range os.Environ() {
		pair := strings.SplitN(e, "=", 2)
		if len(pair) == 2 && pair[1] != "" {
			switch pair[0] {
			case "POSTGRES_URL":
				cfg.PostgresUrl = pair[1]
			case "PASSWORD_SALT":
				cfg.PasswordSalt = pair[1]
			case "APP_LOG":
				cfg.APP.LogDir = pair[1]
			case "JWT_SIGNING_KEY":
				cfg.AUTH.Key = pair[1]
			}
		}
	}
	return cfg, nil
}
