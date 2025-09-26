package collector

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/c2pc/go-pkg/v2/analytics/internal/cache"
	"github.com/c2pc/go-pkg/v2/analytics/internal/fx"
	"github.com/c2pc/go-pkg/v2/utils/apperr"
	"github.com/c2pc/go-pkg/v2/utils/constant"
	loggerServ "github.com/c2pc/go-pkg/v2/utils/logger"
	"github.com/c2pc/go-pkg/v2/utils/translator"
	"github.com/rs/zerolog"
	"gorm.io/gorm"

	"github.com/c2pc/go-pkg/v2/analytics/internal/model"
	"github.com/c2pc/go-pkg/v2/ffm"
	"github.com/c2pc/go-pkg/v2/utils/jsonutil"
	"github.com/c2pc/go-pkg/v2/utils/mcontext"
	"github.com/gin-gonic/gin"
)

type LoggerConfig struct {
	DB                  *gorm.DB
	FlushInterval       int
	BatchSize           int
	ExcludeInputBodies  map[string][]string
	ExcludeOutputBodies map[string][]string
	SkipRequests        map[string][]string
	HiddenKeys          []string
}

type logger struct {
	db                  *gorm.DB
	batchSize           int
	entries             []model.Analytics
	userIDMap           map[int]struct{}
	mu                  sync.Mutex
	flushInterval       time.Duration
	ticker              *time.Ticker
	cancelFunc          context.CancelFunc
	ExcludeInputBodies  map[string][]string
	ExcludeOutputBodies map[string][]string
	SkipRequests        map[string][]string
	HiddenKeys          []string
	configHolder        *fx.ConfigHolder
	cache               *cache.Cache
}

type responseWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
	flag bool
}

func (rw *responseWriter) WriteHeader(statusCode int) {
	contentType := rw.Header().Get("Content-Type")
	mediaType, _, err := mime.ParseMediaType(contentType)
	if err == nil && mediaType == "application/json" {
		rw.flag = true
		rw.body = &bytes.Buffer{}
	}
	rw.ResponseWriter.WriteHeader(statusCode)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	if rw.flag {
		rw.body.Write(b)
	}
	return rw.ResponseWriter.Write(b)
}

func New(cfg LoggerConfig, configHolder *fx.ConfigHolder, cache *cache.Cache) (gin.HandlerFunc, func()) {
	if cfg.FlushInterval <= 10 {
		cfg.FlushInterval = 10
	}
	if cfg.BatchSize == 0 {
		cfg.BatchSize = 100
	}

	cfg.HiddenKeys = append(cfg.HiddenKeys, "pass", "token", "pwd", "code", "secret")

	l := &logger{
		db:                  cfg.DB,
		batchSize:           cfg.BatchSize,
		entries:             make([]model.Analytics, 0, cfg.BatchSize),
		userIDMap:           make(map[int]struct{}),
		flushInterval:       time.Duration(cfg.FlushInterval) * time.Second,
		ExcludeInputBodies:  cfg.ExcludeInputBodies,
		ExcludeOutputBodies: cfg.ExcludeOutputBodies,
		SkipRequests:        cfg.SkipRequests,
		HiddenKeys:          cfg.HiddenKeys,
		configHolder:        configHolder,
		cache:               cache,
	}

	ctx, cancel := context.WithCancel(context.Background())
	l.cancelFunc = cancel
	l.ticker = time.NewTicker(l.flushInterval)

	go l.periodicFlush(ctx)

	return l.middleware, l.shutdown
}

func (l *logger) check(path string, method string) (bool, bool, bool) {
	var skipReq, skipInputBody, skipOutputBody = false, false, false

	if methods, ok := l.SkipRequests[path]; ok {
		if len(methods) == 0 {
			skipReq = true
		}
		for _, m := range methods {
			if m == method {
				skipReq = true
			}
		}
	}

	if methods, ok := l.ExcludeInputBodies[path]; ok {
		if len(methods) == 0 {
			skipInputBody = true
		}
		for _, m := range methods {
			if m == method {
				skipInputBody = true
			}
		}
	}

	if methods, ok := l.ExcludeOutputBodies[path]; ok {
		if len(methods) == 0 {
			skipOutputBody = true
		}
		for _, m := range methods {
			if m == method {
				skipOutputBody = true
			}
		}
	}

	return skipReq, skipInputBody, skipOutputBody
}

