package rest

import (
	"fmt"
	"github.com/c2pc/go-pkg/v2/utils/level"
	"github.com/c2pc/go-pkg/v2/utils/logger"
	"github.com/c2pc/go-pkg/v2/utils/mcontext"
	"github.com/gin-gonic/gin"
	"time"
)

func NewLogger(moduleID, debug string) gin.LoggerConfig {
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
