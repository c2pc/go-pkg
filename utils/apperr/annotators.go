package apperr

import (
	"github.com/c2pc/go-pkg/v2/utils/apperr/code"
	"github.com/c2pc/go-pkg/v2/utils/translator"
)

type Annotator func(*Error)

func WithID(id string) Annotator {
	return func(err *Error) {
		if id == "" {
			return
		}
		err.ID = id
	}
}

func WithTextTranslate(tr translator.Translate) Annotator {
	return func(err *Error) {
		if tr == nil {
			return
		}
		err.TextTranslate = tr
	}
}

func WithCode(code code.Code) Annotator {
	return func(err *Error) {
		err.Code = code
	}
}

func WithText(text string) Annotator {
	return func(err *Error) {
		err.Text = text
	}
}
