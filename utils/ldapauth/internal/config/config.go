package config

import "time"

type Config struct {
	IsEnabled bool
	ServerURL string
	SecretKey string
	ServerID  int
	Timeout   time.Duration
	Debug     string
}
