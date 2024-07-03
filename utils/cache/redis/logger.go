package redis

import (
	"context"
	"fmt"
	"github.com/c2pc/go-pkg/v2/utils/constant"
	logger2 "github.com/c2pc/go-pkg/v2/utils/logger"
	"log"
)

type logger struct {
	log *log.Logger
}

func defaultLogger() logger {
	writer := logger2.NewLogWriter(constant.REDIS_ID, true, 0)
	return logger{
		log: log.New(writer.Stdout, "\r\n\n", log.LstdFlags),
	}
}

func (l logger) Printf(ctx context.Context, format string, v ...interface{}) {
	_ = l.log.Output(2, fmt.Sprintf(format, v...))
}
