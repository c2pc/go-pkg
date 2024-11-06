package i18n

import "github.com/c2pc/go-pkg/v2/utils/translator"

var (
	ErrUserNotFound             = translator.Translate{translator.RU: "Пользователь не найден", translator.EN: "User not found"}
	ErrUserExists               = translator.Translate{translator.RU: "Пользователь с таким логином уже зарегистрирован", translator.EN: "A user with this login is already registered"}
	ErrUserRolesCannotBeChanged = translator.Translate{translator.RU: "Нельзя назначать пользователю другие роли", translator.EN: "User roles cannot be changed"}
	ErrUserCannotBeDeleted      = translator.Translate{translator.RU: "Пользователя нельзя удалять", translator.EN: "User cannot be deleted"}
)

var (
	ErrRoleNotFound        = translator.Translate{translator.RU: "Роль не найдена", translator.EN: "Role not found"}
	ErrRoleExists          = translator.Translate{translator.RU: "Роль уже добавлена", translator.EN: "Role has already been added"}
	ErrRoleCannotBeChanged = translator.Translate{translator.RU: "Роль нельзя редактировать", translator.EN: "Role cannot be changed"}
	ErrRoleCannotBeDeleted = translator.Translate{translator.RU: "Роль нельзя удалять", translator.EN: "Role cannot be deleted"}
)
