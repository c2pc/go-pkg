package main

import (
	"context"
	"errors"
	profile2 "github.com/c2pc/go-pkg/v2/auth/profile"
	profile3 "github.com/c2pc/go-pkg/v2/example/profile"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/c2pc/go-pkg/v2/analytics"
	"github.com/c2pc/go-pkg/v2/auth"
	"github.com/c2pc/go-pkg/v2/example/internal/config"
	database3 "github.com/c2pc/go-pkg/v2/example/internal/database"
	"github.com/c2pc/go-pkg/v2/example/internal/model"
	"github.com/c2pc/go-pkg/v2/example/internal/repository"
	"github.com/c2pc/go-pkg/v2/example/internal/service"
	"github.com/c2pc/go-pkg/v2/example/internal/transport/api"
	restHandler "github.com/c2pc/go-pkg/v2/example/internal/transport/api/handler"
	sse2 "github.com/c2pc/go-pkg/v2/sse"
	"github.com/c2pc/go-pkg/v2/task"
	"github.com/c2pc/go-pkg/v2/utils/cache/redis"
	database "github.com/c2pc/go-pkg/v2/utils/db"
	"github.com/c2pc/go-pkg/v2/utils/logger"
	"github.com/c2pc/go-pkg/v2/utils/mcontext"
	"github.com/c2pc/go-pkg/v2/utils/mw"
	"github.com/c2pc/go-pkg/v2/utils/secret"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := config.Migrate("config.yml"); err != nil {
		log.Fatal("[Migrate] ", err)
		return
	}

	configs, err := config.NewConfig("config.yml")
	if err != nil {
		log.Fatal("[CONFIG]", err)
		return
	}

	logger.Initialize(logger.Config{
		MachineReadable: false,
		Filename:        "app.log",
		Dir:             configs.LOG.Dir,
		MaxSize:         configs.LOG.MaxSize,
		MaxBackups:      configs.LOG.MaxBackups,
		MaxAge:          configs.LOG.MaxAge,
		Compress:        configs.LOG.Compress,
	})

	db, err := database.ConnectPostgres(configs.PostgresUrl, configs.LOG.Debug, 10, 100)
	if err != nil {
		logger.Fatalf("[DB] %s", err.Error())
		return
	}

	sqlDB, err := db.DB()
	if err != nil {
		logger.Fatalf("[DB] %s", err.Error())
		return
	}

	if err := database3.Migrate(sqlDB, "postgres"); err != nil {
		logger.Fatalf("[DB_MIGRATE] %s", err.Error())
		return
	}

	rdb, err := redis.NewRedisClient(&redis.RedisClient{
		ClusterMode: configs.Redis.ClusterMode,
		Address:     configs.Redis.Address,
		Username:    configs.Redis.Username,
		Password:    configs.Redis.Password,
		MaxRetry:    configs.Redis.MaxRetry,
		DB:          configs.Redis.DB,
	}, configs.LOG.Debug)
	if err != nil {
		logger.Fatalf("[REDIS] %s", err.Error())
		return
	}

	hasher, err := secret.New(configs.PasswordSalt)
	if err != nil {
		logger.Fatalf("[SECRET] %s", err.Error())
		return
	}

	trx := mw.NewTransaction(db)

	profileRepo := profile3.NewProfileRepository(db)

	authService, err := auth.New[profile3.Profile, profile3.ProfileCreateInput, profile3.ProfileUpdateInput, profile3.ProfileUpdateProfileInput](auth.Config{
		Debug:         configs.LOG.Debug,
		DB:            db,
		Rdb:           rdb,
		Transaction:   trx,
		Hasher:        hasher,
		AccessExpire:  time.Duration(configs.AUTH.AccessTokenTTL) * time.Minute,
		RefreshExpire: time.Duration(configs.AUTH.RefreshTokenTTL) * time.Minute,
		AccessSecret:  configs.AUTH.Key,
		Permissions:   model.Permissions,
		TTL:           time.Duration(configs.LIMITER.TTL) * time.Second,
		MaxAttempts:   configs.LIMITER.MaxAttempts,
	}, &profile2.Profile[profile3.Profile, profile3.ProfileCreateInput, profile3.ProfileUpdateInput, profile3.ProfileUpdateProfileInput]{
		Service:     profile3.NewService[profile3.Profile, profile3.ProfileCreateInput, profile3.ProfileUpdateInput](profileRepo),
		Request:     profile3.NewRequest[profile3.ProfileCreateInput, profile3.ProfileUpdateInput, profile3.ProfileUpdateProfileInput](),
		Transformer: profile3.NewTransformer[profile3.Profile]()})

	if err != nil {
		logger.Fatalf("[AUTH] %s", err.Error())
		return
	}

	analyticService := analytics.New(analytics.Config{
		DB:            db,
		FlushInterval: 10,
		BatchSize:     20,
		ExcludePaths:  make([]string, 0),
	})
	defer analyticService.ShutDown()

	ctx2 := mcontext.WithOperationIDContext(ctx, strconv.Itoa(int(time.Now().UTC().Unix())))
	if err := database3.SeedersRun(ctx, db, profileRepo, authService.GetAdminID()); err != nil {
		logger.Fatalf("[DB] %s", err.Error())
		return
	}

	repositories := repository.NewRepositories(db)
	services := service.NewServices(service.Deps{Repositories: repositories})
	sse := sse2.New(10)

	taskService, err := task.NewTask(ctx2, task.Config{
		DB:          db,
		Transaction: trx,
		Debug:       configs.LOG.Debug,
		Services: task.Consumers{
			"news": services.NewsService,
		},
		TokenString: "787hhjvYTYTcfcr6556tCTTYChgUYy",
		SseSvc:      sse,
	})
	if err != nil {
		logger.Fatalf("[TASK] %s", err.Error())
		return
	}

	restHandlers := restHandler.NewHandlers(authService, services, trx, taskService, analyticService, sse)
	restServer := api.NewServer(api.Input{
		Host: configs.HTTP.Host,
		Port: configs.HTTP.Port,
	}, restHandlers.Init(configs.LOG.Debug))

	go func() {
		logger.Infof("Starting Rest Server")
		log.Println("Starting Rest Server")
		if err := restServer.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			logger.Infof("Rest ListenAndServe err: %s\n", err.Error())
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)

	<-quit

	const timeout = 5 * time.Second
	ctx3, shutdown := context.WithTimeout(ctx2, timeout)
	defer shutdown()

	if err := restServer.Stop(ctx3); err != nil {
		logger.Infof("Failed to Stop Server: %v", err)
	}

	logger.Infof("Shutting Down Server")
	log.Println("Shutting Down Server")
}
