package auth

import (
	"context"
	"fmt"
	"math"
	"net/http"
	"time"

	"github.com/c2pc/go-pkg/v2/auth/configurator"
	"github.com/c2pc/go-pkg/v2/auth/internal/cache/cachekey"
	authConf "github.com/c2pc/go-pkg/v2/auth/internal/configurator"
	"github.com/c2pc/go-pkg/v2/auth/internal/database"
	"github.com/c2pc/go-pkg/v2/auth/internal/fx"
	"github.com/c2pc/go-pkg/v2/auth/internal/model"
	"github.com/c2pc/go-pkg/v2/auth/internal/repository"
	"github.com/c2pc/go-pkg/v2/auth/internal/service"
	"github.com/c2pc/go-pkg/v2/auth/internal/transport/api/handler"
	"github.com/c2pc/go-pkg/v2/auth/internal/transport/api/middleware"
	"github.com/c2pc/go-pkg/v2/auth/profile"
	"github.com/c2pc/go-pkg/v2/task"
	"github.com/c2pc/go-pkg/v2/utils/app_data"
	"github.com/c2pc/go-pkg/v2/utils/apperr"
	"github.com/c2pc/go-pkg/v2/utils/logger"
	"github.com/c2pc/go-pkg/v2/utils/meta"
	"github.com/c2pc/go-pkg/v2/utils/mw"
	response "github.com/c2pc/go-pkg/v2/utils/response/http"
	"github.com/c2pc/go-pkg/v2/utils/syslog"
	"github.com/c2pc/go-pkg/v2/websocket"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type Tasker interface {
	task.Handler
	InitConsumers(consumers task.Consumers)
}

type IAuth interface {
	//NewHandlerEngine Инициализация gin.Engine
	NewHandlerEngine() *gin.Engine
	//InitHandler Инициализация хендлеров
	InitHandler(engine *gin.Engine, api *gin.RouterGroup, handlers ...gin.HandlerFunc)
	//AuthenticateMW Проверка авторизации
	AuthenticateMW(c *gin.Context)
	//CanPermissionMW проверка прав
	CanPermissionMW(c *gin.Context)
	//GetAdminID Получение ID дефолтного админа
	GetAdminID() int64
	//SetConfig Установка новых конфигураций
	SetConfig(ctx context.Context, key string, value configurator.Configurator) (configurator.Config, error)
	//Start Запуск сервиса
	Start(ctx context.Context) error
	//Stop остановка сервиса
	Stop(ctx context.Context)
	//WebSocket websocket
	WebSocket() websocket.Sender
	//Task планировщик задач
	Task() Tasker
}

type Auth struct {
	handler              handler.IHandler
	adminID              int64
	permissionMiddleware *middleware.PermissionMiddleware
	authCfg              authConf.ACfg
	auditCfg             authConf.AuditCfg
	rdb                  redis.UniversalClient
	db                   *gorm.DB
	tokenMW              *middleware.TokenMiddleware
	analyticsMW          *middleware.AnalyticMiddleware
	configService        service.IConfigService

	cacheHolder   *fx.CacheHolder
	auditHolder   *fx.AuditorHolder
	ldapHolder    *fx.LDAPHolder
	oidcHolder    *fx.OIDCHolder
	samlHolder    *fx.SAMLHolder
	limiterHolder *fx.LimiterHolder
	authHolder    *fx.AuthHolder

	task       task.Tasker
	ws         websocket.WebSocket
	cancelFunc context.CancelFunc
}

type Analytic struct {
	//ExcludeInputBodies Исключать requests
	ExcludeInputBodies map[string][]string
	//ExcludeOutputBodies Исключать responses
	ExcludeOutputBodies map[string][]string
	//SkipRequests Исключать запросы
	SkipRequests map[string][]string
	//HiddenKeys Скрытые ключи
	HiddenKeys []string
}
type Input struct {
	DB          *gorm.DB
	Rdb         redis.UniversalClient
	Permissions []meta.Permission
	Analytic    Analytic
}

type Data struct {
	Vendor     string
	AppName    string
	AppVersion string
	LogPath    string
}

func New(
	ctx context.Context,
	data Data,
	input Input,
	prof *profile.Profile,
) (IAuth, error) {
	ctx, cancelFunc := context.WithCancel(ctx)

	if data.AppName == "" || data.AppVersion == "" || data.LogPath == "" || data.Vendor == "" {
		cancelFunc()
		return nil, fmt.Errorf("app name, app version, vendor and log path are required")
	} else {
		app_data.Vendor = data.Vendor
		app_data.AppName = data.AppName
		app_data.AppVersion = data.AppVersion
		if app_data.IsDevMode() {
			data.LogPath = "logs"
		}
	}

	resetMigration(input.DB)

	if err := database.Migrate(input.DB); err != nil {
		cancelFunc()
		return nil, err
	}

	cachekey.SetServiceName(data.AppName)
	model.SetPermissions(input.Permissions)

	repositories := repository.NewRepositories(input.DB)

	configService := service.NewConfigService(repositories)

	logConfig := authConf.NewLogConfigurator(data.LogPath)
	logCfg, err := getLogCfg(ctx, configService, logConfig)
	if err != nil {
		cancelFunc()
		return nil, err
	}
	resetLog(logCfg)

	authConfig := authConf.NewAuthConfigurator()
	authCfg, err := getAuthCfg(ctx, configService, authConfig)
	if err != nil {
		cancelFunc()
		return nil, err
	}

	analyticConfig := authConf.NewAuditConfigurator(data.LogPath)
	auditCfg, err := getAuditCfg(ctx, configService, analyticConfig)
	if err != nil {
		cancelFunc()
		return nil, err
	}
	resetSyslog(auditCfg)
	PrintAppStartMessage(ctx)

	admin, err := database.SeedersRun(ctx, input.DB, repositories, model.GetPermissionsKeys())
	if err != nil {
		cancelFunc()
		return nil, err
	}

	var profileService profile.IProfileService
	var profileTransformer profile.ITransformer
	var profileRequest profile.IRequest
	if prof != nil {
		profileService = prof.Service
		profileTransformer = prof.Transformer
		profileRequest = prof.Request
	}

	accessTokenTTL := time.Duration(authCfg.AccessTokenTTL) * time.Minute
	refreshExpire := time.Duration(authCfg.RefreshTokenTTL) * time.Minute
	accessSecret := authCfg.AccessKey

	cacheHolder := fx.NewCacheHolder(input.Rdb, accessTokenTTL)
	auditHolder := fx.NewAuditorHolder(fx.Auditor{
		Admin: fx.AuditorDB{
			MaxAgeDays:         auditCfg.DB.Admin.MaxAgeDays,
			MaxCount:           auditCfg.DB.Admin.MaxCount,
			PercentageOfDelete: auditCfg.DB.Admin.PercentageOfDelete,
		},
	})
	ldapHolder := fx.NewLDAPHolder(len(authCfg.LDAP) != 0, getLdapConfig(authCfg.LDAP))
	oidcHolder, err := fx.NewOIDCHolder(ctx, getOIDCConfig(authCfg.SSO.Enabled == "oidc", authCfg.SSO.OIDC, authCfg.SSO.Description))
	if err != nil {
		cancelFunc()
		return nil, err
	}
	samlHolder, err := fx.NewSAMLHolder(ctx, getSAMLConfig(authCfg.SSO.Enabled == "saml", authCfg.SSO.SAML, authCfg.SSO.Description))
	if err != nil {
		cancelFunc()
		return nil, err
	}

	limiterHolder := fx.NewLimiterHolder(getLimiterConfig(authCfg.Limiter))
	authHolder := fx.NewAuthHolder(accessTokenTTL, refreshExpire, string(accessSecret))
	tokenMW := middleware.NewTokenMiddleware(cacheHolder, repositories, authHolder)

	authService := service.NewAuthService(profileService, repositories, cacheHolder, authHolder, ldapHolder, oidcHolder, samlHolder, limiterHolder)
	permissionService := service.NewPermissionService(repositories, cacheHolder)
	roleService := service.NewRoleService(repositories, cacheHolder)
	userService := service.NewUserService(profileService, repositories, cacheHolder)
	settingService := service.NewSettingService(repositories)
	sessionService := service.NewSessionService(repositories, cacheHolder)
	userBlockedService := service.NewUserBlockedService(repositories, cacheHolder)
	filterService := service.NewFilterService(repositories)
	versionService := service.NewVersionService(data.AppVersion, repositories)
	analyticService := service.NewAnalyticService(repositories)
	permissionMiddleware := middleware.NewPermissionMiddleware(cacheHolder.Get(), repositories)

	if input.Analytic.SkipRequests == nil {
		input.Analytic.SkipRequests = make(map[string][]string)
	}
	input.Analytic.SkipRequests["/auth/settings"] = []string{}
	input.Analytic.SkipRequests["/stream"] = []string{}
	input.Analytic.SkipRequests["/version"] = []string{}
	input.Analytic.SkipRequests["/ping"] = []string{}
	analyticMiddleware := middleware.NewAnalyticMiddleware(middleware.AnalyticConfig{}, auditHolder, cacheHolder.Get(), repositories)

	trx := mw.NewTransaction(input.DB)
	ws := websocket.New(100, 10)
	tasker, err := task.NewTask(ctx, task.Config{
		DB:          input.DB,
		Transaction: trx,
		WS:          ws,
	})

	handlers := handler.NewHandlers(
		authService,
		permissionService,
		roleService,
		userService,
		settingService,
		filterService,
		sessionService,
		configService,
		analyticService,
		userBlockedService,
		trx,
		tokenMW,
		permissionMiddleware,
		analyticMiddleware,
		profileTransformer,
		profileRequest,
		oidcHolder,
		samlHolder,
		versionService,
		limiterHolder,
		ws,
		tasker,
	)

	auth := &Auth{
		handler:              handlers,
		adminID:              admin.ID,
		cacheHolder:          cacheHolder,
		ldapHolder:           ldapHolder,
		oidcHolder:           oidcHolder,
		samlHolder:           samlHolder,
		auditHolder:          auditHolder,
		tokenMW:              tokenMW,
		analyticsMW:          analyticMiddleware,
		authHolder:           authHolder,
		limiterHolder:        limiterHolder,
		permissionMiddleware: permissionMiddleware,
		configService:        configService,
		authCfg:              authCfg,
		auditCfg:             auditCfg,
		rdb:                  input.Rdb,
		db:                   input.DB,
		cancelFunc:           cancelFunc,
		task:                 tasker,
		ws:                   ws,
	}

	go auth.startSessionCleaner(ctx)
	go auth.startUserBlockedCleaner(ctx)
	go auth.startAnalyticCleaner(ctx)
	go auth.authConfigWatcher(ctx, authConfig)
	go auth.auditConfigWatcher(ctx, analyticConfig)
	go auth.logConfigWatcher(ctx, logConfig)

	return auth, nil
}

func (a *Auth) InitHandler(engine *gin.Engine, api *gin.RouterGroup, handlers ...gin.HandlerFunc) {
	a.handler.Init(engine, api, handlers...)
}

func (a *Auth) NewHandlerEngine() *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()

	gin.DebugPrintRouteFunc = func(httpMethod, absolutePath, handlerName string, nuHandlers int) {}

	engine.Use(mw.HandlersFunc...)
	engine.Use(a.analyticsMW.Collect)

	engine.NoRoute(func(c *gin.Context) {
		response.Response(c, apperr.ErrNotFound)
	})

	engine.POST("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})

	engine.Any("/", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	return engine
}

func (a *Auth) AuthenticateMW(c *gin.Context) { a.tokenMW.Authenticate(c) }

func (a *Auth) CanPermissionMW(c *gin.Context) { a.permissionMiddleware.Can(c) }

func (a *Auth) GetAdminID() int64 { return a.adminID }

func (a *Auth) startSessionCleaner(ctx context.Context) {
	tm := time.NewTicker(10 * time.Minute)
	defer tm.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-tm.C:
			a.db.Exec(`DELETE FROM auth_tokens WHERE expires_at < ?`,
				time.Now().UTC())
		}
	}
}

