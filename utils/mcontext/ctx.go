package mcontext

import (
	"context"

	"github.com/c2pc/go-pkg/v2/utils/constant"
)

func WithOpUserIDContext(ctx context.Context, opUserID int) context.Context {
	return context.WithValue(ctx, constant.OpUserID, opUserID)
}

func WithOpUserRoleContext(ctx context.Context, opUserRole string) context.Context {
	return context.WithValue(ctx, constant.OpUserRole, opUserRole)
}

func WithOpDeviceIDContext(ctx context.Context, device int) context.Context {
	return context.WithValue(ctx, constant.OpDeviceID, device)
}

func WithOperationIDContext(ctx context.Context, operationID string) context.Context {
	return context.WithValue(ctx, constant.OperationID, operationID)
}
func WithOpActionContext(ctx context.Context, opAction string) context.Context {
	return context.WithValue(ctx, constant.OpAction, opAction)
}

func GetOperationID(ctx context.Context) (string, bool) {
	if ctx.Value(constant.OperationID) != nil {
		s, ok := ctx.Value(constant.OperationID).(string)
		if ok {
			return s, true
		}
	}
	return "", false
}

func GetOperationID2(ctx context.Context) string {
	if ctx.Value(constant.OperationID) != nil {
		s, ok := ctx.Value(constant.OperationID).(string)
		if ok {
			return s
		}
	}
	return ""
}

func GetOpUserID(ctx context.Context) (int, bool) {
	if ctx.Value(constant.OpUserID) != nil {
		s, ok := ctx.Value(constant.OpUserID).(int)
		if ok {
			return s, true
		}
	}
	return 0, false
}

func GetOpUserRole(ctx context.Context) (string, bool) {
	if ctx.Value(constant.OpUserRole) != nil {
		s, ok := ctx.Value(constant.OpUserRole).(string)
		if ok {
			return s, true
		}
	}
	return "", false
}

func GetOpDeviceID(ctx context.Context) (int, bool) {
	if ctx.Value(constant.OpDeviceID) != nil {
		s, ok := ctx.Value(constant.OpDeviceID).(int)
		if ok {
			return s, true
		}
	}
	return 0, false
}

func GetOpAction(ctx context.Context) (string, bool) {
	if ctx.Value(constant.OpAction) != nil {
		s, ok := ctx.Value(constant.OpAction).(string)
		if ok {
			return s, true
		}
	}
	return "", false
}
