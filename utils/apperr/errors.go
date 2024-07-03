package apperr

import (
	"github.com/c2pc/go-pkg/v2/utils/apperr/code"
	"github.com/c2pc/go-pkg/v2/utils/translator"
)

var (
	ErrSyntax = New("syntax_error",
		WithTextTranslate(translator.Translate{
			translator.RU: "Неверный запрос",
			translator.EN: "Syntax error",
		}),
		WithCode(code.InvalidArgument),
	)
	ErrValidation = New("validation_error",
		WithTextTranslate(translator.Translate{
			translator.RU: "Неверный запрос",
			translator.EN: "Validation error",
		}),
		WithCode(code.InvalidArgument),
	)
	ErrEmptyData = New("empty_data_error",
		WithTextTranslate(translator.Translate{
			translator.RU: "Неверный запрос",
			translator.EN: "Empty data error",
		}),
		WithCode(code.InvalidArgument),
	)
	ErrInternal = New("internal_error",
		WithTextTranslate(translator.Translate{
			translator.RU: "Ошибка сервера",
			translator.EN: "Internal error",
		}),
		WithCode(code.Internal),
	)
	ErrForbidden = New("forbidden_error",
		WithTextTranslate(translator.Translate{
			translator.RU: "Нет доступа",
			translator.EN: "Forbidden error",
		}),
		WithCode(code.PermissionDenied),
	)
	ErrUnauthenticated = New("unauthenticated_error",
		WithTextTranslate(translator.Translate{
			translator.RU: "Ошибка аутентификации",
			translator.EN: "Unauthenticated error",
		}),
		WithCode(code.Unauthenticated),
	)
	ErrNotFound = New("not_found_error",
		WithTextTranslate(translator.Translate{
			translator.RU: "Не найдено",
			translator.EN: "Not found error",
		}),
		WithCode(code.NotFound),
	)
	ErrServerIsNotAvailable = New("server_is_not_available",
		WithTextTranslate(translator.Translate{
			translator.RU: "Сервер недоступен",
			translator.EN: "Server is not available",
		}),
		WithCode(code.Unavailable),
	)
)

var (
	ErrDBRecordNotFound = New("db_not_found",
		WithTextTranslate(translator.Translate{
			translator.RU: "Не найдено",
			translator.EN: "Not found error",
		}),
		WithCode(code.NotFound),
	)
	ErrDBDuplicated = New("db_duplicated",
		WithTextTranslate(translator.Translate{
			translator.RU: "Запись с такими данными уже добавлена",
			translator.EN: "Column with this data already exists",
		}),
		WithCode(code.NotFound),
	)
	ErrDBInternal = New("db_internal",
		WithTextTranslate(translator.Translate{
			translator.RU: "Ошибка базы данных",
			translator.EN: "Database internal error",
		}),
		WithCode(code.NotFound),
	)
)
