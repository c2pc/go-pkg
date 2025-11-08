package constant

// Типы для ключей контекста
type contextKey string

const (
	OperationIDHeader = "X-Operation-Id" // Заголовок HTTP для отслеживания операций
)

const (
	OperationID = contextKey("X-Operation-Id")
	TxValue     = contextKey("dbTx")
	OpUserID    = contextKey("opUserID")
	OpUserLogin = contextKey("OpUserLogin")
	OpUserRole  = contextKey("opUserRole")
	OpDeviceID  = contextKey("opDeviceID")
	OpAction    = contextKey("opAction")
	OpError     = contextKey("OpError")
)

// Код состояния токена
const (
	NormalToken  = 0 // Токен действителен и нормален
	InValidToken = 1 // Токен недействителен
	KickedToken  = 2 // Токен был аннулирован
	ExpiredToken = 3 // Токен просрочен
)
