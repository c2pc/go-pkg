package fx

import (
	"sync/atomic"

	"github.com/c2pc/go-pkg/v2/utils/sso/ldap"
)

type LDAPHolder struct{ v atomic.Value }

func NewLDAPHolder(enabled bool, cfg map[string]ldap.Config) *LDAPHolder {
	h := &LDAPHolder{}
	h.Reload(enabled, cfg)
	return h
}

func (h *LDAPHolder) Get() *ldap.Auth { return h.v.Load().(*ldap.Auth) }

func (h *LDAPHolder) Reload(enabled bool, cfg map[string]ldap.Config) {
	svc := ldap.NewAuthService(enabled, cfg)
	h.v.Store(svc)
}