func (a *Auth) startUserBlockedCleaner(ctx context.Context) {
	tm := time.NewTicker(1 * time.Minute)
	defer tm.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-tm.C:
			a.db.Exec(`DELETE FROM auth_users_blocked WHERE blocked_at <= ?`,
				time.Now().UTC().Add(-1*a.limiterHolder.Get().BlockingTTL))
		}
	}
}

func (a *Auth) startAnalyticCleaner(ctx context.Context) {
	fn := func() {
		var deleteSQL string
		if a.db.Dialector.Name() == "mysql" {
			deleteSQL = `DELETE t FROM %s t
								 JOIN (
							SELECT id
							FROM %s
							ORDER BY created_at ASC
							LIMIT 1
						) sub ON t.id = sub.id`
		} else if a.db.Dialector.Name() == "postgres" {
			deleteSQL = `DELETE FROM %s
							WHERE id IN (
								SELECT id
								FROM %s
								ORDER BY created_at ASC
								LIMIT ?
						);`
		}

		err := a.db.Exec(`DELETE FROM auth_analytics_admins WHERE created_at < ?`,
			time.Now().UTC().Add(-1*time.Duration(a.auditHolder.Get().Admin.MaxAgeDays)*24*time.Hour)).Error
		if err != nil {
			//fmt.Println(err)
		}

		if deleteSQL == "" {
			return
		}

		var totalCount int64
		err = a.db.Raw("SELECT COUNT(*) FROM auth_analytics_admins").Scan(&totalCount).Error
		if err == nil && totalCount > int64(a.auditHolder.Get().Admin.MaxCount) {
			//count = 1500 - 1000 + (1000 * 20 / 100)
			//count = 500 + 200
			//count = 700
			count := totalCount - int64(a.auditHolder.Get().Admin.MaxCount) +
				int64(math.Ceil(float64(a.auditHolder.Get().Admin.MaxCount)*
					a.auditHolder.Get().Admin.PercentageOfDelete/100))

			if count > 0 {
				err = a.db.Exec(fmt.Sprintf(deleteSQL, "auth_analytics_admins", "auth_analytics_admins"), count).Error
				if err != nil {
					//fmt.Println(err)
				}
			}
		} else if err != nil {
			//fmt.Println(err)
		}
	}
	tm := time.NewTicker(10 * time.Minute)
	defer tm.Stop()
	fn()

	for {
		select {
		case <-ctx.Done():
			return
		case <-tm.C:
			fn()
		}
	}
}

