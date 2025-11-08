package mw

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/c2pc/go-pkg/v2/utils/constant"
	"github.com/c2pc/go-pkg/v2/utils/jsonutil"
	"github.com/c2pc/go-pkg/v2/utils/logger"
	"github.com/c2pc/go-pkg/v2/utils/mcontext"
	"github.com/gin-gonic/gin"
)

var HandlersFunc = []gin.HandlerFunc{
	gin.Recovery(),
	CorsHandler(),
	GinParseOperationID(),
	LogHandler(),
}

const (
	MODULE = "HTTP"
)

var (
	Formatter = func(param gin.LogFormatterParams) string {
		var statusColor, methodColor, resetColor string
		if param.IsOutputColor() {
			statusColor = param.StatusCodeColor()
			methodColor = param.MethodColor()
			resetColor = param.ResetColor()
		}

		if param.Latency > time.Minute {
			param.Latency = param.Latency.Truncate(time.Second)
		}

		userID, _ := mcontext.GetOpUserID(param.Request.Context())

		return fmt.Sprintf("user_id=%d | %s %3d %s| %13v | %15s |%s %-7s %s %#v %s",
			userID,
			statusColor, param.StatusCode, resetColor,
			param.Latency,
			param.ClientIP,
			methodColor, param.Method, resetColor,
			param.Path,
			param.ErrorMessage,
		)
	}
)

type bodyLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w *bodyLogWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

func LogHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// --- Читаем тело запроса, если JSON ---
		var requestBody []byte
		if logger.IsEnabledForLevel(MODULE, logger.DebugLevel) {
			if c.Request.Body != nil && strings.Contains(c.GetHeader("Content-Type"), "application/json") {
				bodyBytes, err := io.ReadAll(c.Request.Body)
				if err == nil {
					requestBody = bodyBytes
					c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
				}
			}
		}

		blw := &bodyLogWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
		c.Writer = blw

		c.Next()

		if raw != "" {
			path = path + "?" + raw
		}

		param := gin.LogFormatterParams{
			Request:    c.Request,
			Keys:       c.Keys,
			TimeStamp:  time.Now(),
			Latency:    time.Since(start),
			ClientIP:   c.ClientIP(),
			Method:     c.Request.Method,
			StatusCode: c.Writer.Status(),
			BodySize:   c.Writer.Size(),
			Path:       path,
		}

		err, _ := mcontext.GetOpError(param.Request.Context())
		if err != nil {
			param.ErrorMessage = err.Error()
		}

		// --- Формируем текст лога ---
		logLine := Formatter(param)

		if logger.IsEnabledForLevel(MODULE, logger.DebugLevel) {
			if len(requestBody) > 0 && strings.Contains(c.GetHeader("Content-Type"), "application/json") {
				safeReq := string(jsonutil.JsonHideImportantData(
					requestBody, "pass", "token", "pwd", "code", "secret",
				))
				logLine += fmt.Sprintf(" ⇢ Request JSON: %s", safeReq)
			}
		}

		if logger.IsEnabledForLevel(MODULE, logger.DebugLevel) {
			respBody := blw.body.String()
			if len(respBody) > 0 && strings.Contains(c.Writer.Header().Get("Content-Type"), "application/json") {
				safeResp := string(jsonutil.JsonHideImportantData(
					[]byte(respBody), "pass", "token", "pwd", "code", "secret",
				))
				logLine += fmt.Sprintf(" ⇠ Response JSON: %s", safeResp)
			}
		}

		// --- Логируем по уровню ---
		switch {
		case param.StatusCode < 400:
			if param.Method == http.MethodPost || param.Method == http.MethodPut ||
				param.Method == http.MethodPatch || param.Method == http.MethodDelete {
				logger.InfoFLog(c.Request.Context(), MODULE, logLine)
			} else {
				logger.DebugFLog(c.Request.Context(), MODULE, logLine)
			}
		case param.StatusCode < 500:
			logger.ErrorFLog(c.Request.Context(), MODULE, logLine)
		default:
			logger.FatalFLog(c.Request.Context(), MODULE, logLine)
		}
	}
}

func CorsHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "*")
		c.Header("Access-Control-Allow-Headers", "*")
		c.Header(
			"Access-Control-Expose-Headers",
			"Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers,Cache-Control,Content-Language,Content-Type,Expires,Last-Modified,Pragma,FooBar,X-Operation-Id,X-Total-Count,X-Broker",
		) // Cross-domain key settings allow browsers to resolve.
		c.Header(
			"Access-Control-Max-Age",
			"172800",
		) // Cache request information in seconds.
		c.Header(
			"Access-Control-Allow-Credentials",
			"false",
		) //  Whether cross-domain requests need to carry cookie information, the default setting is true.
		c.Header(
			"content-type",
			"application/json",
		) // Set the return format to json.
		// Release all option pre-requests
		if c.Request.Method == http.MethodOptions {
			c.JSON(http.StatusOK, "Options Request!")
			c.Abort()
			return
		}
		c.Next()
	}
}

func GinParseOperationID() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method == http.MethodOptions {
			c.JSON(http.StatusOK, "Options Request!")
			c.Abort()
			return
		}

		operationID := c.Request.Header.Get(constant.OperationIDHeader)
		if operationID == "" {
			operationID = strconv.FormatInt(time.Now().UnixMilli(), 10)
		}

		ctx := c.Request.Context()
		ctx = mcontext.WithOperationIDContext(ctx, operationID)

		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}
