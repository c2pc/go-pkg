package db

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/c2pc/go-pkg/v2/utils/constant"
	loggerServ "github.com/c2pc/go-pkg/v2/utils/logger"
	"github.com/c2pc/go-pkg/v2/utils/mcontext"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
)

type logger struct {
}

func (l *logger) LogMode(level gormLogger.LogLevel) gormLogger.Interface {
	return l
}

func (l *logger) Info(ctx context.Context, msg string, data ...interface{}) {
	loggerServ.Debug().
		Str(string(constant.OperationID), mcontext.GetOperationID2(ctx)).
		Msg(msg)
}

func (l *logger) Warn(ctx context.Context, msg string, data ...interface{}) {
	loggerServ.Warn().
		Str(string(constant.OperationID), mcontext.GetOperationID2(ctx)).
		Msg(msg)
}

func (l *logger) Error(ctx context.Context, msg string, data ...interface{}) {
	loggerServ.Error().
		Str(string(constant.OperationID), mcontext.GetOperationID2(ctx)).
		Msg(msg)
}

func (l *logger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	elapsed := time.Since(begin)
	switch {
	case err != nil && !errors.Is(err, gorm.ErrRecordNotFound):
		sql, rows := fc()
		if rows == -1 {
			loggerServ.Error().
				Str(string(constant.OperationID), mcontext.GetOperationID2(ctx)).
				Str("duration", fmt.Sprintf("%.3fms", float64(elapsed.Nanoseconds())/1e6)).
				Err(err).
				Msg(sql)
		} else {
			loggerServ.Error().
				Str(string(constant.OperationID), mcontext.GetOperationID2(ctx)).
				Str("duration", fmt.Sprintf("%.3fms", float64(elapsed.Nanoseconds())/1e6)).
				Err(err).
				Msg(fmt.Sprintf("[rows:%d] ", rows) + sql)
		}
	case elapsed.Milliseconds() > 1000:
		sql, rows := fc()
		slowLog := fmt.Sprintf("SLOW SQL >= %v", 1000)
		if rows == -1 {
			loggerServ.Warn().
				Str(string(constant.OperationID), mcontext.GetOperationID2(ctx)).
				Str("duration", fmt.Sprintf("%.3fms", float64(elapsed.Nanoseconds())/1e6)).
				Str("slow_sql", slowLog).
				Err(err).
				Msg(sql)
		} else {
			loggerServ.Warn().
				Str(string(constant.OperationID), mcontext.GetOperationID2(ctx)).
				Str("duration", fmt.Sprintf("%.3fms", float64(elapsed.Nanoseconds())/1e6)).
				Str("slow_sql", slowLog).
				Err(err).
				Msg(fmt.Sprintf("[rows:%d] ", rows) + sql)
		}
	default:
		sql, rows := fc()
		if rows == -1 {
			loggerServ.Debug().
				Str(string(constant.OperationID), mcontext.GetOperationID2(ctx)).
				Str("duration", fmt.Sprintf("%.3fms", float64(elapsed.Nanoseconds())/1e6)).
				Err(err).
				Msg(sql)
		} else {
			loggerServ.Debug().
				Str(string(constant.OperationID), mcontext.GetOperationID2(ctx)).
				Str("duration", fmt.Sprintf("%.3fms", float64(elapsed.Nanoseconds())/1e6)).
				Err(err).
				Msg(fmt.Sprintf("[rows:%d] ", rows) + sql)
		}
	}
}

func (l *logger) ParamsFilter(ctx context.Context, sql string, params ...interface{}) (string, []interface{}) {
	masked := make([]interface{}, len(params))
	for i := range params {
		masked[i] = "?"
	}
	return sql, masked
}
