package fx

import (
	"context"
	"strings"
	"sync/atomic"

	"github.com/c2pc/go-pkg/v2/utils/logger"
	"github.com/c2pc/go-pkg/v2/utils/sso/saml"
)

type SAMLHolder struct{ v atomic.Value }

func NewSAMLHolder(ctx context.Context, cfg saml.Config) (*SAMLHolder, error) {
	h := &SAMLHolder{}
	err := h.Reload(ctx, cfg)
	if err != nil {
		return nil, err
	}

	return h, nil
}
func (h *SAMLHolder) Get() *saml.Auth { return h.v.Load().(*saml.Auth) }

func (h *SAMLHolder) Reload(ctx context.Context, cfg saml.Config) error {
	cfg.RootURL = strings.TrimRight(cfg.RootURL, "/") + "/api/v1/auth/sso/login"

	svc, err := saml.NewAuthService(ctx, cfg)
	if err != nil {
		logger.AppErrorFLog(ctx, "saml error: %s\n", err.Error())
	}
	h.v.Store(svc)
	return nil
}
