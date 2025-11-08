package auth

import (
	"context"
	"encoding/json"

	authConf "github.com/c2pc/go-pkg/v2/auth/internal/configurator"
	"github.com/c2pc/go-pkg/v2/auth/internal/fx"
	"github.com/c2pc/go-pkg/v2/auth/internal/model"
	"github.com/c2pc/go-pkg/v2/auth/internal/service"
	"github.com/c2pc/go-pkg/v2/utils/syslog"
)

func getAuditCfg(ctx context.Context, configService service.ConfigService, auditConfig *authConf.AuditConfigurator) (authConf.AuditCfg, error) {
	cfg, err := configService.SetConfig(ctx, model.AuditConfigKey, auditConfig)
	if err != nil {
		return authConf.AuditCfg{}, err
	}

	auditCfg, err := auditConfig.Unmarshal(cfg)
	if err != nil {
		return authConf.AuditCfg{}, err
	}

	return auditCfg.(authConf.AuditCfg), nil
}

func resetSyslog(cfg authConf.AuditCfg) {
	syslogAddrs := make([]syslog.ConfigSyslog, len(cfg.File.Syslog))
	for i, addr := range cfg.File.Syslog {
		syslogAddrs[i] = syslog.ConfigSyslog{
			Network: addr.Network,
			Addr:    addr.Addr,
		}
	}

	c := syslog.Config{
		Path:       cfg.File.Path,
		Filename:   "audit.log",
		MaxSizeMB:  cfg.File.MaxSizeMB,
		MaxBackups: cfg.File.MaxBackups,
		MaxAgeDays: cfg.File.MaxAgeDays,
		Compress:   cfg.File.Compress,
		Syslog:     syslogAddrs,
	}

	syslog.Init(c)
}

func (a *Auth) auditConfigWatcher(ctx context.Context, auditConfig *authConf.AuditConfigurator) {
	for {
		select {
		case <-ctx.Done():
			return
		case newCfg := <-auditConfig.Watch():
			oldCfg := a.auditCfg

			m3, _ := json.Marshal(newCfg.File)
			m4, _ := json.Marshal(oldCfg.File)
			if string(m3) != string(m4) {
				resetSyslog(newCfg)
			}

			m5, _ := json.Marshal(newCfg.DB)
			m6, _ := json.Marshal(oldCfg.DB)

			if string(m5) != string(m6) {
				a.auditHolder.Reload(fx.Auditor{
					Admin: fx.AuditorDB{
						MaxAgeDays:         newCfg.DB.Admin.MaxAgeDays,
						MaxCount:           newCfg.DB.Admin.MaxCount,
						PercentageOfDelete: newCfg.DB.Admin.PercentageOfDelete,
						LogDisabled:        false,
					},
				})
			}

			a.auditCfg = newCfg
		}
	}

}
