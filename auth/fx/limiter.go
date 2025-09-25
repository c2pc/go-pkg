package fx

import (
	"sync/atomic"
	"time"
)

type ConfigLimiter struct {
	MaxAttempts int
	TTL         time.Duration
}

type LimiterHolder struct {
	v atomic.Value
}

func NewLimiterHolder(cfg ConfigLimiter) *LimiterHolder {
	h := &LimiterHolder{}
	h.Reload(cfg)

	return h
}
func (h *LimiterHolder) Get() *ConfigLimiter { return h.v.Load().(*ConfigLimiter) }

func (h *LimiterHolder) Reload(cfg ConfigLimiter) {
	h.v.Store(&ConfigLimiter{
		MaxAttempts: cfg.MaxAttempts,
		TTL:         cfg.TTL,
	})
	return
}
