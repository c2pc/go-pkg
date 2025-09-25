package config

import (
	"github.com/c2pc/go-pkg/v2/utils/config"
)

type Config struct {
	PostgresUrl string       `yaml:"postgres_url"`
	HTTP        config.HTTP  `yaml:"http"`
	Redis       config.Redis `yaml:"redis"`
}

func NewConfig(configPath string) (*Config, error) {
	cfg, err := config.NewConfig[Config](configPath)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}
