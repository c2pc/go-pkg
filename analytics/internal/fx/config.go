package fx

import (
	"sync/atomic"

	"github.com/rs/zerolog"
)

type Config struct {
	Level   zerolog.Level
	Enabled bool
}

type ConfigHolder struct{ v atomic.Value }

func NewConfigHolder(cfg Config) *ConfigHolder {
	h := &ConfigHolder{}
	h.Reload(cfg)
	return h
}

func (h *ConfigHolder) Get() *Config {
	return h.v.Load().(*Config)
}

func (h *ConfigHolder) Reload(cfg Config) {
	svc := &Config{
		Level:   cfg.Level,
		Enabled: cfg.Enabled,
	}
	h.v.Store(svc)
	return
}
