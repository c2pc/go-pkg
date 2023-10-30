package apperr

import (
	"errors"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"strconv"
)

var (
	ErrSyntax = New(StatusInternalServerError, "syntax_error").WithTextTranslate(Translate{
		"ru": "Неверный запрос",
	})
	ErrValidation = New(StatusBadRequest, "validation_error").WithTextTranslate(Translate{
		"ru": "Неверный запрос",
	})
	ErrEmptyData = New(StatusInternalServerError, "empty_data_error").WithTextTranslate(Translate{
		"ru": "Неверный запрос",
	})
	ErrInternal = New(StatusInternalServerError, "internal_error").WithTextTranslate(Translate{
		"ru": "Ошибка сервера",
	})
	ErrForbidden = New(StatusForbidden, "forbidden_error").WithTextTranslate(Translate{
		"ru": "Нет доступа",
	})
	ErrUnauthenticated = New(StatusUnauthenticated, "unauthorized_error").WithTextTranslate(Translate{
		"ru": "Ошибка аутентификации",
	})
	ErrNotFound = New(StatusNotFound, "not_found_error").WithTextTranslate(Translate{
		"ru": "Не найдено",
	})
	ErrServerIsNotAvailable = New(StatusInternalServerError, "server_is_not_available").WithTextTranslate(Translate{
		"ru": "Сервер недоступен",
	})
	Err404 = New(StatusInternalServerError, "invalid_request").WithTextTranslate(Translate{
		"ru": "Неверный запрос",
	})
)

func ParseGRPCError(err error) *APPError {
	st := status.Convert(err)
	appErr := &APPError{ID: st.Message()}
	for _, detail := range st.Details() {
		switch t := detail.(type) {
		case *errdetails.BadRequest:
			for _, v := range t.GetFieldViolations() {
				switch v.GetField() {
				case "id":
					appErr.ID = v.GetDescription()
				case "title":
					appErr.Title = v.GetDescription()
				case "text":
					appErr.Text = v.GetDescription()
				case "errors":
					appErr = appErr.WithError(errors.New(v.GetDescription()))
				case "context":
					appErr.Context = v.GetDescription()
				case "show_message_banner":
					b, _ := strconv.ParseBool(v.GetDescription())
					appErr.ShowMessageBanner = b
				}
			}
		}
	}

	switch st.Code() {
	case codes.InvalidArgument:
		appErr = appErr.WithStatus(StatusBadRequest)
	case codes.Unauthenticated:
		appErr = appErr.WithStatus(StatusUnauthenticated)
	case codes.DeadlineExceeded, codes.Unavailable:
		appErr = ErrServerIsNotAvailable
	case codes.PermissionDenied:
		appErr = appErr.WithStatus(StatusForbidden)
	case codes.NotFound:
		appErr = appErr.WithStatus(StatusNotFound)
	case codes.Internal:
		appErr = appErr.WithStatus(StatusInternalServerError)
	default:
		appErr = appErr.WithStatus(StatusInternalServerError)
	}

	return appErr
}
