package fx

import (
	"sync/atomic"
	"time"
)

type AuthConfig struct {
	AccessExpire  time.Duration
	RefreshExpire time.Duration
	AccessSecret  string
}

type AuthHolder struct{ v atomic.Value }

func NewAuthHolder(accessExpire time.Duration, refreshExpire time.Duration, accessSecret string) *AuthHolder {
	h := &AuthHolder{}
	h.Reload(accessExpire, refreshExpire, accessSecret)
	return h
}

func (h *AuthHolder) Get() *AuthConfig {
	return h.v.Load().(*AuthConfig)
}

func (h *AuthHolder) Reload(accessExpire time.Duration, refreshExpire time.Duration, accessSecret string) {
	svc := &AuthConfig{
		AccessExpire:  accessExpire,
		RefreshExpire: refreshExpire,
		AccessSecret:  accessSecret,
	}
	h.v.Store(svc)
	return
}
