package fx

import (
	"context"
	"strings"
	"sync/atomic"

	"github.com/c2pc/go-pkg/v2/utils/logger"
	"github.com/c2pc/go-pkg/v2/utils/sso/oidc"
)

type OIDCHolder struct{ v atomic.Value }

func NewOIDCHolder(ctx context.Context, cfg oidc.Config) (*OIDCHolder, error) {
	h := &OIDCHolder{}
	err := h.Reload(ctx, cfg)
	if err != nil {
		return nil, err
	}

	return h, nil
}
func (h *OIDCHolder) Get() *oidc.Auth { return h.v.Load().(*oidc.Auth) }

func (h *OIDCHolder) Reload(ctx context.Context, cfg oidc.Config) error {
	cfg.RootURL = strings.TrimRight(cfg.RootURL, "/") + "/api/v1/auth/sso/callback"
	svc, err := oidc.NewAuthService(ctx, cfg)
	if err != nil {
		logger.AppErrorFLog(ctx, "oidc error: %s\n", err.Error())
	}
	h.v.Store(svc)
	return nil
}
