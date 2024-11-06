package i18n

import "github.com/c2pc/go-pkg/v2/utils/translator"

var (
	ErrNewsNotFound    = translator.Translate{translator.RU: "Новость не найдена", translator.EN: "News not found"}
	ErrNewsListIsEmpty = translator.Translate{translator.RU: "Список новостей пуст", translator.EN: "News list is empty"}
	ErrNewsExists      = translator.Translate{translator.RU: "Новость с таким заголовком уже добавлена", translator.EN: "The news with this title is already created"}
)
