package collector

import (
	"bytes"
	"compress/gzip"
	"context"
	"io"
	"log"
	"sync"
	"time"

	"github.com/c2pc/go-pkg/v2/analytics/internal/models"
	"github.com/c2pc/go-pkg/v2/utils/mcontext"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type LoggerConfig struct {
	DB            *gorm.DB
	FlushInterval time.Duration
	BatchSize     int
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
}

type responseWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	n, err := rw.ResponseWriter.Write(b)
	if err == nil {
		rw.body.Write(b)
	}
	return n, err
}

func New(cfg LoggerConfig) (gin.HandlerFunc, func()) {
	if cfg.FlushInterval == 0 {
		cfg.FlushInterval = 10 * time.Second
	}
	if cfg.BatchSize == 0 {
		cfg.BatchSize = 100
	}

	l := &logger{
		db:            cfg.DB,
		batchSize:     cfg.BatchSize,
		entries:       make([]models.Analytics, 0, cfg.BatchSize),
		userIDMap:     make(map[int]struct{}),
		flushInterval: cfg.FlushInterval,
	}

	ctx, cancel := context.WithCancel(context.Background())
	l.cancelFunc = cancel
	l.ticker = time.NewTicker(l.flushInterval)

	go l.periodicFlush(ctx)

	return l.middleware, l.shutdown
}

func (l *logger) middleware(c *gin.Context) {
	var requestBody []byte
	if c.Request.Body != nil {
		requestBody, _ = io.ReadAll(c.Request.Body)
		c.Request.Body = io.NopCloser(bytes.NewBuffer(requestBody))
	}

	w := &responseWriter{
		ResponseWriter: c.Writer,
		body:           bytes.NewBuffer(nil),
	}
	c.Writer = w

	c.Next()

	ctx := c.Request.Context()

	path := c.Request.URL.Path
	method := c.Request.Method
	status := c.Writer.Status()
	clientIP := c.ClientIP()

	var userID *int
	id, ok := mcontext.GetOpUserID(ctx)
	if ok {
		userID = &id
	}

	operationID, _ := mcontext.GetOperationID(ctx)

	compressedRequest := compressData(requestBody)
	compressedResponse := compressData(w.body.Bytes())

	entry := models.Analytics{
		OperationID:  operationID,
		Path:         path,
		UserID:       userID,
		Method:       method,
		StatusCode:   status,
		ClientIP:     clientIP,
		RequestBody:  compressedRequest,
		ResponseBody: compressedResponse,
	}

	l.addEntry(entry)
}

func compressData(data []byte) []byte {
	if len(data) == 0 {
		return data
	}
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	_, _ = gz.Write(data)
	gz.Close()
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
	// Не оборачиваем мьютексами потому что метод flush сам всегда вызывается внутри заблокированного мьютекса.
	if err == nil {
		l.entries = l.entries[:0]
		l.userIDMap = make(map[int]struct{})
	} else {
		log.Printf("error when inserting analytics: %v", err)
	}
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
	}
}
