package i18n

import "github.com/c2pc/go-pkg/v2/utils/translator"

var (
	ErrProfileNotFound = translator.Translate{translator.RU: "Профиль не найден", translator.EN: "Profile not found"}
	ErrProfileExists   = translator.Translate{translator.RU: "Профиль с таким логином уже зарегистрирован", translator.EN: "A profile with this login is already registered"}
)
