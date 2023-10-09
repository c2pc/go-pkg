package logger

import (
	"fmt"
	"os"
)

func InfoLog(module string, stdout bool, msg string) {
	Info(loggersMap.getLogger(module), stdout, msg)
}

// InfofLog logs INFO messages. stdout flag indicates if message is to be written to stdout in addition to log.
func InfofLog(module string, stdout bool, msg string, args ...interface{}) {
	InfoLog(module, stdout, fmt.Sprintf(msg, args...))
}

// ErrorLog logs ERROR messages. stdout flag indicates if message is to be written to stdout in addition to log.
func ErrorLog(module string, stdout bool, msg string) {
	Error(loggersMap.getLogger(module), stdout, msg)
}

// ErrorfLog logs ERROR messages. stdout flag indicates if message is to be written to stdout in addition to log.
func ErrorfLog(module string, stdout bool, msg string, args ...interface{}) {
	ErrorLog(module, stdout, fmt.Sprintf(msg, args...))
}

// WarningLog logs WARNING messages. stdout flag indicates if message is to be written to stdout in addition to log.
func WarningLog(module string, stdout bool, msg string) {
	Warning(loggersMap.getLogger(module), stdout, msg)
}

// WarningfLog logs WARNING messages. stdout flag indicates if message is to be written to stdout in addition to log.
func WarningfLog(module string, stdout bool, msg string, args ...interface{}) {
	WarningfLog(module, stdout, fmt.Sprintf(msg, args...))
}

// FatalLog logs CRITICAL messages and exits. stdout flag indicates if message is to be written to stdout in addition to log.
func FatalLog(module string, stdout bool, msg string) {
	Critical(loggersMap.getLogger(module), msg)
	addFatalError(module, msg)
	write(stdout, getFatalErrorMsg(), os.Stdout)
	os.Exit(1)
}

// FatalfLog logs CRITICAL messages and exits. stdout flag indicates if message is to be written to stdout in addition to log.
func FatalfLog(module string, stdout bool, msg string, args ...interface{}) {
	FatalfLog(module, stdout, fmt.Sprintf(msg, args...))
}

// DebugLog logs DEBUG messages. stdout flag indicates if message is to be written to stdout in addition to log.
func DebugLog(module string, stdout bool, msg string) {
	Debug(loggersMap.getLogger(module), stdout, msg)
}

// DebugfLog logs DEBUG messages. stdout flag indicates if message is to be written to stdout in addition to log.
func DebugfLog(module string, stdout bool, msg string, args ...interface{}) {
	DebugLog(module, stdout, fmt.Sprintf(msg, args...))
}

// HandleWarningMessagesLog logs multiple messages in WARNING mode
func HandleWarningMessagesLog(module string, stdout bool, warnings []string) {
	for _, warning := range warnings {
		WarningLog(module, stdout, warning)
	}
}
