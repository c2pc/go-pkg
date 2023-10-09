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
		logger.InfoLog(module, stdout, "Response "+blw.body.String())
	}
}

func GinRequestBodyLogMiddleware(module string, stdout bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		var buf bytes.Buffer
		tee := io.TeeReader(c.Request.Body, &buf)
		body, _ := io.ReadAll(tee)
		c.Request.Body = io.NopCloser(&buf)
		logger.InfoLog(module, stdout, "Request "+string(body))
		c.Next()
	}
}
