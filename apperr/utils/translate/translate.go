package translate

import "fmt"

type Language string

const (
	RU Language = "ru"
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
