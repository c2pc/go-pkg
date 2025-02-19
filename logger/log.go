package logger

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/op/go-logging"
	"gopkg.in/natefinch/lumberjack.v2"
)

const (
	ModuleID = "LOG"
	logs     = "logs"
)

var level logging.Level
var initialized bool
var loggersMap logCache
var fatalErrors []string

var fileLogFormat = logging.MustStringFormatter("%{time:2006/01/02 - 15:04:05.000} [%{module}] %{message}")
var fileLoggerLeveled logging.LeveledBackend

// ActiveLogFile log file represents the file which will be used for the backend logging
var ActiveLogFile string
var machineReadable bool

type logCache struct {
	mutex   sync.RWMutex
	loggers map[string]*logging.Logger
}

// getLogger gets logger for given modules. It creates a new logger for the module if not exists
func (l *logCache) getLogger(module string) *logging.Logger {
	if !initialized {
		return nil
	}
	l.mutex.RLock()
	defer l.mutex.RUnlock()
	if module == "" {
		return l.loggers[ModuleID]
	}
	if _, ok := l.loggers[module]; !ok {
		l.mutex.RUnlock()
		l.addLogger(module)
		l.mutex.RLock()
	}
	return l.loggers[module]
}

func (l *logCache) addLogger(module string) {
	logger := logging.MustGetLogger(module)
	logger.SetBackend(fileLoggerLeveled)
	l.mutex.Lock()
	defer l.mutex.Unlock()
	l.loggers[module] = logger
}

type Config struct {
	MachineReadable bool
	Dir             string
	Filename        string
	MaxSize         int
	MaxBackups      int
	MaxAge          int
	Compress        bool
}

// Initialize logger with given level
func Initialize(config Config) {
	machineReadable = config.MachineReadable
	level = loggingLevel("")
	ActiveLogFile = getLogFile(config)
	initFileLoggerBackend(config)
	loggersMap = logCache{loggers: make(map[string]*logging.Logger)}
	loggersMap.addLogger(ModuleID)
	initialized = true

	Debug("")
}

func logInfo(logger *logging.Logger, stdout bool, msg string) {
	if level >= logging.INFO {
		write(stdout, msg, os.Stdout)
	}
	if !initialized {
		return
	}
	logger.Infof(msg)
}

func logError(logger *logging.Logger, stdout bool, msg string) {
	if level >= logging.ERROR {
		write(stdout, msg, os.Stdout)
	}
	if !initialized {
		fmt.Fprint(os.Stderr, msg)
		return
	}
	logger.Errorf(msg)
}

func logWarning(logger *logging.Logger, stdout bool, msg string) {
	if level >= logging.WARNING {
		write(stdout, msg, os.Stdout)
	}
	if !initialized {
		return
	}
	logger.Warningf(msg)
}

func logDebug(logger *logging.Logger, stdout bool, msg string) {
	if level >= logging.DEBUG {
		write(stdout, msg, os.Stdout)
	}
	if !initialized {
		return
	}
	logger.Debugf(msg)
}

func logCritical(logger *logging.Logger, msg string) {
	if !initialized {
		fmt.Fprint(os.Stderr, msg)
		return
	}
	logger.Criticalf(msg)
}

func write(stdout bool, msg string, writer io.Writer) {
	//if stdout {
	//	if machineReadable {
	//		machineReadableLog(msg)
	//	} else {
	//		fmt.Fprintln(writer, msg)
	//	}
	//}
}

// OutMessage contains information for output log
type OutMessage struct {
	MessageType string `json:"type"`
	Message     string `json:"message"`
}

// ToJSON converts OutMessage into JSON
func (out *OutMessage) ToJSON() (string, error) {
	jsonMsg, err := json.Marshal(out)
	if err != nil {
		return "", err
	}
	return string(jsonMsg), nil
}

func machineReadableLog(msg string) {
	strs := strings.Split(msg, "\n")
	for _, m := range strs {
		outMessage := &OutMessage{MessageType: "out", Message: m}
		m, _ = outMessage.ToJSON()
	}
}

func initFileLoggerBackend(config Config) {
	var backend = createFileLogger(ActiveLogFile, config)
	fileFormatter := logging.NewBackendFormatter(backend, fileLogFormat)
	fileLoggerLeveled = logging.AddModuleLevel(fileFormatter)
	fileLoggerLeveled.SetLevel(logging.DEBUG, "")
}

var createFileLogger = func(name string, config Config) logging.Backend {
	err := os.MkdirAll(filepath.Dir(name), os.ModePerm)
	if err != nil {
		fmt.Println(err)
	}

	var maxSize, maxBackups, maxAge int
	var compress bool
	if config.MaxSize != 0 {
		maxSize = config.MaxSize
	} else {
		maxSize = 10
	}

	if config.MaxBackups != 0 {
		maxBackups = config.MaxBackups
	} else {
		maxBackups = 5
	}

	if config.MaxAge != 0 {
		maxAge = config.MaxAge
	} else {
		maxAge = 28
	}

	if config.Compress {
		compress = true
	} else {
		compress = false
	}

	return logging.NewLogBackend(&lumberjack.Logger{
		Filename:   name,
		MaxSize:    maxSize, // megabytes
		MaxBackups: maxBackups,
		MaxAge:     maxAge, //days
		Compress:   compress,
	}, "", 0)
}

func addLogsDirPath(logFileName string, customLogsDir string) string {
	if logFileName == "" {
		logFileName = "app.log"
	}

	if strings.Index(customLogsDir, ".") != -1 {
		customLogsDir = filepath.Dir(customLogsDir)
	}

	if customLogsDir == "" || customLogsDir == "." {
		return filepath.Join(logs, logFileName)
	}
	if strings.Index(customLogsDir, ".") == -1 {
		return filepath.Join(customLogsDir, logFileName)
	}
	return customLogsDir
}

func getLogFile(config Config) string {
	logDirPath := addLogsDirPath(config.Filename, config.Dir)
	if filepath.IsAbs(logDirPath) {
		return logDirPath
	}

	path, _ := os.Getwd()

	return filepath.Join(path, logDirPath)
}

func loggingLevel(logLevel string) logging.Level {
	if logLevel != "" {
		switch strings.ToLower(logLevel) {
		case "debug":
			return logging.DEBUG
		case "info":
			return logging.INFO
		case "warning":
			return logging.WARNING
		case "error":
			return logging.ERROR
		}
	}
	return logging.INFO
}

func getFatalErrorMsg() string {
	return fmt.Sprintf(`%s`, strings.Join(fatalErrors, "\n\n"))
}

func addFatalError(module, msg string) {
	msg = strings.TrimSpace(msg)
	fatalErrors = append(fatalErrors, fmt.Sprintf("[%s]\n%s", module, msg))
}
