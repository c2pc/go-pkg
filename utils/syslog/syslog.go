package syslog

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/RackSec/srslog"
	"github.com/c2pc/go-pkg/v2/utils/logger"
	"github.com/rs/zerolog"
)

var (
	Vendor  string
	Product string
	Version string
)

var (
	syslog = Writer{level: srslog.LOG_ERR}
)

type AddrConfig struct {
	Network string // "udp", "tcp" или "unix"
	Addr    string // "host:port" или "/dev/log"
}

type Config struct {
	Level zerolog.Level
	Addrs []AddrConfig
}

type Writer struct {
	level srslog.Priority
	mu    sync.RWMutex
	w     []*srslog.Writer
}

type Msg struct {
	Signature    string
	Severity     int
	Message      string
	Extension    string
	ExtensionMap map[string]interface{}
}

func Init(cfg Config) error {
	return reload(cfg)
}

func Reload(cfg Config) error {
	return reload(cfg)
}

func reload(cfg Config) error {
	if strings.TrimSpace(logger.AppName) == "" {
		panic("AppName must be set before Init/Reload")
	}

	if strings.TrimSpace(Vendor) == "" || strings.TrimSpace(Product) == "" || strings.TrimSpace(Version) == "" {
		panic("Vendor, Product and Version must be set before Init/Reload")
	}

	writers := make([]*srslog.Writer, 0)
	var errs []string
	for _, addr := range cfg.Addrs {
		if strings.TrimSpace(addr.Network) == "" || strings.TrimSpace(addr.Addr) == "" {
			errs = append(errs, fmt.Sprintf("invalid addr: network=%q addr=%q", addr.Network, addr.Addr))
			continue
		}
		w, err := srslog.Dial(addr.Network, addr.Addr, srslog.LOG_LOCAL0|srslog.LOG_DEBUG, logger.AppName)
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "syslog connect failed (%s %s): %v\n", addr.Network, addr.Addr, err)
			errs = append(errs, fmt.Sprintf("%s %s: %v", addr.Network, addr.Addr, err))
			continue
		}
		w.SetFormatter(srslog.RFC5424Formatter)
		writers = append(writers, w)
	}

	syslog.mu.Lock()
	defer syslog.mu.Unlock()
	syslog.w = writers
	syslog.level = cefSeverity(cfg.Level)

	if len(writers) == 0 {
		if len(errs) > 0 {
			return fmt.Errorf("no syslog writers available: %s", strings.Join(errs, "; "))
		}
	}

	return nil
}

func BuildExtension(m map[string]interface{}) string {
	if len(m) == 0 {
		return ""
	}
	parts := make([]string, 0, len(m))
	for k, v := range m {
		if strings.TrimSpace(k) == "" {
			continue
		}
		val := fmt.Sprintf("%v", v)
		val = escapeCEF(val)
		parts = append(parts, fmt.Sprintf("%s=%s", k, val))
	}
	return strings.Join(parts, " ")
}

func escapeCEF(s string) string {
	replacer := strings.NewReplacer(
		"\\", "\\\\",
		"=", "\\=",
		"|", "\\|",
		"\n", "\\n",
		"\r", "\\r",
		"\"", "",
		"'", "",
	)
	return replacer.Replace(s)
}

func (w *Writer) buildCEF(msg Msg) []byte {
	var ext string
	if msg.ExtensionMap != nil && len(msg.ExtensionMap) > 0 {
		ext = BuildExtension(msg.ExtensionMap)
	} else {
		ext = msg.Extension
	}

	var buf bytes.Buffer
	_, _ = fmt.Fprintf(&buf,
		"CEF:0|%s|%s|%s|%s|%s|%d|%s",
		Vendor, Product, Version, msg.Signature, msg.Message, msg.Severity, ext,
	)
	return buf.Bytes()
}

func (w *Writer) WriteContext(ctx context.Context, level srslog.Priority, msg Msg) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	payload := w.buildCEF(msg)

	w.mu.RLock()
	writers := make([]*srslog.Writer, len(w.w))
	copy(writers, w.w)
	w.mu.RUnlock()

	if len(writers) == 0 {
		return errors.New("no syslog writers available")
	}

	var mu sync.Mutex
	errs := make([]string, 0)
	var wg sync.WaitGroup
	wg.Add(len(writers))
	for _, writer := range writers {
		go func(writer *srslog.Writer) {
			defer wg.Done()
			select {
			case <-ctx.Done():
				mu.Lock()
				errs = append(errs, fmt.Sprintf("write canceled: %v", ctx.Err()))
				mu.Unlock()
				return
			default:
			}
			_, err := writer.WriteWithPriority(level, payload)
			if err != nil {
				mu.Lock()
				errs = append(errs, err.Error())
				mu.Unlock()
			}
		}(writer)
	}
	wg.Wait()

	if len(errs) > 0 {
		_, _ = fmt.Fprintf(os.Stderr, "syslog write errors: %s\n", strings.Join(errs, "; "))
		return fmt.Errorf("write errors: %s", strings.Join(errs, "; "))
	}

	return nil
}

func (w *Writer) Write(level srslog.Priority, msg Msg) error {
	return w.WriteContext(context.Background(), level, msg)
}

func Close() error {
	syslog.mu.Lock()
	writers := syslog.w
	syslog.w = nil
	syslog.mu.Unlock()

	if len(writers) == 0 {
		return nil
	}

	errs := make([]string, 0)
	for _, closer := range writers {
		if err := closer.Close(); err != nil {
			errs = append(errs, err.Error())
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("close errors: %s", strings.Join(errs, "; "))
	}
	return nil
}

func cefSeverity(level zerolog.Level) srslog.Priority {
	switch level {
	case zerolog.DebugLevel:
		return srslog.LOG_DEBUG
	case zerolog.InfoLevel:
		return srslog.LOG_INFO
	case zerolog.WarnLevel:
		return srslog.LOG_WARNING
	case zerolog.ErrorLevel:
		return srslog.LOG_ERR
	case zerolog.FatalLevel, zerolog.PanicLevel:
		return srslog.LOG_CRIT
	default:
		return srslog.LOG_INFO
	}
}

func Debug(msg Msg) error {
	if syslog.level <= srslog.LOG_DEBUG {
		return syslog.Write(srslog.LOG_DEBUG, msg)
	}
	return nil
}

func Info(msg Msg) error {
	if syslog.level <= srslog.LOG_INFO {
		return syslog.Write(srslog.LOG_INFO, msg)
	}
	return nil
}

func Warn(msg Msg) error {
	if syslog.level <= srslog.LOG_WARNING {
		return syslog.Write(srslog.LOG_WARNING, msg)
	}
	return nil
}

func Error(msg Msg) error {
	if syslog.level <= srslog.LOG_ERR {
		return syslog.Write(srslog.LOG_ERR, msg)
	}
	return nil
}

func Fatal(msg Msg) error {
	if syslog.level <= srslog.LOG_CRIT {
		return syslog.Write(srslog.LOG_CRIT, msg)
	}
	return nil
}

func Panic(msg Msg) error {
	return Fatal(msg)
}

func WithLevel(level srslog.Priority, msg Msg) error {
	return syslog.Write(level, msg)
}
