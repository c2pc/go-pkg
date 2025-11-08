package logger

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync/atomic"

	"github.com/c2pc/go-pkg/v2/utils/app_data"
	"github.com/op/go-logging"
	"gopkg.in/natefinch/lumberjack.v2"
)

var (
	AppModule          = "APP"
	initialized        = atomic.Bool{}
	backendFileLeveled logging.LeveledBackend
	loggersMap         logCache
	format             = logging.MustStringFormatter("%{time:2006-01-02 15:04:05} | %{level:-8s} | %{module:-8s} %{message}")
)

type Config struct {
	Level      Level
	Path       string
	Filename   string
	MaxSizeMB  int
	MaxBackups int
	MaxAgeDays int
	Compress   bool
}

func Init(cfg Config) {
	if cfg.Filename == "" {
		cfg.Filename = "app.log"
	}

	if cfg.Path == "" {
		cfg.Path = "logs"
	}

	logDir := filepath.Join(cfg.Path, app_data.AppName)
	logPath := filepath.Join(logDir, cfg.Filename)

	_ = os.MkdirAll(filepath.Dir(logDir), 0o740)
	f, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o640)
	if err == nil {
		_ = f.Close()
	}

	rotator := &lumberjack.Logger{
		Filename:   logPath,
		MaxSize:    cfg.MaxSizeMB,
		MaxBackups: cfg.MaxBackups,
		MaxAge:     cfg.MaxAgeDays,
		Compress:   cfg.Compress,
	}

	level := convertLevel(cfg.Level)

	backendFile := logging.NewLogBackend(rotator, "", 0)
	backendFileFormatted := logging.NewBackendFormatter(backendFile, format)

	backendFileLeveled = logging.AddModuleLevel(backendFileFormatted)
	backendFileLeveled.SetLevel(level, "")

	appFileLeveled := logging.AddModuleLevel(backendFileFormatted)
	appFileLeveled.SetLevel(logging.DEBUG, AppModule)

	logging.SetBackend(backendFileLeveled, appFileLeveled)

	loggersMap.mutex.Lock()
	loggersMap.loggers = make(map[string]*logging.Logger)
	loggersMap.mutex.Unlock()

	loggersMap.addLogger(AppModule, appFileLeveled)
	initialized.Store(true)

	return
}

func convertLevel(lvl Level) logging.Level {
	switch lvl {
	case DebugLevel:
		return logging.DEBUG
	case InfoLevel:
		return logging.INFO
	case WarnLevel:
		return logging.WARNING
	case ErrorLevel:
		return logging.ERROR
	case PanicLevel:
		return logging.CRITICAL
	default:
		return logging.ERROR
	}
}

func logInfo(logger *logging.Logger, stdout bool, msg string) {
	write(stdout, msg, os.Stdout)
	if !initialized.Load() {
		return
	}
	logger.Infof(msg)
}

func logError(logger *logging.Logger, stdout bool, msg string) {
	write(stdout, msg, os.Stdout)
	if !initialized.Load() {
		return
	}
	logger.Errorf(msg)
}

func logWarning(logger *logging.Logger, stdout bool, msg string) {
	write(stdout, msg, os.Stdout)
	if !initialized.Load() {
		return
	}
	logger.Warningf(msg)
}

func logDebug(logger *logging.Logger, stdout bool, msg string) {
	write(stdout, msg, os.Stdout)
	if !initialized.Load() {
		return
	}
	logger.Debugf(msg)
}

func logCritical(logger *logging.Logger, stdout bool, msg string) {
	write(stdout, msg, os.Stdout)
	if !initialized.Load() {
		return
	}
	logger.Criticalf(msg)
}

func isEnabledForLevel(logger *logging.Logger, level logging.Level) bool {
	if !initialized.Load() {
		return false
	}

	return logger.IsEnabledFor(level)
}

func write(stdout bool, msg string, writer io.Writer) {
	if !stdout {
		return
	}
	_, _ = fmt.Fprintln(writer, msg)
}
