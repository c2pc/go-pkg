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
	DB            *gorm.DB
	FlushInterval int
	BatchSize     int
}

func New(config Config) Analytics {
	repo := repository.NewAnalyticRepository(config.DB)

	svc := service.NewAnalyticService(repo)

	handler := handlers.NewAnalyticsHandler(svc)

	collectorConfig := collector.LoggerConfig{
		DB:            config.DB,
		FlushInterval: config.FlushInterval,
		BatchSize:     config.BatchSize,
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
