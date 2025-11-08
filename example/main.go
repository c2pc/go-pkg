package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/c2pc/go-pkg/v2/auth"
	profile2 "github.com/c2pc/go-pkg/v2/auth/profile"
	"github.com/c2pc/go-pkg/v2/example/internal/config"
	"github.com/c2pc/go-pkg/v2/example/internal/database"
	"github.com/c2pc/go-pkg/v2/example/internal/model"
	"github.com/c2pc/go-pkg/v2/example/internal/repository"
	"github.com/c2pc/go-pkg/v2/example/internal/service"
	"github.com/c2pc/go-pkg/v2/example/internal/transport/api"
	restHandler "github.com/c2pc/go-pkg/v2/example/internal/transport/api/handler"
	"github.com/c2pc/go-pkg/v2/example/profile"
	"github.com/c2pc/go-pkg/v2/task"
	"github.com/c2pc/go-pkg/v2/utils/cache/redis"
	database2 "github.com/c2pc/go-pkg/v2/utils/db"
	"github.com/c2pc/go-pkg/v2/utils/logger"
	"github.com/c2pc/go-pkg/v2/utils/mw"
)

var (
	Vendor  = "FLAT"
	Name    = "example-backend"
	LogPath = "/var/log/flat"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := config.Migrate("config.yml"); err != nil {
		logger.AppFatalFLog(ctx, "[Migrate] %s", err)
		return
	}

	configs, err := config.NewConfig("config.yml")
	if err != nil {
		logger.AppFatalFLog(ctx, "[CONFIG] %s", err)
		return
	}

	db, err := database2.ConnectPostgres(configs.PostgresUrl, 10, 100)
	if err != nil {
		logger.AppFatalFLog(ctx, "[DB] %s", err.Error())
		return
	}

	rdb, err := redis.NewRedisClient(&redis.RedisClient{
		ClusterMode: configs.Redis.ClusterMode,
		Address:     configs.Redis.Address,
		Username:    configs.Redis.Username,
		Password:    configs.Redis.Password,
		MaxRetry:    configs.Redis.MaxRetry,
		DB:          configs.Redis.DB,
	})
	if err != nil {
		logger.AppFatalFLog(ctx, "[REDIS] %s", err.Error())
		return
	}

	profileRepo := profile.NewProfileRepository(db)

	authService, err := auth.New(
		ctx,
		auth.Data{
			Vendor:     Vendor,
			AppName:    Name,
			AppVersion: "1.0.0",
			LogPath:    LogPath,
		},
		auth.Input{
			DB:          db,
			Rdb:         rdb,
			Permissions: model.Permissions,
			Analytic:    auth.Analytic{},
		}, &profile2.Profile{
			Service:     profile.NewService(profileRepo),
			Request:     profile.NewRequest(),
			Transformer: profile.NewTransformer(),
		},
	)
	if err != nil {
		logger.AppFatalFLog(ctx, "[AUTH] %s", err.Error())
		return
	}
	defer authService.Stop(ctx)

	sqlDB, err := db.DB()
	if err != nil {
		logger.AppFatalFLog(ctx, "[DB] %s", err.Error())
		return
	}

	if err := database.Migrate(sqlDB, "postgres"); err != nil {
		logger.AppFatalFLog(ctx, "[DB_MIGRATE] %s", err.Error())
		return
	}

	if err := database.SeedersRun(ctx, db, profileRepo, authService.GetAdminID()); err != nil {
		logger.AppFatalFLog(ctx, "[DB] %s", err.Error())
		return
	}

	repositories := repository.NewRepositories(db)
	services := service.NewServices(service.Deps{Repositories: repositories})

	authService.Task().InitConsumers(task.Consumers{
		"news": services.News,
	})

	err = authService.Start(ctx)
	if err != nil {
		logger.AppFatalFLog(ctx, "[AUTH] %s", err.Error())
	}

	trx := mw.NewTransaction(db)
	restHandlers := restHandler.NewHandlers(authService, services, trx)
	restServer := api.NewServer(api.Input{
		Host: configs.HTTP.Host,
		Port: configs.HTTP.Port,
	}, restHandlers.Init())

	logger.AppInfoFLog(ctx, "Starting REST Server")
	defer logger.AppInfoFLog(ctx, "Stopping REST Server")

	go func() {
		if err := restServer.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			logger.AppFatalFLog(ctx, "Failed to start Server: %s", err.Error())
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)

	<-quit

	const timeout = 5 * time.Second
	ctx3, shutdown := context.WithTimeout(ctx, timeout)
	defer shutdown()

	if err := restServer.Stop(ctx3); err != nil {
		logger.AppWarningFLog(ctx, "Failed to stop Server: %s", err.Error())
	}
}
