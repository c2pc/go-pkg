package apperr

import (
	"github.com/c2pc/go-pkg/v2/utils/apperr/code"
	"github.com/c2pc/go-pkg/v2/utils/translator"
)

// Annotator - функция для аннотирования ошибки.
type Annotator func(*Error)

// WithID создает аннотатор, который устанавливает ID для ошибки.
func WithID(id string) Annotator {
	return func(err *Error) {
		if id != "" {
			err.ID = id
		}
	}
}

// WithTextTranslate создает аннотатор, который устанавливает переводчик для ошибки.
func WithTextTranslate(tr translator.Translator) Annotator {
	return func(err *Error) {
		if tr != nil {
			err.TextTranslate = tr
		}
	}
}

// WithCode создает аннотатор, который устанавливает код для ошибки.
func WithCode(code code.Code) Annotator {
	return func(err *Error) {
		err.Code = code
	}
}

// WithText создает аннотатор, который устанавливает текст для ошибки.
func WithText(text string) Annotator {
	return func(err *Error) {
		err.Text = text
	}
}
