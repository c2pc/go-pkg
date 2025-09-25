package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/c2pc/go-pkg/v2/analytics"
	"github.com/c2pc/go-pkg/v2/auth"
	profile2 "github.com/c2pc/go-pkg/v2/auth/profile"
	"github.com/c2pc/go-pkg/v2/auth_config"
	"github.com/c2pc/go-pkg/v2/example/internal/config"
	database3 "github.com/c2pc/go-pkg/v2/example/internal/database"
	"github.com/c2pc/go-pkg/v2/example/internal/model"
	"github.com/c2pc/go-pkg/v2/example/internal/repository"
	"github.com/c2pc/go-pkg/v2/example/internal/service"
	"github.com/c2pc/go-pkg/v2/example/internal/transport/api"
	restHandler "github.com/c2pc/go-pkg/v2/example/internal/transport/api/handler"
	profile3 "github.com/c2pc/go-pkg/v2/example/profile"
	"github.com/c2pc/go-pkg/v2/task"
	"github.com/c2pc/go-pkg/v2/utils/cache/redis"
	"github.com/c2pc/go-pkg/v2/utils/constant"
	database "github.com/c2pc/go-pkg/v2/utils/db"
	"github.com/c2pc/go-pkg/v2/utils/logger"
	"github.com/c2pc/go-pkg/v2/utils/mcontext"
	"github.com/c2pc/go-pkg/v2/utils/mw"
	"github.com/c2pc/go-pkg/v2/utils/syslog"
	"github.com/c2pc/go-pkg/v2/websocket"
)

func main() {
	ctx, cancel := context.WithCancel(mcontext.WithOperationIDContext(context.Background(), strconv.FormatInt(time.Now().UnixMilli(), 10)))
	defer cancel()

	logger.AppName = "example"
	syslog.Vendor = "Example"
	syslog.Product = "Product"
	syslog.Version = "1.0.0"
	logger.Init(logger.DefaultLoggerConfig)
	defer logger.Close()

	if err := config.Migrate("config.yml"); err != nil {
		log.Fatal("[Migrate] ", err)
		return
	}

	configs, err := config.NewConfig("config.yml")
	if err != nil {
		log.Fatal("[CONFIG]", err)
		return
	}

	db, err := database.ConnectPostgres(configs.PostgresUrl, 10, 100)
	if err != nil {
		log.Fatalf("[DB] %s", err.Error())
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
		logger.Error().Msgf("[REDIS] %s", err.Error())
		return
	}

	trx := mw.NewTransaction(db)
	authConfigService := auth_config.NewAuthConfig(db, trx)

	analyticService, err := analytics.New(ctx, "example", analytics.Config{
		DB:           db,
		SkipRequests: map[string][]string{},
		Configurator: authConfigService,
		LogPath:      "logs",
		Rdb:          rdb,
	})
	if err != nil {
		logger.Error().Msgf("[ANALYSIS] %s", err.Error())
		return
	}
	defer analyticService.ShutDown()

	sqlDB, err := db.DB()
	if err != nil {
		logger.Error().Msgf("[DB] %s", err.Error())
		return
	}

	if err := database3.Migrate(sqlDB, "postgres"); err != nil {
		logger.Error().Msgf("[DB_MIGRATE] %s", err.Error())
		return
	}

	profileRepo := profile3.NewProfileRepository(db)

	authService, err := auth.New(
		ctx,
		"example",
		"0.0.1",
		auth.Input{
			DB:           db,
			Rdb:          rdb,
			Transaction:  trx,
			Permissions:  model.Permissions,
			Configurator: authConfigService,
		}, &profile2.Profile{
			Service:     profile3.NewService(profileRepo),
			Request:     profile3.NewRequest(),
			Transformer: profile3.NewTransformer(),
		})

	if err != nil {
		logger.Error().Msgf("[AUTH] %s", err.Error())
		return
	}

	if err := database3.SeedersRun(ctx, db, profileRepo, authService.GetAdminID()); err != nil {
		logger.Error().Msgf("[DB] %s", err.Error())
		return
	}

	repositories := repository.NewRepositories(db)
	services := service.NewServices(service.Deps{Repositories: repositories})
	ws := websocket.New(10)

	taskService, err := task.NewTask(ctx, task.Config{
		DB:          db,
		Transaction: trx,
		Services: task.Consumers{
			"news": services.News,
		},
		TokenString: "787hhjvYTYTcfcr6556tCTTYChgUYy",
		WS:          ws,
	})
	if err != nil {
		logger.Error().Msgf("[TASK] %s", err.Error())
		return
	}

	restHandlers := restHandler.NewHandlers(authService, authConfigService, services, trx, taskService, analyticService, ws)
	restServer := api.NewServer(api.Input{
		Host: configs.HTTP.Host,
		Port: configs.HTTP.Port,
	}, restHandlers.Init())

	go func() {
		logger.Info().Str(string(constant.OperationID), mcontext.GetOperationID2(ctx)).Msgf("Starting Rest Server")
		fmt.Printf("Starting Rest Server\n")
		if err := restServer.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			logger.Error().Str(string(constant.OperationID), mcontext.GetOperationID2(ctx)).Msgf("Rest ListenAndServe err: %s\n", err.Error())
			fmt.Printf("Rest ListenAndServe err: %s\n", err.Error())
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)

	<-quit

	const timeout = 5 * time.Second
	ctx3, shutdown := context.WithTimeout(ctx, timeout)
	defer shutdown()

	if err := restServer.Stop(ctx3); err != nil {
		logger.Error().Str(string(constant.OperationID), mcontext.GetOperationID2(ctx)).Msgf("Failed to Stop Server: %v", err)
		fmt.Printf("Failed to Stop Server: %s\n", err.Error())
	}

	logger.Info().Str(string(constant.OperationID), mcontext.GetOperationID2(ctx)).Msgf("Shutting Down Server")
	fmt.Printf("Starting Rest Server\n")
}
