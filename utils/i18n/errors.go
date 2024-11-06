package i18n

import "github.com/c2pc/go-pkg/v2/utils/translator"

var (
	ErrTokenMalformed   = translator.Translate{translator.RU: "Неверный токен", translator.EN: "Token malformed"}
	ErrTokenNotValidYet = translator.Translate{translator.RU: "Недействительный токен", translator.EN: "Token not valid yet"}
	ErrTokenUnknown     = translator.Translate{translator.RU: "Неизвестный токен", translator.EN: "Token unknown"}
	ErrTokenExpired     = translator.Translate{translator.RU: "Срок действия токена истек", translator.EN: "Token expired"}
	ErrTokenNotExist    = translator.Translate{translator.RU: "Токен не найден", translator.EN: "Token not exist"}
	ErrTokenKicked      = translator.Translate{translator.RU: "Токен удален", translator.EN: "Token kicked"}
)

var (
	ErrorValidationId   = translator.Translate{translator.RU: "Неверный ID", translator.EN: "Invalid ID"}
	ErrorValidationUUID = translator.Translate{translator.RU: "Неверный uuid", translator.EN: "Invalid UUID"}
)

var (
	ErrEmptyOperationID = translator.Translate{translator.RU: "Неверный запрос", translator.EN: "Invalid request"}
)

var (
	ErrSyntax               = translator.Translate{translator.RU: "Неверный запрос", translator.EN: "Syntax error"}
	ErrValidation           = translator.Translate{translator.RU: "Неверный запрос", translator.EN: "Validation error"}
	ErrEmptyData            = translator.Translate{translator.RU: "Неверный запрос", translator.EN: "Empty data error"}
	ErrInternal             = translator.Translate{translator.RU: "Ошибка сервера", translator.EN: "Internal error"}
	ErrForbidden            = translator.Translate{translator.RU: "Нет доступа", translator.EN: "Forbidden error"}
	ErrUnauthenticated      = translator.Translate{translator.RU: "Ошибка аутентификации", translator.EN: "Unauthenticated error"}
	ErrNotFound             = translator.Translate{translator.RU: "Не найдено", translator.EN: "Not found error"}
	ErrServerIsNotAvailable = translator.Translate{translator.RU: "Сервер недоступен", translator.EN: "Server is not available"}
	ErrContextCanceled      = translator.Translate{translator.RU: "Запрос отменен", translator.EN: "Request canceled"}
)

var (
	ErrDBRecordNotFound = translator.Translate{translator.RU: "Не найдено", translator.EN: "Not found error"}
	ErrDBDuplicated     = translator.Translate{translator.RU: "Запись с такими данными уже добавлена", translator.EN: "Column with this data already exists"}
	ErrDBInternal       = translator.Translate{translator.RU: "Ошибка базы данных", translator.EN: "Database internal error"}

	ErrFilterUnknownOperator = translator.Translate{translator.RU: "Неизвестный оператор (%s) для столбца (%s)", translator.EN: "Unknown operator (%s) for column (%s)"}
	ErrFilterInvalidOperator = translator.Translate{translator.RU: "Неправильный оператор (%s)", translator.EN: "Invalid operator (%s)"}
	ErrFilterUnknownColumn   = translator.Translate{translator.RU: "Неизвестный столбец (%s)", translator.EN: "Unknown column (%s)"}
	ErrFilterInvalidValue    = translator.Translate{translator.RU: "Неправильное значение (%s) для столбца (%s)", translator.EN: "Invalid operator (%s) for column (%s)"}
	ErrOrderByUnknownColumn  = translator.Translate{translator.RU: "Неизвестный столбец (%s)", translator.EN: "Unknown column (%s)"}
)
