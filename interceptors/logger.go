package interceptors

import (
	"context"
	"github.com/c2pc/go-pkg/logger"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"log"
	"strings"
)

func Logger(loggerID string, newLine bool) logging.Logger {
	writer := logger.NewLogWriter(loggerID, true, 0)
	logg := log.New(writer.Stdout, "\r\n\n", log.Lmsgprefix)
	return logging.LoggerFunc(func(_ context.Context, lvl logging.Level, msg string, fields ...interface{}) {
		prefix := ""
		if newLine && msg == "started call" {
			prefix = "\t\n"
		}
		msg = strings.ToUpper(msg)

		if len(fields) < 1000 {
			logg.Printf("%s: %v - %+v", prefix, msg, fields)
		} else {
			logg.Printf("%s: %v - %+v...", prefix, msg, fields[:1000])
		}
	})
}
