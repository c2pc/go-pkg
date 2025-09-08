package logger

import (
	"context"

	"github.com/c2pc/go-pkg/v2/utils/logger"
)

const module = "AUTH_CONFIG"

func LogInfo(ctx context.Context, msg string, a ...interface{}) {
	logger.InfofLog(ctx, module, msg, a...)
}
