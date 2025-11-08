package middleware

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"io"
	"log"
	"mime"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/c2pc/go-pkg/v2/auth/internal/cache"
	"github.com/c2pc/go-pkg/v2/auth/internal/fx"
	"github.com/c2pc/go-pkg/v2/auth/internal/model"
	"github.com/c2pc/go-pkg/v2/auth/internal/repository"
	"github.com/c2pc/go-pkg/v2/utils/apperr"
	"github.com/c2pc/go-pkg/v2/utils/jsonutil"
	"github.com/c2pc/go-pkg/v2/utils/mcontext"
	response2 "github.com/c2pc/go-pkg/v2/utils/response/http"
	"github.com/c2pc/go-pkg/v2/utils/translator"
	"github.com/gin-gonic/gin"
)

type AnalyticConfig struct {
	HiddenKeys []string
}

type AnalyticMiddleware struct {
	cfg AnalyticConfig

	batchSize int

	adminEntries []model.Analytic

	mu         sync.Mutex
	cancelFunc context.CancelFunc

	auditorHolder *fx.AuditorHolder
	cache         *cache.Cache
	repositories  repository.Repositories
}

// NewAnalyticMiddleware создает новый экземпляр middleware для аналитики
func NewAnalyticMiddleware(cfg AnalyticConfig, auditorHolder *fx.AuditorHolder, cache *cache.Cache, repositories repository.Repositories) *AnalyticMiddleware {
	cfg.HiddenKeys = append(cfg.HiddenKeys, "pass", "token", "pwd", "code", "secret")

	l := &AnalyticMiddleware{
		cfg:           cfg,
		batchSize:     500,
		adminEntries:  make([]model.Analytic, 0, 500),
		auditorHolder: auditorHolder,
		cache:         cache,
		repositories:  repositories,
	}

	ctx, cancel := context.WithCancel(context.Background())
	l.cancelFunc = cancel

	go l.periodicFlush(ctx)

	return l
}

type responseWriter struct {
	gin.ResponseWriter
	body          *bytes.Buffer
	shouldCapture bool
}

func (rw *responseWriter) WriteHeader(statusCode int) {
	contentType := rw.Header().Get("Content-Type")
	mediaType, _, err := mime.ParseMediaType(contentType)
	if err == nil && mediaType == "application/json" {
		rw.shouldCapture = true
		rw.body = &bytes.Buffer{}
	}
	rw.ResponseWriter.WriteHeader(statusCode)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	if rw.shouldCapture {
		_, _ = rw.body.Write(b)
	}
	return rw.ResponseWriter.Write(b)
}

