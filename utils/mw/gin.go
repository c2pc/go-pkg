package mw

import (
	"bytes"
	"fmt"
	"github.com/c2pc/go-pkg/v2/utils/apperr"
	"github.com/c2pc/go-pkg/v2/utils/apperr/code"
	"github.com/c2pc/go-pkg/v2/utils/constant"
	"github.com/c2pc/go-pkg/v2/utils/i18n"
	"github.com/c2pc/go-pkg/v2/utils/level"
	"github.com/c2pc/go-pkg/v2/utils/logger"
	"github.com/c2pc/go-pkg/v2/utils/mcontext"
	response "github.com/c2pc/go-pkg/v2/utils/response/http"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

var (
	ErrEmptyOperationID = apperr.New("empty_operation_id_header", apperr.WithTextTranslate(i18n.ErrEmptyOperationID), apperr.WithCode(code.InvalidArgument))
)

func CorsHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "*")
		c.Header("Access-Control-Allow-Headers", "*")
		c.Header(
			"Access-Control-Expose-Headers",
			"Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers,Cache-Control,Content-Language,Content-Type,Expires,Last-Modified,Pragma,FooBar",
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

func LogHandler(moduleID, debug string) gin.LoggerConfig {
	writer := logger.NewLogWriter(moduleID, true, 0)

	prefix := ""
	if level.Is(debug, level.TEST) {
		prefix = "\t"
	}

	return gin.LoggerConfig{
		Formatter: func(param gin.LogFormatterParams) string {
			var statusColor, methodColor, resetColor string
			if param.IsOutputColor() {
				statusColor = param.StatusCodeColor()
				methodColor = param.MethodColor()
				resetColor = param.ResetColor()
			}

			if param.Latency > time.Minute {
				param.Latency = param.Latency.Truncate(time.Second)
			}

			operationID, _ := mcontext.GetOperationID(param.Request.Context())

			return fmt.Sprintf(" | %s | %s %3d %s| %13v | %15s |%s %-7s %s %#v\n%s\n%s",
				operationID,
				statusColor, param.StatusCode, resetColor,
				param.Latency,
				param.ClientIP,
				methodColor, param.Method, resetColor,
				param.Path,
				param.ErrorMessage, prefix,
			)
		},
		Output: writer.Stdout,
	}
}

func GinParseOperationID() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method == http.MethodOptions {
			c.JSON(http.StatusOK, "Options Request!")
			c.Abort()
			return
		}

		operationID := c.Request.Header.Get(constant.OperationID)
		if operationID == "" {
			response.Response(c, ErrEmptyOperationID)
			c.Abort()
			return
		}

		ctx := c.Request.Context()
		ctx = mcontext.WithOperationIDContext(ctx, operationID)

		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}

type bodyLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w bodyLogWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

func GinBodyLogMiddleware(module, debug string) gin.HandlerFunc {
	if level.Is(debug, level.TEST) {
		return func(c *gin.Context) {
			var buf bytes.Buffer
			blw := &bodyLogWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
			c.Writer = blw

			tee := io.TeeReader(c.Request.Body, &buf)
			body, _ := io.ReadAll(tee)
			c.Request.Body = io.NopCloser(&buf)

			if len(string(body)) < 1000 {
				logger.InfofLog(c.Request.Context(), module, "Request: %s", string(body))
			} else {
				logger.InfofLog(c.Request.Context(), module, "Request: %s...", string(body)[:1000])
			}

			c.Next()

			if len(blw.body.String()) < 1000 {
				logger.InfofLog(c.Request.Context(), module, "Response: %s", blw.body.String())
			} else {
				logger.InfofLog(c.Request.Context(), module, "Response: %s...", blw.body.String()[:1000])
			}
		}
	}

	return func(c *gin.Context) {}
}
