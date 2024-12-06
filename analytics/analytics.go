package analytics

import (
	"time"

	"github.com/c2pc/go-pkg/v2/analytics/internal/repository"
	"github.com/c2pc/go-pkg/v2/analytics/internal/service"
	"github.com/c2pc/go-pkg/v2/analytics/internal/transport/api/handlers"
	collector "github.com/c2pc/go-pkg/v2/analytics/internal/transport/api/middlewares"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Analytics interface {
	InitHandler(api *gin.RouterGroup)
	InitCollect() gin.HandlerFunc
}

type analyticsImpl struct {
	handler   *handlers.AnalyticsHandler
	collector gin.HandlerFunc
}

type Config struct {
	DB            *gorm.DB
	FlushInterval time.Duration
	BatchSize     int
}

func New(config Config) Analytics {
	repo := repository.NewAnalyticRepository(config.DB)

	service := service.NewAnalyticService(repo)

	handler := handlers.NewAnalyticsHandler(service)

	collectorConfig := collector.LoggerConfig{
		DB:            config.DB,
		FlushInterval: config.FlushInterval,
		BatchSize:     config.BatchSize,
	}

	logger := collector.New(collectorConfig)

	return &analyticsImpl{
		handler:   handler,
		collector: logger,
	}
}

func (a *analyticsImpl) InitHandler(api *gin.RouterGroup) {
	a.handler.Init(api)
}

func (a *analyticsImpl) InitCollect() gin.HandlerFunc {
	return a.collector
}