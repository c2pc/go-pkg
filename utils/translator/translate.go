package translator

import "fmt"

type Translator interface {
	Translate(acceptLang string, args ...any) string
}

type Language string

func (l Language) String() string {
	return string(l)
}

const (
	RU Language = "ru"
	EN Language = "en"
)

type Translate map[Language]string

func (t Translate) Translate(acceptLang string, args ...any) string {
	tr, found := t[Language(acceptLang)]
	if !found {
		tr, found = t[RU]
		if !found {
			return ""
		}
	}

	if len(args) > 0 {
		return fmt.Sprintf(tr, args...)
	}

	return tr
}
