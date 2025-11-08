package i18n

import "github.com/c2pc/go-pkg/v2/utils/translator"

var (
	RolesPermission        = translator.Translate{translator.RU: "Роли", translator.EN: "Roles"}
	UsersPermission        = translator.Translate{translator.RU: "Администраторы", translator.EN: "Administrators"}
	UsersBlockedPermission = translator.Translate{translator.RU: "Заблокированные администраторы", translator.EN: "Blocked Administrators"}
	SessionsPermission     = translator.Translate{translator.RU: "Активные сеансы", translator.EN: "Active sessions"}
	ConfigPermission       = translator.Translate{translator.RU: "Настройки конфигураций", translator.EN: "Config settings"}
	AnalyticPermission     = translator.Translate{translator.RU: "Журналы безопасности", translator.EN: "Security logs"}
	TaskPermission         = translator.Translate{translator.RU: "Задачи", translator.EN: "Tasks"}
)
