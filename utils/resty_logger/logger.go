package resty_logger

import (
	"fmt"

	"github.com/c2pc/go-pkg/v2/utils/logger"
)

type RestyLogger struct {
	LoggerID string
}

func (l *RestyLogger) Errorf(format string, v ...any) {
	logger.Error().
		Msg(fmt.Sprintf(format, v...))
}

func (l *RestyLogger) Warnf(format string, v ...any) {
	logger.Warn().
		Msg(fmt.Sprintf(format, v...))
}

func (l *RestyLogger) Debugf(format string, v ...any) {
	logger.Debug().
		Msg(fmt.Sprintf(format, v...))
}
