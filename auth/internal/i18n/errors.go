package i18n

import "github.com/c2pc/go-pkg/v2/utils/translator"

var (
	ErrUserNotFound             = translator.Translate{translator.RU: "Пользователь не найден", translator.EN: "User not found"}
	ErrUserExists               = translator.Translate{translator.RU: "Пользователь с таким логином уже зарегистрирован", translator.EN: "A user with this login is already registered"}
	ErrUserRolesCannotBeChanged = translator.Translate{translator.RU: "Нельзя назначать пользователю другие роли", translator.EN: "User roles cannot be changed"}
	ErrUserCannotBeBlocked      = translator.Translate{translator.RU: "Пользователь не может быть заблокирован", translator.EN: "User cannot be blocked"}
	ErrUserCannotBeDeleted      = translator.Translate{translator.RU: "Пользователь не может быть удален", translator.EN: "User cannot be deleted"}
	ErrSelfCannotBeDeleted      = translator.Translate{translator.RU: "Запрещено удалять свою учетную запись", translator.EN: "It is forbidden to delete your account"}
)

var (
	ErrRoleNotFound        = translator.Translate{translator.RU: "Роль не найдена", translator.EN: "Role not found"}
	ErrRoleExists          = translator.Translate{translator.RU: "Роль уже добавлена", translator.EN: "Role has already been added"}
	ErrRoleCannotBeChanged = translator.Translate{translator.RU: "Системные роли запрещены редактировать", translator.EN: "System roles cannot be changed"}
	ErrRoleCannotBeDeleted = translator.Translate{translator.RU: "Системные роли запрещены удалять", translator.EN: "System roles cannot be deleted"}
)

var (
	ErrSessionNotFound = translator.Translate{translator.RU: "Сессия не найдена", translator.EN: "Session not found"}
	ErrAuthNoAccess    = translator.Translate{translator.RU: "Нет доступа", translator.EN: "No access"}
	ErrSSONotSupported = translator.Translate{translator.RU: "Не поддерживается", translator.EN: "Not supported"}
)

var (
	ErrFilterNotFound = translator.Translate{translator.RU: "Фильтр не найден", translator.EN: "Filter not found"}
	ErrFilterExists   = translator.Translate{translator.RU: "Фильтр уже добавлен", translator.EN: "A filter is already created"}
)
