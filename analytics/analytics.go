package analytics

import (
	"github.com/c2pc/go-pkg/v2/analytics/internal/repository"
	"github.com/c2pc/go-pkg/v2/analytics/internal/service"
	"github.com/c2pc/go-pkg/v2/analytics/internal/transport/api/handlers"
	collector "github.com/c2pc/go-pkg/v2/analytics/internal/transport/api/middlewares"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Analytics interface {
	InitHandler(api *gin.RouterGroup)
	CollectAnalytic(c *gin.Context)
	ShutDown()
}

type analyticsImpl struct {
	handler    *handlers.AnalyticsHandler
	middleware gin.HandlerFunc
	shutdown   func()
}

type Config struct {
	DB                  *gorm.DB
	FlushInterval       int
	BatchSize           int
	ExcludeInputBodies  map[string][]string
	ExcludeOutputBodies map[string][]string
	SkipRequests        map[string][]string
}

func New(config Config) Analytics {
	repo := repository.NewAnalyticRepository(config.DB)

	svc := service.NewAnalyticService(repo)

	handler := handlers.NewAnalyticsHandler(svc)

	if config.SkipRequests == nil {
		config.SkipRequests = make(map[string][]string)
	}

	config.SkipRequests["/auth/settings"] = []string{}
	config.SkipRequests["/stream"] = []string{}
	config.SkipRequests["/version"] = []string{}

	collectorConfig := collector.LoggerConfig{
		DB:                  config.DB,
		FlushInterval:       config.FlushInterval,
		BatchSize:           config.BatchSize,
		ExcludeInputBodies:  config.ExcludeInputBodies,
		ExcludeOutputBodies: config.ExcludeOutputBodies,
		SkipRequests:        config.SkipRequests,
	}

	middleware, shutdown := collector.New(collectorConfig)

	return &analyticsImpl{
		handler:    handler,
		middleware: middleware,
		shutdown:   shutdown,
	}
}

func (a *analyticsImpl) InitHandler(api *gin.RouterGroup) {
	a.handler.Init(api)
}

func (a *analyticsImpl) CollectAnalytic(c *gin.Context) {
	a.middleware(c)
}

func (a *analyticsImpl) ShutDown() {
	a.shutdown()
}
