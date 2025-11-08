package mcontext

import (
	"context"
	"net/http"

	"github.com/c2pc/go-pkg/v2/utils/constant"
)

func WithOpUserIDContext(ctx context.Context, opUserID int64) context.Context {
	return context.WithValue(ctx, constant.OpUserID, opUserID)
}

func WithOpUserLoginContext(ctx context.Context, opUserLogin string) context.Context {
	return context.WithValue(ctx, constant.OpUserLogin, opUserLogin)
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

func WithOpActionRequest(request *http.Request, opAction string) *http.Request {
	return request.WithContext(WithOpActionContext(request.Context(), opAction))
}

func WithOpErrorContext(ctx context.Context, error error) context.Context {
	return context.WithValue(ctx, constant.OpError, error)
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

func GetOpUserID(ctx context.Context) (int64, bool) {
	if ctx.Value(constant.OpUserID) != nil {
		s, ok := ctx.Value(constant.OpUserID).(int64)
		if ok {
			return s, true
		}
	}
	return 0, false
}

func GetOpUserLogin(ctx context.Context) (string, bool) {
	if ctx.Value(constant.OpUserLogin) != nil {
		s, ok := ctx.Value(constant.OpUserLogin).(string)
		if ok {
			return s, true
		}
	}
	return "", false
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

func GetOpError(ctx context.Context) (error, bool) {
	if ctx.Value(constant.OpError) != nil {
		err, ok := ctx.Value(constant.OpError).(error)
		if ok {
			return err, false
		}
	}
	return nil, false
}
