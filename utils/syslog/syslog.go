package syslog

import (
	"context"
	"fmt"
	"log/syslog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/smithy-go/ptr"
	"github.com/c2pc/go-pkg/v2/utils/app_data"
	"github.com/c2pc/go-pkg/v2/utils/logger"
	"github.com/c2pc/go-pkg/v2/utils/secret"
	"github.com/sirupsen/logrus"
	logrussyslog "github.com/sirupsen/logrus/hooks/syslog"
	"gopkg.in/natefinch/lumberjack.v2"
)

type OnlyMessageFormatter struct{}

func (f *OnlyMessageFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	return []byte(fmt.Sprintln(entry.Message)), nil
}

var (
	log *logrus.Logger
)

type ConfigSyslog struct {
	Network string
	Addr    string
}

type Config struct {
	Path       string
	Filename   string
	MaxSizeMB  int
	MaxBackups int
	MaxAgeDays int
	Compress   bool
	Syslog     []ConfigSyslog
}

func Init(cfg Config) {
	log = logrus.New()
	log.SetFormatter(&OnlyMessageFormatter{})

	if cfg.Filename == "" {
		cfg.Filename = "audit.log"
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

	lumber := &lumberjack.Logger{
		Filename:   logPath,
		MaxSize:    cfg.MaxSizeMB,
		MaxBackups: cfg.MaxBackups,
		MaxAge:     cfg.MaxAgeDays,
		Compress:   cfg.Compress,
	}
	log.SetOutput(lumber)

	for _, sys := range cfg.Syslog {
		w, err := logrussyslog.NewSyslogHook(sys.Network, sys.Addr, syslog.LOG_LOCAL0|syslog.LOG_DEBUG, app_data.AppName)
		if err != nil {
			logger.AppWarningFLog(context.Background(), "Failed to connect to syslog: %v", err)
			continue
		}
		log.AddHook(w)
	}
}

type Record struct {
	EventID   string   //Системный идентификатор сообщения о событии
	EventName string   //Краткое наименование события
	Severity  Severity //Уровень важности
	Success   bool     //Успех/Отказ
}

func escapeField(s string) string {
	s = strings.ReplaceAll(s, "\"", "")
	s = strings.ReplaceAll(s, "'", "")
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, "|", `\|`)
	s = strings.ReplaceAll(s, "=", `\=`)
	s = strings.ReplaceAll(s, "\r\n", `\n`)
	s = strings.ReplaceAll(s, "\n", `\n`)
	s = strings.ReplaceAll(s, "\r", `\r`)
	return s
}

func addPair(pairs *[]string, key, value string, alwaysInclude bool) {
	if value == "" {
		if alwaysInclude {
			*pairs = append(*pairs, key+"=")
		}
		return
	}
	*pairs = append(*pairs, fmt.Sprintf("%s=%s", key, escapeField(value)))
}

func Write(ctx context.Context, r Record, msg string, args ...any) {
	header := fmt.Sprintf(
		"CEF:0|%s|%s|%s|%s|%s|%d|",
		escapeField(app_data.Vendor),
		escapeField(app_data.AppName),
		escapeField(app_data.AppVersion),
		escapeField(r.EventID),
		escapeField(r.EventName),
		r.Severity,
	)

	d := CtxGetData(ctx)
	if d.StartTime == nil {
		d.StartTime = ptr.Time(time.Now())
	}

	var outcome = "Отказ"
	if r.Success {
		outcome = "Успех"
	}

	var pairs []string
	addPair(&pairs, "externalId", secret.GenerateSecureID(), true)
	addPair(&pairs, "suser", d.UserLogin, true)
	addPair(&pairs, "sntdom", d.ClientHost, true)
	addPair(&pairs, "src", d.ServerIP, true)
	addPair(&pairs, "smac", d.ServerMac, false)
	addPair(&pairs, "shost", d.ServerHost, false)
	addPair(&pairs, "duser", "", false)
	addPair(&pairs, "dntdom", "", false)
	addPair(&pairs, "dst", "", false)
	addPair(&pairs, "dmac", "", false)
	addPair(&pairs, "dhost", "", false)
	addPair(&pairs, "spt", d.ClientPort, true)
	addPair(&pairs, "dpt", d.ServerPort, false)
	addPair(&pairs, "app", d.ServerProto, false)
	addPair(&pairs, "start", fmt.Sprint(d.StartTime.Unix()), true)
	addPair(&pairs, "end", "", false)
	addPair(&pairs, "rt", "", false)
	addPair(&pairs, "msg", fmt.Sprintf(msg, args...), true)
	addPair(&pairs, "deviceProcessName", app_data.AppName, true)
	addPair(&pairs, "outcome", outcome, true)

	log.Print(header + strings.Join(pairs, " "))
}
