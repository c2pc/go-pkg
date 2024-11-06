package logger

import (
	"context"

	"github.com/c2pc/go-pkg/v2/utils/logger"
)

const module = "TASK"

func LogInfo(ctx context.Context, msg string, a ...interface{}) {
	logger.InfofLog(ctx, module, msg, a...)
}
