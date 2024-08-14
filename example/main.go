package main

import (
	"context"
	"errors"
	"github.com/c2pc/go-pkg/v2/auth"
	"github.com/c2pc/go-pkg/v2/auth/profile"
	config2 "github.com/c2pc/go-pkg/v2/example/config"
	database2 "github.com/c2pc/go-pkg/v2/example/database"
	"github.com/c2pc/go-pkg/v2/example/model"
	"github.com/c2pc/go-pkg/v2/example/repository"
	"github.com/c2pc/go-pkg/v2/example/service"
	"github.com/c2pc/go-pkg/v2/example/transport/rest"
	restHandler "github.com/c2pc/go-pkg/v2/example/transport/rest/handler"
	"github.com/c2pc/go-pkg/v2/example/transport/rest/request"
	"github.com/c2pc/go-pkg/v2/example/transport/rest/transformer"
	"github.com/c2pc/go-pkg/v2/utils/cache/redis"
	database "github.com/c2pc/go-pkg/v2/utils/db"
	logger2 "github.com/c2pc/go-pkg/v2/utils/logger"
	"github.com/c2pc/go-pkg/v2/utils/mcontext"
	"github.com/c2pc/go-pkg/v2/utils/mw"
	"github.com/c2pc/go-pkg/v2/utils/secret"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"
)

func main() {
	if err := config2.Migrate("config.yml"); err != nil {
		log.Fatal("[Migrate] ", err)
		return
	}

	configs, err := config2.NewConfig("config.yml")
	if err != nil {
		log.Fatal("[CONFIG]", err)
		return
	}

	logger2.Initialize(false, "app.log", configs.APP.LogDir)

	db, err := database.ConnectPostgres(configs.PostgresUrl, configs.APP.Debug)
	if err != nil {
		logger2.Fatalf("[DB] %s", err.Error())
		return
	}

	rdb, err := redis.NewRedisClient(&redis.RedisClient{
		ClusterMode: configs.Redis.ClusterMode,
		Address:     configs.Redis.Address,
		Username:    configs.Redis.Username,
		Password:    configs.Redis.Password,
		MaxRetry:    configs.Redis.MaxRetry,
		DB:          configs.Redis.DB,
	}, configs.APP.Debug)
	if err != nil {
		logger2.Fatalf("[REDIS] %s", err.Error())
		return
	}

	hasher, err := secret.New(configs.PasswordSalt)
	if err != nil {
		logger2.Fatalf("[SECRET] %s", err.Error())
		return
	}

	trx := mw.NewTransaction(db)

	profileRepo := repository.NewProfileRepository(db)

	authService, err := auth.New[model.Profile, service.ProfileCreateInput, service.ProfileUpdateInput, service.ProfileUpdateProfileInput](auth.Config{
		Debug:         configs.APP.Debug,
		DB:            db,
		Rdb:           rdb,
		Transaction:   trx,
		Hasher:        hasher,
		AccessExpire:  time.Duration(configs.AUTH.AccessTokenTTL) * time.Minute,
		RefreshExpire: time.Duration(configs.AUTH.RefreshTokenTTL) * time.Minute,
		AccessSecret:  configs.AUTH.Key,
		Permissions:   model.Permissions,
	}, &profile.Profile[model.Profile, service.ProfileCreateInput, service.ProfileUpdateInput, service.ProfileUpdateProfileInput]{
		Service:     service.NewProfileService[model.Profile, service.ProfileCreateInput, service.ProfileUpdateInput](profileRepo),
		Request:     request.NewProfileRequest[service.ProfileCreateInput, service.ProfileUpdateInput, service.ProfileUpdateProfileInput](),
		Transformer: transformer.NewProfileTransformer[model.Profile](),
	})
	if err != nil {
		logger2.Fatalf("[AUTH] %s", err.Error())
		return
	}

	ctx := mcontext.WithOperationIDContext(context.Background(), strconv.Itoa(int(time.Now().UTC().Unix())))
	if err := database2.SeedersRun(ctx, db, profileRepo, authService.GetAdminID()); err != nil {
		logger2.Fatalf("[DB] %s", err.Error())
		return
	}

	restHandlers := restHandler.NewHandlers(authService)
	restServer := rest.NewServer(rest.Input{
		Host: configs.HTTP.Host,
		Port: configs.HTTP.Port,
	}, restHandlers.Init(configs.APP.Debug))

	go func() {
		logger2.Infof("Starting Rest Server")
		if err := restServer.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			logger2.Infof("Rest ListenAndServe err: %s\n", err.Error())
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)

	<-quit

	const timeout = 5 * time.Second
	ctx, shutdown := context.WithTimeout(context.Background(), timeout)
	defer shutdown()

	if err := restServer.Stop(ctx); err != nil {
		logger2.Infof("Failed to Stop Server: %v", err)
	}

	logger2.Infof("Shutting Down Server")
}
