package config

import "time"

type Config struct {
	ServerURL string `yaml:"server_url"`
	SecretKey string `yaml:"secret_key"`
	ServerID  int    `yaml:"server_id"`
	Timeout   time.Duration
	Debug     string
}
