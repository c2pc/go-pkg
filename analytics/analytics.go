package analytics

import (
	"context"
	"encoding/json"

	"github.com/c2pc/go-pkg/v2/analytics/internal/cache"
	"github.com/c2pc/go-pkg/v2/analytics/internal/cache/cachekey"
	logConf "github.com/c2pc/go-pkg/v2/analytics/internal/configurator"
	"github.com/c2pc/go-pkg/v2/analytics/internal/fx"
	"github.com/c2pc/go-pkg/v2/analytics/internal/repository"
	"github.com/c2pc/go-pkg/v2/analytics/internal/service"
	"github.com/c2pc/go-pkg/v2/analytics/internal/transport/api/handlers"
	collector "github.com/c2pc/go-pkg/v2/analytics/internal/transport/api/middlewares"
	"github.com/c2pc/go-pkg/v2/auth_config"
	"github.com/c2pc/go-pkg/v2/utils/logger"
	"github.com/c2pc/go-pkg/v2/utils/syslog"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
	"gorm.io/gorm"
)

type Analytics interface {
	InitHandler(api *gin.RouterGroup, handlers ...gin.HandlerFunc)
	CollectAnalytic(c *gin.Context)
	ShutDown()
}

type analyticsImpl struct {
	handler      *handlers.AnalyticsHandler
	middleware   gin.HandlerFunc
	shutdown     func()
	cfg          logConf.Cfg
	configHolder *fx.ConfigHolder
}

type Config struct {
	Rdb                 redis.UniversalClient
	DB                  *gorm.DB
	ExcludeInputBodies  map[string][]string
	ExcludeOutputBodies map[string][]string
	SkipRequests        map[string][]string
	HiddenKeys          []string
	Configurator        auth_config.Config
	LogPath             string
}

func New(ctx context.Context, serviceName string, config Config) (Analytics, error) {
	cachekey.SetServiceName(serviceName)

	cch := cache.NewCache(config.Rdb)

	logPath := config.LogPath
	if logPath == "" {
		logPath = "/var/log/flat"
	}
	loggerConfig := logConf.NewConfigurator(logPath)

	err := config.Configurator.SetConfig(ctx, "logger", loggerConfig)
	if err != nil {
		return nil, err
	}

	cfgByte, err := config.Configurator.GetService().GetWithoutTransform(ctx, "logger")
	if err != nil {
		return nil, err
	}

	cfgW, err := loggerConfig.Unmarshal(cfgByte.Value)
	if err != nil {
		return nil, err
	}

	cfg := cfgW.(logConf.Cfg)

	repo := repository.NewAnalyticRepository(config.DB)
	svc := service.NewAnalyticService(repo)
	handler := handlers.NewAnalyticsHandler(svc)

	if config.SkipRequests == nil {
		config.SkipRequests = make(map[string][]string)
	}

	config.SkipRequests["/auth/settings"] = []string{}
	config.SkipRequests["/stream"] = []string{}
	config.SkipRequests["/version"] = []string{}

	configHolder := fx.NewConfigHolder(fx.Config{
		Level:   zerolog.Level(cfg.Level),
		Enabled: cfg.DB.Enabled,
	})

	collectorConfig := collector.LoggerConfig{
		DB:                  config.DB,
		FlushInterval:       10,
		BatchSize:           100,
		ExcludeInputBodies:  config.ExcludeInputBodies,
		ExcludeOutputBodies: config.ExcludeOutputBodies,
		SkipRequests:        config.SkipRequests,
		HiddenKeys:          config.HiddenKeys,
	}

	middleware, shutdown := collector.New(collectorConfig, configHolder, cch)

	analytics := &analyticsImpl{
		handler:      handler,
		middleware:   middleware,
		shutdown:     shutdown,
		cfg:          cfg,
		configHolder: configHolder,
	}

	analytics.resetLog(cfg)

	go analytics.configWatcher(ctx, loggerConfig)

	return analytics, nil
}

func (a *analyticsImpl) InitHandler(api *gin.RouterGroup, handlers ...gin.HandlerFunc) {
	a.handler.Init(api, handlers...)
}

func (a *analyticsImpl) CollectAnalytic(c *gin.Context) {
	a.middleware(c)
}

func (a *analyticsImpl) ShutDown() {
	a.shutdown()
}

func (a *analyticsImpl) resetLog(cfg logConf.Cfg) {
	c := logger.Config{
		Level: zerolog.Level(cfg.Level),
		File: &logger.FileConfig{
			Enabled:    cfg.File.Enabled,
			Path:       cfg.File.Path,
			MaxSizeMB:  cfg.File.MaxSizeMB,
			MaxBackups: cfg.File.MaxBackups,
			MaxAgeDays: cfg.File.MaxAgeDays,
			Compress:   cfg.File.Compress,
		},
	}

	logger.Reload(c)
}

func (a *analyticsImpl) resetSyslog(cfg logConf.Cfg) {
	addrs := make([]syslog.AddrConfig, len(cfg.Syslog))
	for i, addr := range cfg.Syslog {
		addrs[i] = syslog.AddrConfig{
			Network: addr.Network,
			Addr:    addr.Addr,
		}
	}

	c := syslog.Config{
		Level: zerolog.Level(cfg.Level),
		Addrs: addrs,
	}

	_ = syslog.Reload(c)
}

func (a *analyticsImpl) configWatcher(ctx context.Context, loggerConfig *logConf.Configurator) {
	for {
		select {
		case <-ctx.Done():
			return
		case newCfg := <-loggerConfig.Watch():
			oldCfg := a.cfg

			m1, _ := json.Marshal(newCfg.File)
			m2, _ := json.Marshal(oldCfg.File)

			if oldCfg.Level != newCfg.Level || string(m1) != string(m2) {
				a.resetLog(newCfg)
			}

			m3, _ := json.Marshal(newCfg.Syslog)
			m4, _ := json.Marshal(oldCfg.Syslog)
			if oldCfg.Level != newCfg.Level || string(m3) != string(m4) {
				a.resetSyslog(newCfg)
			}

			m5, _ := json.Marshal(newCfg.DB)
			m6, _ := json.Marshal(oldCfg.DB)

			if oldCfg.Level != newCfg.Level || string(m5) != string(m6) {
				a.configHolder.Reload(fx.Config{
					Level:   zerolog.Level(newCfg.Level),
					Enabled: newCfg.DB.Enabled,
				})
			}

			a.cfg = newCfg
		}
	}

}
