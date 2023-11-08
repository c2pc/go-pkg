package apperr

import (
	"github.com/c2pc/go-pkg/apperr/utils/code"
	"github.com/c2pc/go-pkg/apperr/utils/translate"
)

type Annotator func(Apperr) Apperr

func WithID(id string) Annotator {
	return func(err Apperr) Apperr {
		if id == "" {
			return err
		}
		return err.SetID(id)
	}
}

func WithContext(context string) Annotator {
	return func(err Apperr) Apperr {
		if context == "" {
			return err
		}
		return err.SetContext(context)
	}
}

func WithShowMessage(showMessage bool) Annotator {
	return func(err Apperr) Apperr {
		return err.SetShowMessage(showMessage)
	}
}

func WithTextTranslate(tr translate.Translate) Annotator {
	return func(err Apperr) Apperr {
		if tr == nil {
			return err
		}
		return err.SetTextTranslate(tr)
	}
}

func WithTitleTranslate(tr translate.Translate) Annotator {
	return func(err Apperr) Apperr {
		if tr == nil {
			return err
		}
		return err.SetTitleTranslate(tr)
	}
}

func WithCode(code code.Code) Annotator {
	return func(err Apperr) Apperr {
		return err.SetCode(code)
	}
}

func WithTextTranslateArgs(args ...any) Annotator {
	return func(err Apperr) Apperr {
		return err.SetTextTranslateArgs(args)
	}
}

func WithTitleTranslateArgs(args ...any) Annotator {
	return func(err Apperr) Apperr {
		return err.SetTitleTranslateArgs(args)
	}
}

func WithErr(r error) Annotator {
	return func(err Apperr) Apperr {
		return err.SetErr(r)
	}
}
