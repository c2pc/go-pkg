package logger

import (
	"fmt"
	"os"
)

// Info logs INFO messages. stdout flag indicates if message is to be written to stdout in addition to log.
func Info(msg string) {
	logInfo(loggersMap.getLogger(ModuleID), false, msg)
}

// Infof logs INFO messages. stdout flag indicates if message is to be written to stdout in addition to log.
func Infof(msg string, args ...interface{}) {
	Info(fmt.Sprintf(msg, args...))
}

// Error logs ERROR messages. stdout flag indicates if message is to be written to stdout in addition to log.
func Error(msg string) {
	logError(loggersMap.getLogger(ModuleID), false, msg)
}

// Errorf logs ERROR messages. stdout flag indicates if message is to be written to stdout in addition to log.
func Errorf(msg string, args ...interface{}) {
	Error(fmt.Sprintf(msg, args...))
}

// Warning logs WARNING messages. stdout flag indicates if message is to be written to stdout in addition to log.
func Warning(msg string) {
	logWarning(loggersMap.getLogger(ModuleID), false, msg)
}

// Warningf logs WARNING messages. stdout flag indicates if message is to be written to stdout in addition to log.
func Warningf(msg string, args ...interface{}) {
	Warning(fmt.Sprintf(msg, args...))
}

// Fatal logs CRITICAL messages and exits. stdout flag indicates if message is to be written to stdout in addition to log.
func Fatal(msg string) {
	logCritical(loggersMap.getLogger(ModuleID), msg)
	addFatalError(ModuleID, msg)
	write(false, getFatalErrorMsg(), os.Stdout)
	os.Exit(1)
}

// Fatalf logs CRITICAL messages and exits. stdout flag indicates if message is to be written to stdout in addition to log.
func Fatalf(msg string, args ...interface{}) {
	Fatal(fmt.Sprintf(msg, args...))
}

// Debug logs DEBUG messages. stdout flag indicates if message is to be written to stdout in addition to log.
func Debug(msg string) {
	logDebug(loggersMap.getLogger(ModuleID), false, msg)
}

// Debugf logs DEBUG messages. stdout flag indicates if message is to be written to stdout in addition to log.
func Debugf(msg string, args ...interface{}) {
	Debug(fmt.Sprintf(msg, args...))
}

// HandleWarningMessages logs multiple messages in WARNING mode
func HandleWarningMessages(warnings []string) {
	for _, warning := range warnings {
		Warning(warning)
	}
}