// Collect собирает данные аналитики о запросах
func (l *AnalyticMiddleware) Collect(c *gin.Context) {
	startTime := time.Now().UTC()
	realPath := c.Request.URL.Path
	method := c.Request.Method
	clientIP := c.ClientIP()

	var requestBody []byte
	if !strings.Contains(realPath, "sso") {
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

	w := &responseWriter{
		ResponseWriter: c.Writer,
	}
	c.Writer = w

	c.Next()

	status := c.Writer.Status()

	var userAction string
	action, ok := mcontext.GetOpAction(c.Request.Context())
	if ok && action != "" {
		userAction = action
	} else {
		return
	}

	var userID *int64
	id, ok := mcontext.GetOpUserID(c.Request.Context())
	if ok && id != 0 {
		userID = &id
	}

	var userData *model.User
	if userID != nil {
		usr, _ := l.cache.UserCache.GetUserInfo(c.Request.Context(), *userID, func(ctx context.Context) (*model.User, error) {
			return l.repositories.UserRepository.GetUserWithPermissions(ctx, "id = ?", *userID)
		})
		userData = usr
	}

	var roleName string
	if userData != nil {
		for _, r := range userData.Roles {
			if model.IsRole(r.Name, model.SuperAdmin) {
				roleName = r.Name
				break
			}
		}
	}

	if roleName == "" {
		roleName, _ = mcontext.GetOpUserRole(c.Request.Context())
	}

	if roleName != "" {
		c.Request = c.Request.WithContext(mcontext.WithOpUserRoleContext(c.Request.Context(), roleName))
	}

	operationID, _ := mcontext.GetOperationID(c.Request.Context())

	var request []byte
	var compressedRequest []byte
	if len(requestBody) > 0 {
		request = jsonutil.JsonHideImportantData(requestBody, l.cfg.HiddenKeys...)
		data := compressData(request)
		compressedRequest = data
	} else {
		compressedRequest = nil
	}

	errResponse := &response2.ErrorResponse{}

	var response []byte
	var compressedResponse []byte

	if w != nil && !strings.Contains(realPath, "sso") {
		if w.shouldCapture && w.body != nil && w.body.Len() > 0 {
			if w.Status() >= 400 {
				err := json.Unmarshal(w.body.Bytes(), errResponse)
				if err != nil {
					errResponse = nil
				}
			}
			response = jsonutil.JsonHideImportantData(w.body.Bytes(), l.cfg.HiddenKeys...)
			data := compressData(response)
			compressedResponse = data
		} else {
			compressedResponse = nil
		}
	} else {
		compressedResponse = nil
	}

	var errDetailRU *string
	appError, ok := mcontext.GetOpError(c.Request.Context())
	if ok {
		e := appError.Error()
		errDetailRU = &e
	} else {
		if errResponse != nil {
			errFromMap, err := apperr.ErrMapManager.Get(errResponse.ID)
			if err == nil {
				tr := apperr.Translate(errFromMap, string(translator.RU))
				errDetailRU = &tr
			}
		}
	}

	entry := model.Analytic{
		OperationID: operationID,
		Path:        realPath,
		UserID:      userID,
		Method:      method,
		StatusCode:  status,
		ClientIP:    clientIP,
		Error:       errDetailRU,
		CreatedAt:   startTime,
		Action:      &userAction,
	}

	if userData != nil {
		entry.Login = &userData.Login
		var name []string
		if userData.LastName != nil {
			name = append(name, *userData.LastName)
		}
		name = append(name, userData.FirstName)
		if userData.SecondName != nil {
			name = append(name, *userData.SecondName)
		}
		nm := strings.Join(name, " ")
		entry.Name = &nm
	}

	if strings.ToUpper(method) != http.MethodGet {
		entry.RequestBody = compressedRequest
		entry.ResponseBody = compressedResponse
	}

	l.addEntry(entry)
}

func compressData(data []byte) []byte {
	if len(data) == 0 {
		return nil
	}

	var buf bytes.Buffer
	gz, err := gzip.NewWriterLevel(&buf, gzip.BestSpeed)
	if err != nil {
		log.Printf("gzip init error: %v", err)
		return nil
	}

	if _, err := gz.Write(data); err != nil {
		log.Printf("gzip write error: %v", err)
		return nil
	}

	if err := gz.Close(); err != nil {
		log.Printf("gzip close error: %v", err)
		return nil
	}

	return buf.Bytes()
}
func (l *AnalyticMiddleware) addEntry(entry model.Analytic) {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.adminEntries = append(l.adminEntries, entry)
	if len(l.adminEntries) >= l.batchSize {
		l.flush()
	}
}

func (l *AnalyticMiddleware) periodicFlush(ctx context.Context) {
	ticker := time.NewTicker(10 * time.Second)
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			l.mu.Lock()
			if len(l.adminEntries) > 0 {
				l.flush()
			}
			l.mu.Unlock()
		}
	}
}

func (l *AnalyticMiddleware) flush() {
	var entries []model.Analytic
	tableName := model.Analytic{}.TableName()
	entries = l.adminEntries

	if len(entries) == 0 {
		return
	}

	ctx := mcontext.WithOperationIDContext(context.Background(), strconv.FormatInt(time.Now().UnixMilli(), 10))

	_, _ = l.repositories.AnalyticRepository.WithTable(tableName).Create2(ctx, &entries)

	entriesCapacity := cap(entries)
	l.adminEntries = make([]model.Analytic, 0, entriesCapacity)
}

func (l *AnalyticMiddleware) Shutdown(ctx context.Context) {
	l.cancelFunc()

	l.mu.Lock()
	defer l.mu.Unlock()

	if len(l.adminEntries) > 0 {
		_, _ = l.repositories.AnalyticRepository.WithTable(model.Analytic{}.TableName()).Create2(ctx, &l.adminEntries)
	}

	l.adminEntries = nil
}
