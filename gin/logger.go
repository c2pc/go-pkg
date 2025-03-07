package gin

import (
	"fmt"
	"strconv"
	"time"

	"github.com/c2pc/go-pkg/logger"
	"github.com/c2pc/go-pkg/rbac"
	"github.com/gin-gonic/gin"
)

func LoggerWithConfig(loggerID string, newLine bool) gin.LoggerConfig {
	writer := logger.NewLogWriter(loggerID, false, 0)
	return gin.LoggerConfig{
		Formatter: func(param gin.LogFormatterParams) string {
			userData := "unknow"
			authUser, err := rbac.User(param.Request.Context())
			if err == nil {
				userData = strconv.Itoa(authUser.ID) + ":" + authUser.Login
			}

			var statusColor, methodColor, resetColor string
			if param.IsOutputColor() {
				statusColor = param.StatusCodeColor()
				methodColor = param.MethodColor()
				resetColor = param.ResetColor()
			}

			if param.Latency > time.Minute {
				param.Latency = param.Latency.Truncate(time.Second)
			}
			prefix := ""
			if newLine {
				prefix = "\t"
			}
			return fmt.Sprintf(" %s %3d %s| %13v | %15s | %15s |%s %-7s %s %#v\n%s\n%s",
				statusColor, param.StatusCode, resetColor,
				param.Latency,
				param.ClientIP,
				userData,
				methodColor, param.Method, resetColor,
				param.Path,
				param.ErrorMessage, prefix)
		},
		Output: writer.Stdout,
	}
}
