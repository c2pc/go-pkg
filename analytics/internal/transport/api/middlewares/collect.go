package collector

import (
	"bytes"
	"compress/gzip"
	"context"
	"io"
	"log"
	"mime"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/c2pc/go-pkg/v2/analytics/internal/models"
	"github.com/c2pc/go-pkg/v2/utils/mcontext"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type LoggerConfig struct {
	DB            *gorm.DB
	FlushInterval int
	BatchSize     int
	ExcludePaths  []string
}

type logger struct {
	db            *gorm.DB
	batchSize     int
	entries       []models.Analytics
	userIDMap     map[int]struct{}
	mu            sync.Mutex
	flushInterval time.Duration
	ticker        *time.Ticker
	cancelFunc    context.CancelFunc
	excludePaths  map[string]struct{}
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

func New(cfg LoggerConfig) (gin.HandlerFunc, func()) {
	if cfg.FlushInterval <= 10 {
		cfg.FlushInterval = 10
	}
	if cfg.BatchSize == 0 {
		cfg.BatchSize = 100
	}

	excludeMap := make(map[string]struct{})
	for _, p := range cfg.ExcludePaths {
		excludeMap[p] = struct{}{}
	}

	l := &logger{
		db:            cfg.DB,
		batchSize:     cfg.BatchSize,
		entries:       make([]models.Analytics, 0, cfg.BatchSize),
		userIDMap:     make(map[int]struct{}),
		flushInterval: time.Duration(cfg.FlushInterval) * time.Second,
		excludePaths:  excludeMap,
	}

	ctx, cancel := context.WithCancel(context.Background())
	l.cancelFunc = cancel
	l.ticker = time.NewTicker(l.flushInterval)

	go l.periodicFlush(ctx)

	return l.middleware, l.shutdown
}

func (l *logger) middleware(c *gin.Context) {
	startTime := time.Now()

	path := c.FullPath()
	re := regexp.MustCompile(`^/api/v\d+`)
	cleanedPath := re.ReplaceAllString(path, "")
	skipBodies := false
	if _, excluded := l.excludePaths[cleanedPath]; excluded {
		skipBodies = true
	}

	var requestBody []byte
	if !skipBodies {
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
	if !skipBodies {
		w = &responseWriter{
			ResponseWriter: c.Writer,
		}
		c.Writer = w
	}

	c.Next()

	duration := time.Since(startTime).Milliseconds()

	ctx := c.Request.Context()

	realPath := c.Request.URL.Path
	method := c.Request.Method
	status := c.Writer.Status()
	clientIP := c.ClientIP()

	var userID *int
	id, ok := mcontext.GetOpUserID(ctx)
	if ok {
		userID = &id
	}

	operationID, _ := mcontext.GetOperationID(ctx)

	var compressedRequest []byte
	if len(requestBody) > 0 {
		data := compressData(requestBody)
		compressedRequest = data
	} else {
		compressedRequest = nil
	}

	var compressedResponse []byte
	if w.flag && w.body != nil && w.body.Len() > 0 {
		data := compressData(w.body.Bytes())
		compressedResponse = data
	} else {
		compressedResponse = nil
	}

	entry := models.Analytics{
		OperationID:  operationID,
		Path:         realPath,
		UserID:       userID,
		Method:       method,
		StatusCode:   status,
		ClientIP:     clientIP,
		RequestBody:  compressedRequest,
		ResponseBody: compressedResponse,
		Duration:     duration,
	}

	l.addEntry(entry)
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

func (l *logger) addEntry(entry models.Analytics) {
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

	err := l.db.Create(&l.entries).Error
	if err != nil {
		log.Printf("error when inserting analytics: %v", err)
	}
	l.entries = l.entries[:0]
	l.userIDMap = make(map[int]struct{})
}

func (l *logger) analyticWithUserData() error {
	userIDs := make([]int, 0, len(l.userIDMap))
	for id := range l.userIDMap {
		userIDs = append(userIDs, id)
	}

	var users []models.User
	if len(userIDs) > 0 {
		if err := l.db.Where("id IN ?", userIDs).Find(&users).Error; err != nil {
			return err
		}
	}

	userMap := make(map[int]models.User)
	for _, user := range users {
		userMap[user.ID] = user
	}

	for i := range l.entries {
		if l.entries[i].UserID != nil {
			if user, exists := userMap[*l.entries[i].UserID]; exists {
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

		if err := l.db.Create(&l.entries).Error; err != nil {
			log.Printf("error when inserting analytics: %v", err)
		}
		l.entries = l.entries[:0]
		l.userIDMap = make(map[int]struct{})
		l.excludePaths = make(map[string]struct{})
	}
}
