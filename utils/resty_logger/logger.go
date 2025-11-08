package resty_logger

import (
	"context"

	"github.com/c2pc/go-pkg/v2/utils/logger"
)

type RestyLogger struct {
	LoggerID string
}

func (l *RestyLogger) Errorf(format string, v ...any) {
	logger.ErrorFLog(context.Background(), l.LoggerID, format, v...)
}

func (l *RestyLogger) Warnf(format string, v ...any) {
	logger.WarningFLog(context.Background(), l.LoggerID, format, v...)
}

func (l *RestyLogger) Debugf(format string, v ...any) {
	logger.DebugFLog(context.Background(), l.LoggerID, format, v...)
}
