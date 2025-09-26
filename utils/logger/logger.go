package logger

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/c2pc/go-pkg/v2/utils/constant"
	"github.com/rs/zerolog"
	"gopkg.in/natefinch/lumberjack.v2"
)

var (
	AppName string
)

var (
	log     zerolog.Logger
	closers []io.Closer
)

var DefaultLoggerConfig = Config{
	Level: zerolog.ErrorLevel,
}

type FileConfig struct {
	Enabled    bool
	Path       string
	MaxSizeMB  int
	MaxBackups int
	MaxAgeDays int
	Compress   bool
}

type Config struct {
	Level zerolog.Level
	File  *FileConfig
}

func Init(cfg Config) {
	reload(cfg)
}

func Reload(cfg Config) {
	reload(cfg)
}

func Close() {
	for _, closer := range closers {
		_ = closer.Close()
	}
}

func reload(cfg Config) {
	var writers []io.Writer
	var clrs []io.Closer

	if cfg.File != nil && cfg.File.Enabled {
		dirPath := filepath.Join(cfg.File.Path, AppName)
		_ = os.MkdirAll(dirPath, 0740)

		fileWriter := &lumberjack.Logger{
			Filename:   filepath.Join(dirPath, "app.log"),
			MaxSize:    cfg.File.MaxSizeMB,
			MaxBackups: cfg.File.MaxBackups,
			MaxAge:     cfg.File.MaxAgeDays,
			Compress:   cfg.File.Compress,
		}

		writers = append(writers, zerolog.ConsoleWriter{
			Out:        fileWriter,
			NoColor:    true,
			TimeFormat: time.RFC3339,
			FormatLevel: func(i interface{}) string {
				return strings.ToUpper(fmt.Sprintf("| %-6s|", i))
			},
			FormatMessage: func(i interface{}) string {
				if i == nil {
					return ""
				}
				return fmt.Sprintf("| %s |", i)
			},
			FormatFieldValue: func(i interface{}) string {
				if i == nil {
					return ""
				}
				return fmt.Sprintf("%s", i)
			},
			PartsOrder: []string{
				zerolog.TimestampFieldName,
				zerolog.LevelFieldName,
				string(constant.OperationID),
				string(constant.OpAction),
				zerolog.CallerFieldName,
				zerolog.MessageFieldName,
			},
			FieldsExclude: []string{string(constant.OperationID), string(constant.OpAction)},
		})

		clrs = append(clrs, fileWriter)
	}

	if len(writers) == 0 && cfg.File == nil {
		writers = append(writers, zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: time.RFC3339,
			FormatLevel: func(i interface{}) string {
				return strings.ToUpper(fmt.Sprintf("| %-6s|", i))
			},
			FormatMessage: func(i interface{}) string {
				if i == nil {
					return ""
				}
				return fmt.Sprintf("| %s |", i)
			},
			FormatFieldValue: func(i interface{}) string {
				if i == nil {
					return ""
				}
				return fmt.Sprintf("%s", i)
			},
			PartsOrder: []string{
				zerolog.TimestampFieldName,
				zerolog.LevelFieldName,
				string(constant.OperationID),
				string(constant.OpAction),
				zerolog.CallerFieldName,
				zerolog.MessageFieldName,
			},
			FieldsExclude: []string{string(constant.OperationID), string(constant.OpAction)},
		})
	}

	multi := zerolog.MultiLevelWriter(writers...)

	log = zerolog.New(multi).Level(cfg.Level).With().
		Timestamp().
		Logger()
	closers = clrs
}

func Debug() *zerolog.Event   { return log.Debug() }
func Info() *zerolog.Event    { return log.Info() }
func Warn() *zerolog.Event    { return log.Warn() }
func Error() *zerolog.Event   { return log.Error() }
func Fatal() *zerolog.Event   { return log.WithLevel(zerolog.FatalLevel) }
func Panic() *zerolog.Event   { return log.WithLevel(zerolog.PanicLevel) }
func NoLevel() *zerolog.Event { return log.WithLevel(zerolog.NoLevel) }
func GetLevel() zerolog.Level { return log.GetLevel() }
