package i18n

import "github.com/c2pc/go-pkg/v2/utils/translator"

var (
	ErrUserNotFound             = translator.Translate{translator.RU: "Пользователь не найден", translator.EN: "User not found"}
	ErrUserExists               = translator.Translate{translator.RU: "Пользователь с таким логином уже зарегистрирован", translator.EN: "A user with this login is already registered"}
	ErrUserRolesCannotBeChanged = translator.Translate{translator.RU: "Нельзя назначать пользователю другие роли", translator.EN: "User roles cannot be changed"}
	ErrUserCannotBeBlocked      = translator.Translate{translator.RU: "Пользователь не может быть заблокирован", translator.EN: "User cannot be blocked"}
	ErrUserCannotBeDeleted      = translator.Translate{translator.RU: "Пользователя нельзя удалять", translator.EN: "User cannot be deleted"}
)

var (
	ErrRoleNotFound        = translator.Translate{translator.RU: "Роль не найдена", translator.EN: "Role not found"}
	ErrRoleExists          = translator.Translate{translator.RU: "Роль уже добавлена", translator.EN: "Role has already been added"}
	ErrRoleCannotBeChanged = translator.Translate{translator.RU: "Роль нельзя редактировать", translator.EN: "Role cannot be changed"}
	ErrRoleCannotBeDeleted = translator.Translate{translator.RU: "Роль нельзя удалять", translator.EN: "Role cannot be deleted"}
)

var (
	ErrSessionNotFound = translator.Translate{translator.RU: "Сессия не найдена", translator.EN: "Session not found"}
	ErrAuthNoAccess    = translator.Translate{translator.RU: "Нет доступа", translator.EN: "No access"}
)

var (
	ErrFilterNotFound = translator.Translate{translator.RU: "Фильтр не найден", translator.EN: "Filter not found"}
	ErrFilterExists   = translator.Translate{translator.RU: "Фильтр уже добавлен", translator.EN: "A filter is already created"}
)
