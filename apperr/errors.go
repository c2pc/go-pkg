package apperr

import "net/http"

var (
	ErrSyntax = New(http.StatusInternalServerError, "syntax").WithTranslate(Translate{
		"ru": "Синтаксическая ошибка",
	})
	ErrValidation = New(http.StatusBadRequest, "validation").WithTranslate(Translate{
		"ru": "Ошибка валидации",
	})
	ErrEmptyData = New(http.StatusInternalServerError, "empty_data").WithTranslate(Translate{
		"ru": "Пустые данные",
	})
	ErrInternal = New(http.StatusInternalServerError, "internal").WithTranslate(Translate{
		"ru": "Ошибка сервера",
	})
	ErrForbidden = New(http.StatusForbidden, "forbidden").WithTranslate(Translate{
		"ru": "Нет доступа",
	})
	ErrUnauthorized = New(http.StatusUnauthorized, "unauthorized").WithTranslate(Translate{
		"ru": "Ошибка авторизации",
	})
	ErrNotFoundPage = New(http.StatusNotFound, "route_not_found").WithTranslate(Translate{
		"ru": "Ресурс не найден",
	})
)
