package profile

import "github.com/c2pc/go-pkg/v2/utils/translator"

var (
	ErrNotFoundTranslate = translator.Translate{translator.RU: "Профиль не найден", translator.EN: "Profile not found"}
	ErrExistsTranslate   = translator.Translate{translator.RU: "Профиль уже зарегистрирован", translator.EN: "A profile is already registered"}
)
