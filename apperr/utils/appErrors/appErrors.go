package appErrors

import (
	"github.com/c2pc/go-pkg/apperr"
	"github.com/c2pc/go-pkg/apperr/utils/code"
	"github.com/c2pc/go-pkg/apperr/utils/translate"
)

var (
	ErrSyntax = apperr.New("syntax_error",
		apperr.WithTextTranslate(translate.Translate{translate.RU: "Неверный запрос"}),
		apperr.WithCode(code.InvalidArgument),
	)
	ErrValidation = apperr.New("validation_error",
		apperr.WithTextTranslate(translate.Translate{translate.RU: "Неверный запрос"}),
		apperr.WithCode(code.InvalidArgument),
	)
	ErrEmptyData = apperr.New("empty_data_error",
		apperr.WithTextTranslate(translate.Translate{translate.RU: "Неверный запрос"}),
		apperr.WithCode(code.InvalidArgument),
	)
	ErrInternal = apperr.New("internal_error",
		apperr.WithTextTranslate(translate.Translate{translate.RU: "Ошибка сервера"}),
		apperr.WithCode(code.Internal),
	)
	ErrForbidden = apperr.New("forbidden_error",
		apperr.WithTextTranslate(translate.Translate{translate.RU: "Нет доступа"}),
		apperr.WithCode(code.PermissionDenied),
	)
	ErrUnauthenticated = apperr.New("unauthenticated_error",
		apperr.WithTextTranslate(translate.Translate{translate.RU: "Ошибка аутентификации"}),
		apperr.WithCode(code.Unauthenticated),
	)
	ErrNotFound = apperr.New("not_found_error",
		apperr.WithTextTranslate(translate.Translate{translate.RU: "Не найдено"}),
		apperr.WithCode(code.NotFound),
	)
	ErrServerIsNotAvailable = apperr.New("server_is_not_available",
		apperr.WithTextTranslate(translate.Translate{translate.RU: "Сервер недоступен"}),
		apperr.WithCode(code.Unavailable),
	)
)
