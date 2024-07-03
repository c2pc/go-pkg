package httperr

import (
	"github.com/c2pc/go-pkg/v2/utils/apperr/code"
	"net/http"
)

func CodeToHttp(c code.Code) int {
	switch c {
	case code.Canceled:
		return 499
	case code.Unknown:
		return http.StatusInternalServerError
	case code.InvalidArgument:
		return http.StatusBadRequest
	case code.DeadlineExceeded:
		return http.StatusGatewayTimeout
	case code.NotFound:
		return http.StatusNotFound
	case code.AlreadyExists:
		return http.StatusConflict
	case code.PermissionDenied:
		return http.StatusForbidden
	case code.ResourceExhausted:
		return http.StatusTooManyRequests
	case code.FailedPrecondition:
		return http.StatusBadRequest
	case code.Aborted:
		return http.StatusConflict
	case code.OutOfRange:
		return http.StatusBadRequest
	case code.Unimplemented:
		return http.StatusNotImplemented
	case code.Internal:
		return http.StatusInternalServerError
	case code.Unavailable:
		return http.StatusServiceUnavailable
	case code.DataLoss:
		return http.StatusInternalServerError
	case code.Unauthenticated:
		return http.StatusUnauthorized
	default:
		return CodeToHttp(code.Unknown)
	}
}