func (l *logger) middleware(c *gin.Context) {
	path := c.FullPath()
	re := regexp.MustCompile(`^/api/v\d+`)
	cleanedPath := re.ReplaceAllString(path, "")

	skipReq, skipInputBody, skipOutputBody := l.check(cleanedPath, c.Request.Method)

	var logDisabled bool
	if c.Request.Context().Value("log_disabled") != nil {
		logDisabled = c.Request.Context().Value("log_disabled").(bool)
	}

	if !skipReq && !logDisabled {
		startTime := time.Now()

		realPath := c.Request.URL.Path
		method := c.Request.Method
		clientIP := c.ClientIP()

		var requestBody []byte
		if !skipInputBody && !strings.Contains(realPath, "sso") {
			if c.Request.Method == http.MethodGet {
				query := c.Request.URL.Query()
				if query.Encode() != "" {
					q, _ := url.QueryUnescape(query.Encode())
					requestBody = []byte(q)
				}
			} else if c.Request.Body != nil && c.Request.Body != http.NoBody && strings.Contains(c.Request.Header.Get("Content-Type"), "application/json") {
				var err error
				requestBody, err = io.ReadAll(c.Request.Body)
				if err == nil {
					c.Request.Body = io.NopCloser(bytes.NewBuffer(requestBody))
				} else {
					requestBody = nil
				}
			}
		}

		var w *responseWriter
		if !skipOutputBody {
			w = &responseWriter{
				ResponseWriter: c.Writer,
			}
			c.Writer = w
		}

		c.Next()

		duration := time.Since(startTime).Milliseconds()
		status := c.Writer.Status()

		ctx := c.Request.Context()

		var userID *int
		id, ok := mcontext.GetOpUserID(ctx)
		if ok && id != 0 {
			userID = &id
		}

		var userData *model.User
		if userID != nil {
			usr, _ := l.cache.UserCache.GetUserInfo(c.Request.Context(), *userID, func(ctx context.Context) (*model.User, error) {
				var usr model.User
				err := l.db.
					WithContext(ctx).
					Where("id = ?", *userID).
					Find(&usr).Error
				if err != nil {
					return nil, err
				}

				return &usr, nil
			})
			userData = usr
		}

		operationID, _ := mcontext.GetOperationID(ctx)

		var request []byte
		var compressedRequest []byte
		if len(requestBody) > 0 {
			request = jsonutil.JsonHideImportantData(requestBody, l.HiddenKeys...)
			data := compressData(request)
			compressedRequest = data
		} else {
			compressedRequest = nil
		}

		errResponse := &ffm.ErrorResponse{}

		var response []byte
		var compressedResponse []byte
		if w != nil && !strings.Contains(realPath, "sso") {
			if w.flag && w.body != nil && w.body.Len() > 0 {
				if w.Status() >= 400 {
					err := json.Unmarshal(w.body.Bytes(), errResponse)
					if err != nil {
						errResponse = nil
					}
				}
				response = jsonutil.JsonHideImportantData(w.body.Bytes(), l.HiddenKeys...)
				data := compressData(response)
				compressedResponse = data
			} else {
				compressedResponse = nil
			}
		} else {
			compressedResponse = nil
		}

		var userAction *string
		action, ok := mcontext.GetOpAction(ctx)
		if ok && action != "" {
			userAction = &action
		}

		var errDetailRU *string
		if errResponse != nil {
			errFromMap, err := apperr.ErrMapManager.Get(errResponse.ID)
			if err == nil {
				tr := apperr.Translate(errFromMap, string(translator.RU))
				errDetailRU = &tr
			}
		}

		entry := model.Analytics{
			OperationID: operationID,
			Path:        realPath,
			UserID:      userID,
			Method:      method,
			StatusCode:  status,
			ClientIP:    clientIP,
			Duration:    duration,
			Error:       errDetailRU,
			CreatedAt:   time.Now().UTC(),
		}

		if userData != nil {
			entry.Login = &userData.Login
			entry.FirstName = userData.FirstName
			entry.LastName = userData.LastName
		}

		if loggerServ.GetLevel() == zerolog.DebugLevel || strings.ToUpper(method) != http.MethodGet || userAction != nil {
			entry.RequestBody = compressedRequest
			entry.ResponseBody = compressedResponse
		}

		var level zerolog.Level
		logPkg := loggerServ.Info()
		if status >= 400 {
			if userAction != nil {
				logPkg = loggerServ.NoLevel()
			} else {
				logPkg = loggerServ.Error()
			}
			level = zerolog.ErrorLevel
		} else {
			if userAction != nil {
				logPkg = loggerServ.NoLevel()
				level = zerolog.InfoLevel
			} else {
				if strings.ToUpper(method) == http.MethodGet {
					logPkg = loggerServ.Debug()
					level = zerolog.DebugLevel
				} else {
					logPkg = loggerServ.Info()
					level = zerolog.InfoLevel
				}
			}
		}

		var act string
		if userAction != nil {
			act = *userAction
		} else {
			switch strings.ToUpper(method) {
			case http.MethodGet:
				act = "Чтение объекта"
			case http.MethodPost:
				if strings.Contains(realPath, "mass-update") {
					act = "Массовое изменение объектов"
				} else if strings.Contains(realPath, "update") {
					act = "Изменение объекта"
				} else if strings.Contains(realPath, "mass-delete") {
					act = "Массовое удаление объектов"
				} else if strings.Contains(realPath, "import") {
					act = "Импорт объектов"
				} else if strings.Contains(realPath, "export") {
					act = "Экспорт объектов"
				} else if strings.Contains(realPath, "delete") {
					act = "Удаление объекта"
				} else {
					act = "Создание объекта"
				}
			case http.MethodPut, http.MethodPatch:
				act = "Изменение объекта"
			case http.MethodDelete:
				act = "Удаление объекта"
			}
		}

		if act != "" {
			entry.Action = &act
		}

		logPkg = logPkg.
			Str(string(constant.OperationID), operationID).
			Str("ip", clientIP).
			Str("duration", fmt.Sprintf("%dms", duration))

		if userID != nil {
			logPkg = logPkg.Int("action_user_id", *userID)
		}

		if userData != nil {
			logPkg = logPkg.Str("action_user_login", userData.Login)
		}

		if act != "" {
			logPkg = logPkg.Str(string(constant.OpAction), "| "+act)
		}

		if errDetailRU != nil {
			logPkg = logPkg.Str("error", *errDetailRU)
		}

		if len(c.Errors.Errors()) > 0 {
			logPkg = logPkg.Str("errors", strings.Join(c.Errors.Errors(), ";"))
		}

		if loggerServ.GetLevel() == zerolog.DebugLevel || userAction != nil {
			logPkg = logPkg.Str("request", string(request))
			logPkg = logPkg.Str("response", string(response))
		}

		logPkg.Msgf("%d | %s %s", status, method, realPath)

		if l.configHolder.Get().Enabled {
			if userAction != nil || level >= l.configHolder.Get().Level {
				l.addEntry(entry)
			}
		}
	}
}

