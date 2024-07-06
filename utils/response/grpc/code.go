package grpc

import (
	"github.com/c2pc/go-pkg/v2/utils/apperr/code"
	"google.golang.org/grpc/codes"
)

func CodeToGrpc(c code.Code) codes.Code {
	switch c {
	case code.Canceled:
		return codes.Canceled
	case code.Unknown:
		return codes.Unknown
	case code.InvalidArgument:
		return codes.InvalidArgument
	case code.DeadlineExceeded:
		return codes.DeadlineExceeded
	case code.NotFound:
		return codes.NotFound
	case code.AlreadyExists:
		return codes.AlreadyExists
	case code.PermissionDenied:
		return codes.PermissionDenied
	case code.ResourceExhausted:
		return codes.ResourceExhausted
	case code.FailedPrecondition:
		return codes.FailedPrecondition
	case code.Aborted:
		return codes.Aborted
	case code.OutOfRange:
		return codes.OutOfRange
	case code.Unimplemented:
		return codes.Unimplemented
	case code.Internal:
		return codes.Internal
	case code.Unavailable:
		return codes.Unavailable
	case code.DataLoss:
		return codes.DataLoss
	case code.Unauthenticated:
		return codes.Unauthenticated
	default:
		return CodeToGrpc(code.Unknown)
	}
}
