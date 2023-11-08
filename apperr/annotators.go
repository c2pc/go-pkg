package apperr

import (
	"github.com/c2pc/go-pkg/apperr/utils/code"
	"github.com/c2pc/go-pkg/apperr/utils/translate"
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

func WithContext(context string) Annotator {
	return func(err *Error) {
		if context == "" {
			return
		}
		err.Context = context
	}
}

func WithShowMessage(showMessage bool) Annotator {
	return func(err *Error) {
		err.ShowMessage = showMessage
	}
}

func WithTextTranslate(tr translate.Translate) Annotator {
	return func(err *Error) {
		if tr == nil {
			return
		}
		err.TextTranslate = tr
	}
}

func WithTitleTranslate(tr translate.Translate) Annotator {
	return func(err *Error) {
		if tr == nil {
			return
		}
		err.TitleTranslate = tr
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

func WithTitle(title string) Annotator {
	return func(err *Error) {
		err.Title = title
	}
}
