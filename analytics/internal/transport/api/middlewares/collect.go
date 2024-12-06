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
	"github.com/c2pc/go-pkg/v2/utils/apperr"
	"github.com/c2pc/go-pkg/v2/utils/mcontext"
	response "github.com/c2pc/go-pkg/v2/utils/response/http"
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

func New(cfg LoggerConfig) gin.HandlerFunc {
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
		flushInterval: cfg.FlushInterval,
	}

	ctx, cancel := context.WithCancel(context.Background())
	l.cancelFunc = cancel
	l.ticker = time.NewTicker(l.flushInterval)

	go l.periodicFlush(ctx)

	return l.middleware
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

	path := c.FullPath()
	if path == "" {
		path = c.Request.URL.Path
	}
	method := c.Request.Method
	status := c.Writer.Status()
	clientIP := c.ClientIP()

	userID, ok := mcontext.GetOpUserID(ctx)

	if !ok {
		response.Response(c, apperr.ErrInternal.WithErrorText("error to get operation user id"))
		c.Abort()
		return
	}

	operationID, ok := mcontext.GetOperationID(ctx)
	if !ok {
		response.Response(c, apperr.ErrInternal.WithErrorText("error to get operation id"))
		c.Abort()
		return
	}

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

func compressData(data []byte) string {
	if len(data) == 0 {
		return ""
	}
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	_, _ = gz.Write(data)
	gz.Close()
	return buf.String()
}

func (l *logger) addEntry(entry models.Analytics) {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.entries = append(l.entries, entry)
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

	err := l.db.Create(&l.entries).Error
	if err == nil {
		l.entries = l.entries[:0]
	} else {
		log.Printf("Error inserting analytics batch: %v", err)
	}
}

func (l *logger) Shutdown() {
	l.cancelFunc()
	l.ticker.Stop()
	l.mu.Lock()
	defer l.mu.Unlock()
	if len(l.entries) > 0 {
		_ = l.db.Create(&l.entries).Error
		l.entries = l.entries[:0]
	}
}
