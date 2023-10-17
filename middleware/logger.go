package middleware

import (
	"bytes"
	"github.com/c2pc/go-pkg/logger"
	"github.com/gin-gonic/gin"
	"io"
)

type bodyLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w bodyLogWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

func GinResponseBodyLogMiddleware(module string, stdout bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		blw := &bodyLogWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
		c.Writer = blw
		c.Next()
		if len(blw.body.String()) < 1000 {
			logger.InfofLog(module, stdout, "Response: %s"+blw.body.String())
		} else {
			logger.InfofLog(module, stdout, "Response: %s..."+blw.body.String()[:1000])
		}

	}
}

func GinRequestBodyLogMiddleware(module string, stdout bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		var buf bytes.Buffer
		tee := io.TeeReader(c.Request.Body, &buf)
		body, _ := io.ReadAll(tee)
		c.Request.Body = io.NopCloser(&buf)
		if len(string(body)) < 1000 {
			logger.InfofLog(module, stdout, "Request: %s"+string(body))
		} else {
			logger.InfofLog(module, stdout, "Request: %s..."+string(body)[:1000])
		}

		c.Next()
	}
}
