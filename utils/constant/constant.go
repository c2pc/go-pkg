package constant

// Типы для ключей контекста
type contextKey string

// Константы для заголовков
const (
	OperationIDHeader = "X-Operation-Id"             // Заголовок HTTP для отслеживания операций
	OperationID       = contextKey("X-Operation-Id") // Ключ контекста для отслеживания операций
	TxValue           = contextKey("dbTx")           // Ключ контекста для базы данных
	OpUserID          = contextKey("opUserID")       // Ключ контекста для идентификатора пользователя операции
	OpDeviceID        = contextKey("opDeviceID")     // Ключ контекста для идентификатора устройства операции
)

// Константы для идентификаторов
const (
	APP_ID   = "APP"   // Идентификатор приложения
	DB_ID    = "DB"    // Идентификатор базы данных
	REDIS_ID = "REDIS" // Идентификатор Redis
)

// Код состояния токена
const (
	NormalToken  = 0 // Токен действителен и нормален
	InValidToken = 1 // Токен недействителен
	KickedToken  = 2 // Токен был аннулирован
	ExpiredToken = 3 // Токен просрочен
)
