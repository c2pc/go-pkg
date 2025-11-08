package logger

import (
	"context"
	"fmt"
	"strings"

	"github.com/c2pc/go-pkg/v2/utils/mcontext"
)

func WithOperationID(ctx context.Context, msg string) string {
	msg = strings.ReplaceAll(msg, "\n", " ")
	msg = strings.ReplaceAll(msg, "\r", " ")

	operationID, ok := mcontext.GetOperationID(ctx)
	if ok {
		return fmt.Sprintf("| %19s | %s", operationID, msg)
	}

	return fmt.Sprintf("| %19s | %s", "", msg)
}

func InfoLog(ctx context.Context, module string, msg string) {
	logInfo(loggersMap.getLogger(module), false, WithOperationID(ctx, msg))
}

func InfoFLog(ctx context.Context, module string, msg string, args ...interface{}) {
	InfoLog(ctx, module, fmt.Sprintf(msg, args...))
}

func ErrorLog(ctx context.Context, module string, msg string) {
	logError(loggersMap.getLogger(module), false, WithOperationID(ctx, msg))
}

func ErrorFLog(ctx context.Context, module string, msg string, args ...interface{}) {
	ErrorLog(ctx, module, fmt.Sprintf(msg, args...))
}

func WarningLog(ctx context.Context, module string, msg string) {
	logWarning(loggersMap.getLogger(module), false, WithOperationID(ctx, msg))
}

func WarningFLog(ctx context.Context, module string, msg string, args ...interface{}) {
	WarningLog(ctx, module, fmt.Sprintf(msg, args...))
}

func FatalLog(ctx context.Context, module string, msg string) {
	logCritical(loggersMap.getLogger(module), false, WithOperationID(ctx, msg))
}

func FatalFLog(ctx context.Context, module string, msg string, args ...interface{}) {
	FatalLog(ctx, module, fmt.Sprintf(msg, args...))
}

func DebugLog(ctx context.Context, module string, msg string) {
	logDebug(loggersMap.getLogger(module), false, WithOperationID(ctx, msg))
}

func DebugFLog(ctx context.Context, module string, msg string, args ...interface{}) {
	DebugLog(ctx, module, fmt.Sprintf(msg, args...))
}

func IsEnabledForLevel(module string, level Level) bool {
	return isEnabledForLevel(loggersMap.getLogger(module), convertLevel(level))
}

func AppInfoLog(ctx context.Context, msg string) {
	logInfo(loggersMap.getLogger(AppModule), true, WithOperationID(ctx, msg))
}

func AppInfoFLog(ctx context.Context, msg string, args ...interface{}) {
	AppInfoLog(ctx, fmt.Sprintf(msg, args...))
}

func AppErrorLog(ctx context.Context, msg string) {
	logError(loggersMap.getLogger(AppModule), true, WithOperationID(ctx, msg))
}

func AppErrorFLog(ctx context.Context, msg string, args ...interface{}) {
	AppErrorLog(ctx, fmt.Sprintf(msg, args...))
}

func AppWarningLog(ctx context.Context, msg string) {
	logWarning(loggersMap.getLogger(AppModule), true, WithOperationID(ctx, msg))
}

func AppWarningFLog(ctx context.Context, msg string, args ...interface{}) {
	AppWarningLog(ctx, fmt.Sprintf(msg, args...))
}

func AppFatalLog(ctx context.Context, msg string) {
	logCritical(loggersMap.getLogger(AppModule), true, WithOperationID(ctx, msg))
}

func AppFatalFLog(ctx context.Context, msg string, args ...interface{}) {
	AppFatalLog(ctx, fmt.Sprintf(msg, args...))
}

func AppDebugLog(ctx context.Context, msg string) {
	logDebug(loggersMap.getLogger(AppModule), true, WithOperationID(ctx, msg))
}

func AppDebugFLog(ctx context.Context, msg string, args ...interface{}) {
	AppDebugLog(ctx, fmt.Sprintf(msg, args...))
}