func compressData(data []byte) []byte {
	if len(data) == 0 {
		return nil
	}
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	_, err := gz.Write(data)
	if err != nil {
		log.Printf("gzip write error: %v", err)
		return nil
	}
	err = gz.Close()
	if err != nil {
		log.Printf("gzip close error: %v", err)
		return nil
	}
	return buf.Bytes()
}

func (l *logger) addEntry(entry model.Analytics) {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.entries = append(l.entries, entry)

	if entry.UserID != nil {
		l.userIDMap[*entry.UserID] = struct{}{}
	}

	if len(l.entries) >= l.batchSize {
		l.flush()
	}
}

func (l *logger) periodicFlush(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-l.ticker.C:
			l.mu.Lock()
			if len(l.entries) > 0 {
				l.flush()
			}
			l.mu.Unlock()
		}
	}
}

func (l *logger) flush() {
	if len(l.entries) == 0 {
		return
	}

	if err := l.analyticWithUserData(); err != nil {
		log.Printf("error when adding user data to the request")
	}

	if len(l.entries) > 0 {
		_ = l.db.
			WithContext(mcontext.WithOperationIDContext(context.Background(), strconv.FormatInt(time.Now().UnixMilli(), 10))).
			Create(&l.entries).Error
	}

	l.entries = l.entries[:0]
	l.userIDMap = make(map[int]struct{})
}

func (l *logger) analyticWithUserData() error {
	userIDs := make([]int, 0, len(l.userIDMap))
	for id := range l.userIDMap {
		userIDs = append(userIDs, id)
	}

	var users []model.User
	if len(userIDs) > 0 {
		if err := l.db.
			WithContext(mcontext.WithOperationIDContext(context.Background(), strconv.FormatInt(time.Now().UnixMilli(), 10))).
			Where("id IN ?", userIDs).Find(&users).Error; err != nil {
			return err
		}
	}

	userMap := make(map[int]model.User)
	for _, user := range users {
		userMap[user.ID] = user
	}

	for i := range l.entries {
		if l.entries[i].UserID != nil {
			if user, exists := userMap[*l.entries[i].UserID]; exists {
				l.entries[i].Login = &user.Login
				l.entries[i].FirstName = user.FirstName
				l.entries[i].SecondName = user.SecondName
				l.entries[i].LastName = user.LastName
			}
		}
	}

	return nil
}

func (l *logger) shutdown() {
	l.cancelFunc()
	l.ticker.Stop()
	l.mu.Lock()
	defer l.mu.Unlock()
	if len(l.entries) > 0 {
		if err := l.analyticWithUserData(); err != nil {
			log.Printf("error when adding user data to the request")
		}

		_ = l.db.
			WithContext(mcontext.WithOperationIDContext(context.Background(), strconv.FormatInt(time.Now().UnixMilli(), 10))).
			Create(&l.entries)
		l.entries = l.entries[:0]
		l.userIDMap = make(map[int]struct{})
	}
}
