package auth

import (
	"context"

	authConf "github.com/c2pc/go-pkg/v2/auth/internal/configurator"
	"github.com/c2pc/go-pkg/v2/auth/internal/model"
	"github.com/c2pc/go-pkg/v2/auth/internal/service"
	"github.com/c2pc/go-pkg/v2/utils/app_data"
	"github.com/c2pc/go-pkg/v2/utils/logger"
)

func getLogCfg(ctx context.Context, configService service.ConfigService, logConfig *authConf.LogConfigurator) (authConf.LogCfg, error) {
	cfg, err := configService.SetConfig(ctx, model.LogConfigKey, logConfig)
	if err != nil {
		return authConf.LogCfg{}, err
	}

	logCfg, err := logConfig.Unmarshal(cfg)
	if err != nil {
		return authConf.LogCfg{}, err
	}

	return logCfg.(authConf.LogCfg), nil
}

func (a *Auth) logConfigWatcher(ctx context.Context, logConfig *authConf.LogConfigurator) {
	for {
		select {
		case <-ctx.Done():
			return
		case newCfg := <-logConfig.Watch():
			resetLog(newCfg)
		}
	}
}

func resetLog(cfg authConf.LogCfg) {
	if app_data.IsDevMode() {
		cfg.Level = 0
	}

	logger.Init(logger.Config{
		Level:      logger.Level(cfg.Level),
		Path:       cfg.Path,
		Filename:   "app.log",
		MaxSizeMB:  cfg.MaxSizeMB,
		MaxBackups: cfg.MaxBackups,
		MaxAgeDays: cfg.MaxAgeDays,
		Compress:   cfg.Compress,
	})
}
