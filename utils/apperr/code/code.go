package code

import "strconv"

// Code представляет собой тип для кодов ошибок.
type Code uint32

const (
	// OK указывает на успешное выполнение операции.
	OK Code = iota

	// Canceled указывает на отмену операции.
	Canceled

	// Unknown указывает на неизвестную ошибку.
	Unknown

	// InvalidArgument указывает на недопустимый аргумент.
	InvalidArgument

	// DeadlineExceeded указывает на истечение срока выполнения.
	DeadlineExceeded

	// NotFound указывает на то, что запрашиваемый ресурс не найден.
	NotFound

	// AlreadyExists указывает на то, что ресурс уже существует.
	AlreadyExists

	// PermissionDenied указывает на отсутствие разрешения.
	PermissionDenied

	// ResourceExhausted указывает на исчерпание ресурсов.
	ResourceExhausted

	// FailedPrecondition указывает на ошибку предварительного условия.
	FailedPrecondition

	// Aborted указывает на прерывание операции.
	Aborted

	// OutOfRange указывает на значение за пределами допустимого диапазона.
	OutOfRange

	// Unimplemented указывает на не реализованную функцию.
	Unimplemented

	// Internal указывает на внутреннюю ошибку сервера.
	Internal

	// Unavailable указывает на недоступность сервиса.
	Unavailable

	// DataLoss указывает на потерю данных.
	DataLoss

	// Unauthenticated указывает на отсутствие аутентификации.
	Unauthenticated
)

// String возвращает строковое представление кода ошибки.
func (c Code) String() string {
	switch c {
	case OK:
		return "OK"
	case Canceled:
		return "Canceled"
	case Unknown:
		return "Unknown"
	case InvalidArgument:
		return "InvalidArgument"
	case DeadlineExceeded:
		return "DeadlineExceeded"
	case NotFound:
		return "NotFound"
	case AlreadyExists:
		return "AlreadyExists"
	case PermissionDenied:
		return "PermissionDenied"
	case ResourceExhausted:
		return "ResourceExhausted"
	case FailedPrecondition:
		return "FailedPrecondition"
	case Aborted:
		return "Aborted"
	case OutOfRange:
		return "OutOfRange"
	case Unimplemented:
		return "Unimplemented"
	case Internal:
		return "Internal"
	case Unavailable:
		return "Unavailable"
	case DataLoss:
		return "DataLoss"
	case Unauthenticated:
		return "Unauthenticated"
	default:
		return "Code(" + strconv.FormatInt(int64(c), 10) + ")"
	}
}
