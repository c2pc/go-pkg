package resty_logger

import (
	"context"

	"github.com/c2pc/go-pkg/v2/utils/logger"
)

type RestyLogger struct {
	Ctx      context.Context
	LoggerID string
}

func (l *RestyLogger) Errorf(format string, v ...any) {
	logger.ErrorfLog(l.Ctx, l.LoggerID, format, v...)
}

func (l *RestyLogger) Warnf(format string, v ...any) {
	logger.WarningfLog(l.Ctx, l.LoggerID, format, v...)
}

func (l *RestyLogger) Debugf(format string, v ...any) {
	logger.DebugfLog(l.Ctx, l.LoggerID, format, v...)
}
