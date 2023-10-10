package apperr

import (
	"google.golang.org/grpc/codes"
	"net/http"
)

type Code int

const (
	StatusBadRequest Code = iota
	StatusUnauthenticated
	StatusForbidden
	StatusNotFound

	StatusInternalServerError
)

func (c Code) HTTP() int {
	switch c {
	case StatusBadRequest:
		return http.StatusBadRequest
	case StatusUnauthenticated:
		return http.StatusUnauthorized
	case StatusForbidden:
		return http.StatusForbidden
	case StatusNotFound:
		return http.StatusNotFound
	case StatusInternalServerError:
		return http.StatusInternalServerError
	default:
		return http.StatusInternalServerError
	}
}

func (c Code) GRPC() codes.Code {
	switch c {
	case StatusBadRequest:
		return codes.InvalidArgument
	case StatusUnauthenticated:
		return codes.Unauthenticated
	case StatusForbidden:
		return codes.PermissionDenied
	case StatusNotFound:
		return codes.NotFound
	case StatusInternalServerError:
		return codes.Internal
	default:
		return codes.Internal
	}
}
