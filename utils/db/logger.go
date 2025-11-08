package db

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/c2pc/go-pkg/v2/utils/app_data"
	log "github.com/c2pc/go-pkg/v2/utils/logger"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
)

const MODULE = "DB"

// NewLogger initialize logger
func NewLogger(config gormLogger.Config) gormLogger.Interface {
	var (
		infoStr      = "%s\n[debug] "
		warnStr      = "%s\n[warn] "
		errStr       = "%s\n[error] "
		traceStr     = "[%.3fms] [rows:%v] %s"
		traceWarnStr = "%s [%.3fms] [rows:%v] %s"
		traceErrStr  = "%s [%.3fms] [rows:%v] %s"
	)

	if config.Colorful {
		infoStr = gormLogger.Green + "%s\n" + gormLogger.Reset + gormLogger.Green + "[info] " + gormLogger.Reset
		warnStr = gormLogger.BlueBold + "%s\n" + gormLogger.Reset + gormLogger.Magenta + "[warn] " + gormLogger.Reset
		errStr = gormLogger.Magenta + "%s\n" + gormLogger.Reset + gormLogger.Red + "[error] " + gormLogger.Reset
		traceStr = gormLogger.Green + "%s\n" + gormLogger.Reset + gormLogger.Yellow + "[%.3fms] " + gormLogger.BlueBold + "[rows:%v]" + gormLogger.Reset + " %s"
		traceWarnStr = gormLogger.Green + "%s " + gormLogger.Yellow + "%s\n" + gormLogger.Reset + gormLogger.RedBold + "[%.3fms] " + gormLogger.Yellow + "[rows:%v]" + gormLogger.Magenta + " %s" + gormLogger.Reset
		traceErrStr = gormLogger.RedBold + "%s " + gormLogger.MagentaBold + "%s\n" + gormLogger.Reset + gormLogger.Yellow + "[%.3fms] " + gormLogger.BlueBold + "[rows:%v]" + gormLogger.Reset + " %s"
	}

	if config.ParameterizedQueries {
		config.ParameterizedQueries = !app_data.IsDevMode()
	}

	return &logger{
		Config:       config,
		infoStr:      infoStr,
		warnStr:      warnStr,
		errStr:       errStr,
		traceStr:     traceStr,
		traceWarnStr: traceWarnStr,
		traceErrStr:  traceErrStr,
	}
}

type logger struct {
	gormLogger.Config
	infoStr, warnStr, errStr            string
	traceStr, traceErrStr, traceWarnStr string
}

// LogMode log mode
func (l *logger) LogMode(level gormLogger.LogLevel) gormLogger.Interface {
	newlogger := *l
	newlogger.LogLevel = level
	return &newlogger
}

// Info print info
func (l *logger) Info(ctx context.Context, msg string, data ...interface{}) {
	log.DebugFLog(ctx, MODULE, l.infoStr+msg)
}

// Warn print warn messages
func (l *logger) Warn(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= gormLogger.Warn {
		log.WarningFLog(ctx, MODULE, l.warnStr+msg)
	}
}

// Error print error messages
func (l *logger) Error(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= gormLogger.Error {
		log.ErrorFLog(ctx, MODULE, l.errStr+msg)
	}
}

// Trace print sql message
//
//nolint:cyclop
func (l *logger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	if l.LogLevel <= gormLogger.Silent {
		return
	}

	elapsed := time.Since(begin)
	switch {
	case err != nil && l.LogLevel >= gormLogger.Error && (!errors.Is(err, gorm.ErrRecordNotFound) || !l.IgnoreRecordNotFoundError):
		sql, rows := fc()
		if rows == -1 {
			log.ErrorFLog(ctx, MODULE, l.traceErrStr, err, float64(elapsed.Nanoseconds())/1e6, "-", sql)
		} else {
			log.ErrorFLog(ctx, MODULE, l.traceErrStr, err, float64(elapsed.Nanoseconds())/1e6, rows, sql)
		}
	case elapsed > l.SlowThreshold && l.SlowThreshold != 0 && l.LogLevel >= gormLogger.Warn:
		sql, rows := fc()
		slowLog := fmt.Sprintf("SLOW SQL >= %v", l.SlowThreshold)
		if rows == -1 {
			log.WarningFLog(ctx, MODULE, l.traceWarnStr, slowLog, float64(elapsed.Nanoseconds())/1e6, "-", sql)
		} else {
			log.WarningFLog(ctx, MODULE, l.traceWarnStr, slowLog, float64(elapsed.Nanoseconds())/1e6, rows, sql)
		}
	case l.LogLevel == gormLogger.Info:
		sql, rows := fc()
		if rows == -1 {
			log.DebugFLog(ctx, MODULE, l.traceStr, float64(elapsed.Nanoseconds())/1e6, "-", sql)
		} else {
			log.DebugFLog(ctx, MODULE, l.traceStr, float64(elapsed.Nanoseconds())/1e6, rows, sql)
		}
	}
}

// ParamsFilter filter params
func (l *logger) ParamsFilter(ctx context.Context, sql string, params ...interface{}) (string, []interface{}) {
	if l.Config.ParameterizedQueries {
		return sql, nil
	}

	return sql, params
}
