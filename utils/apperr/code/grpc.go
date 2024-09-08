package code

import (
	"google.golang.org/grpc/codes"
)

// GrpcToCode преобразует коды ошибок gRPC в коды ошибок вашего приложения.
func GrpcToCode(c codes.Code) Code {
	switch c {
	case codes.Canceled:
		return Canceled
	case codes.Unknown:
		return Unknown
	case codes.InvalidArgument:
		return InvalidArgument
	case codes.DeadlineExceeded:
		return DeadlineExceeded
	case codes.NotFound:
		return NotFound
	case codes.AlreadyExists:
		return AlreadyExists
	case codes.PermissionDenied:
		return PermissionDenied
	case codes.ResourceExhausted:
		return ResourceExhausted
	case codes.FailedPrecondition:
		return FailedPrecondition
	case codes.Aborted:
		return Aborted
	case codes.OutOfRange:
		return OutOfRange
	case codes.Unimplemented:
		return Unimplemented
	case codes.Internal:
		return Internal
	case codes.Unavailable:
		return Unavailable
	case codes.DataLoss:
		return DataLoss
	case codes.Unauthenticated:
		return Unauthenticated
	default:
		return Unknown // Возвращаем Unknown вместо рекурсии
	}
}
