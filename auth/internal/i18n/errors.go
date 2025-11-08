package i18n

import "github.com/c2pc/go-pkg/v2/utils/translator"

var (
	ErrUserNotFound             = translator.Translate{translator.RU: "Администратор не найден", translator.EN: "Admin not found"}
	ErrUserExists               = translator.Translate{translator.RU: "Администратор с таким логином уже существует", translator.EN: "A admin with this login is already exists"}
	ErrUserRolesCannotBeChanged = translator.Translate{translator.RU: "В системе должен быть хотя бы один администратор с ролью SuperAdmin", translator.EN: "The system must have at least one admin with the SuperAdmin role"}
	ErrUserCannotBeBlocked      = translator.Translate{translator.RU: "В системе должен быть хотя бы один активный администратор с ролью SuperAdmin", translator.EN: "The system must have at least one active admin with the SuperAdmin role"}
	ErrUserCannotBeDeleted      = translator.Translate{translator.RU: "В системе должен быть хотя бы один администратор с ролью SuperAdmin", translator.EN: "The system must have at least one admin with the SuperAdmin role"}
	ErrSelfCannotBeDeleted      = translator.Translate{translator.RU: "Запрещено удалять свою учетную запись", translator.EN: "It is forbidden to delete your account"}

	ErrUserCannotCreateDomain = translator.Translate{translator.RU: "Нет прав на создание доменного администратора", translator.EN: "You do not have permission to create a domain administrator"}
	ErrLocalCannotBeDomain    = translator.Translate{translator.RU: "Локального администратора нельзя сделать доменным", translator.EN: "A local admin cannot be made domain"}
	ErrDomainCannotBeLocal    = translator.Translate{translator.RU: "Доменного администратора нельзя сделать локальным", translator.EN: "A domain admin cannot be made local"}
	ErrDomainLoginChange      = translator.Translate{translator.RU: "Нельзя изменить логин доменного администратора", translator.EN: "Cannot change login for domain admin"}
	ErrDomainPasswordChange   = translator.Translate{translator.RU: "Нельзя изменить пароль доменного администратора", translator.EN: "Cannot change password for domain admin"}
)

var (
	ErrRoleNotFound        = translator.Translate{translator.RU: "Роль не найдена", translator.EN: "Role not found"}
	ErrRoleExists          = translator.Translate{translator.RU: "Роль с таким именем уже существует", translator.EN: "A role with this name is already exists"}
	ErrRoleCannotBeChanged = translator.Translate{translator.RU: "Запрещено редактировать системные роли", translator.EN: "System roles cannot be changed"}
	ErrRoleCannotBeDeleted = translator.Translate{translator.RU: "Запрещено удалять системные роли", translator.EN: "System roles cannot be deleted"}
)

var (
	ErrUserBlockedNotFound = translator.Translate{translator.RU: "Запись не найдена", translator.EN: "Record not found"}
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