func (a *Auth) SetConfig(ctx context.Context, key string, value configurator.Configurator) (configurator.Config, error) {
	var v configurator.Config
	err := a.db.Transaction(func(tx *gorm.DB) error {
		val, err := a.configService.Trx(tx).SetConfig(ctx, key, value)
		if err != nil {
			return err
		}
		v = val
		return nil
	})
	return v, err
}

func (a *Auth) WebSocket() websocket.Sender {
	return a.ws
}

func (a *Auth) Task() Tasker {
	return a.task
}

func (a *Auth) Start(ctx context.Context) error {
	return a.db.Transaction(func(tx *gorm.DB) error {
		return a.configService.Trx(tx).Init(ctx)
	})
}

func (a *Auth) Stop(ctx context.Context) {
	if a.cancelFunc != nil {
		a.cancelFunc()
	}

	a.analyticsMW.Shutdown(ctx)
	PrintAppStopMessage(ctx)
}

func resetMigration(db *gorm.DB) {
	sqlDB, err := db.DB()
	if err != nil {
		return
	}

	_, schemaMigrationsTableCheck := sqlDB.Query("select * from schema_migrations;")
	if schemaMigrationsTableCheck == nil {
		db.Exec(`DELETE FROM schema_migrations WHERE version < 1 OR (version = 1 AND dirty = true)`)
		db.Exec(`UPDATE schema_migrations SET version = version - 1, dirty = false WHERE dirty = true`)
	}

	_, schemaAuthMigrationsTableCheck := sqlDB.Query("select * from schema_auth_migrations;")
	if schemaAuthMigrationsTableCheck == nil {
		db.Exec(`DELETE FROM schema_auth_migrations WHERE version < 1 OR (version = 1 AND dirty = true)`)
		db.Exec(`UPDATE schema_auth_migrations SET version = version - 1, dirty = false WHERE dirty = true`)
	}
}

func PrintAppStartMessage(ctx context.Context) {
	logger.AppInfoFLog(ctx, "Начало работы системы")
	syslog.Write(ctx, syslog.Record{
		EventID:   "starting-app",
		EventName: "Запуск системы",
		Severity:  syslog.SeverityLow,
		Success:   true,
	}, "Начало работы системы")
}

func PrintAppStopMessage(ctx context.Context) {
	logger.AppInfoFLog(ctx, "Окончание работы системы")
	syslog.Write(ctx, syslog.Record{
		EventID:   "stopping-app",
		EventName: "Остановка системы",
		Severity:  syslog.SeverityHigh,
		Success:   true,
	}, "Окончание работы системы")
}
