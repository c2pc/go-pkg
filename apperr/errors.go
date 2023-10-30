package apperr

var (
	ErrSyntax = New(StatusInternalServerError, "syntax").WithTextTranslate(Translate{
		"ru": "Неверный запрос",
	})
	ErrValidation = New(StatusBadRequest, "validation").WithTextTranslate(Translate{
		"ru": "Неверный запрос",
	})
	ErrEmptyData = New(StatusInternalServerError, "empty_data").WithTextTranslate(Translate{
		"ru": "Неверный запрос",
	})
	ErrInternal = New(StatusInternalServerError, "internal").WithTextTranslate(Translate{
		"ru": "Ошибка сервера",
	})
	ErrForbidden = New(StatusForbidden, "forbidden").WithTextTranslate(Translate{
		"ru": "Нет доступа",
	})
	ErrUnauthenticated = New(StatusUnauthenticated, "unauthorized").WithTextTranslate(Translate{
		"ru": "Ошибка аутентификации",
	})
	ErrNotFound = New(StatusNotFound, "not_found").WithTextTranslate(Translate{
		"ru": "Не найдено",
	})
	ErrServerIsNotAvailable = New(StatusInternalServerError, "server_is_not_available").WithTextTranslate(Translate{
		"ru": "Сервер недоступен",
	})
	Err404 = New(StatusInternalServerError, "invalid_request").WithTextTranslate(Translate{
		"ru": "Неверный запрос",
	})
)
