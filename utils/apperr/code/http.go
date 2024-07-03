package code

import (
	"net/http"
)

func HttpToCode(c int) Code {
	switch c {
	case 499:
		return Canceled
	case http.StatusBadRequest:
		return InvalidArgument
	case http.StatusGatewayTimeout:
		return DeadlineExceeded
	case http.StatusNotFound:
		return NotFound
	case http.StatusConflict:
		return AlreadyExists
	case http.StatusForbidden:
		return PermissionDenied
	case http.StatusTooManyRequests:
		return ResourceExhausted
	case http.StatusNotImplemented:
		return Unimplemented
	case http.StatusInternalServerError:
		return Internal
	case http.StatusServiceUnavailable:
		return Unavailable
	case http.StatusUnauthorized:
		return Unauthenticated
	default:
		return HttpToCode(http.StatusInternalServerError)
	}
}
