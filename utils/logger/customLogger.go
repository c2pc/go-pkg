package logger

import (
	"context"
	"fmt"
	"os"

	"github.com/c2pc/go-pkg/v2/utils/mcontext"
)

func WithOperationID(ctx context.Context, msg string) string {
	operationID, ok := mcontext.GetOperationID(ctx)
	if ok {
		return fmt.Sprintf("| %s | %s", operationID, msg)
	}

	return msg
}

func InfoLog(ctx context.Context, module string, msg string) {
	logInfo(loggersMap.getLogger(module), true, WithOperationID(ctx, msg))
}

func InfofLog(ctx context.Context, module string, msg string, args ...interface{}) {
	InfoLog(ctx, module, fmt.Sprintf(msg, args...))
}

func ErrorLog(ctx context.Context, module string, msg string) {
	logError(loggersMap.getLogger(module), true, WithOperationID(ctx, msg))
}

func ErrorfLog(ctx context.Context, module string, msg string, args ...interface{}) {
	ErrorLog(ctx, module, fmt.Sprintf(msg, args...))
}

func WarningLog(ctx context.Context, module string, msg string) {
	logWarning(loggersMap.getLogger(module), true, WithOperationID(ctx, msg))
}

func WarningfLog(ctx context.Context, module string, msg string, args ...interface{}) {
	WarningLog(ctx, module, fmt.Sprintf(msg, args...))
}

func FatalLog(ctx context.Context, module string, msg string) {
	logCritical(loggersMap.getLogger(module), WithOperationID(ctx, msg))
	addFatalError(module, msg)
	write(true, getFatalErrorMsg(), os.Stdout)
	os.Exit(1)
}

func FatalfLog(ctx context.Context, module string, msg string, args ...interface{}) {
	FatalLog(ctx, module, fmt.Sprintf(msg, args...))
}

func DebugLog(ctx context.Context, module string, msg string) {
	logDebug(loggersMap.getLogger(module), true, WithOperationID(ctx, msg))
}

func DebugfLog(ctx context.Context, module string, msg string, args ...interface{}) {
	DebugLog(ctx, module, fmt.Sprintf(msg, args...))
}

func HandleWarningMessagesLog(ctx context.Context, module string, warnings []string) {
	for _, warning := range warnings {
		WarningLog(ctx, module, warning)
	}
}
