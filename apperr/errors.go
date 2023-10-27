package apperr

var (
	ErrSyntax = New(StatusInternalServerError, "syntax").WithTranslate(Translate{
		"ru": "Синтаксическая ошибка",
	})
	ErrValidation = New(StatusBadRequest, "validation").WithTranslate(Translate{
		"ru": "Ошибка валидации",
	})
	ErrEmptyData = New(StatusInternalServerError, "empty_data").WithTranslate(Translate{
		"ru": "Пустые данные",
	})
	ErrInternal = New(StatusInternalServerError, "internal").WithTranslate(Translate{
		"ru": "Ошибка сервера",
	})
	ErrForbidden = New(StatusForbidden, "forbidden").WithTranslate(Translate{
		"ru": "Нет доступа",
	})
	ErrUnauthenticated = New(StatusUnauthenticated, "unauthorized").WithTranslate(Translate{
		"ru": "Ошибка аутентификации",
	})
	ErrNotFound = New(StatusNotFound, "not_found").WithTranslate(Translate{
		"ru": "Не найдено",
	})
	ErrServerIsNotAvailable = New(StatusInternalServerError, "server_is_not_available").WithTranslate(Translate{
		"ru": "Сервер недоступен",
	})
	Err404 = New(StatusInternalServerError, "404_error").WithTranslate(Translate{
		"ru": "Ресурс не найден",
	})
)
