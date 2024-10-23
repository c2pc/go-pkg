package apperr

import (
	"errors"
	"fmt"
	"strings"

	"github.com/c2pc/go-pkg/apperr/utils/code"
	"github.com/c2pc/go-pkg/apperr/utils/translate"
)

type Error struct {
	ID                 string
	Context            string
	ShowMessage        bool
	Code               code.Code
	Text               string
	TextTranslate      translate.Translate
	TextTranslateArgs  []interface{}
	Title              string
	TitleTranslate     translate.Translate
	TitleTranslateArgs []interface{}
	Err                error
}

func New(id string, annotators ...Annotator) Error {
	err := Error{}

	WithID(id)(&err)

	for _, f := range annotators {
		f(&err)
	}

	return err
}

func Replace(err Error, annotators ...Annotator) Error {
	for _, f := range annotators {
		f(&err)
	}

	return err
}

func (e Error) Error() string {
	if e.Err == nil {
		return e.ID
	}

	var appError Error
	if errors.As(e.Err, &appError) {
		return e.ID + "." + e.Err.Error()
	}

	return e.ID + "." + fmt.Sprintf("(%s)", e.Err.Error())
}

func (e Error) LastError() error {
	var appError Error
	if errors.As(e.Err, &appError) {
		return appError.LastError()
	}

	return e.Err
}

func (e Error) WithError(err error) Error {
	e.Err = err
	return e
}

func (e Error) WithTextArgs(args ...any) Error {
	e.TextTranslateArgs = args
	return e
}

func (e Error) WithTitleArgs(args ...any) Error {
	e.TitleTranslateArgs = args
	return e
}

func (e Error) NewID(id string) Error {
	e.ID = id
	return e
}

func Is(err, target error) bool {
	var appError Error
	var appTarget Error
	if errors.As(err, &appError) && errors.As(target, &appTarget) {
		return appError.ID == appTarget.ID
	}

	return errors.Is(err, target)
}

func (e Error) GetIDSuffix() string {
	ids := strings.Split(e.ID, ".")
	if len(ids) == 0 {
		return e.ID
	}
	return ids[len(ids)-1]
}

func Translate(err Error, lang string) (string, string) {
	title := err.TitleTranslate.Translate(lang, err.TitleTranslateArgs...)
	text := err.TextTranslate.Translate(lang, err.TextTranslateArgs...)

	if text == "" {
		text = err.Text
	}

	return title, text
}

func Unwrap(err Error) Error {
	unwrappedError := err
	var lastError Error

	if err.Err != nil {
		if r, ok := err.Err.(Error); ok {
			lastError = r
		} else {
			return unwrappedError
		}
	}

	unwrappedError.Text = lastError.Text
	unwrappedError.TextTranslate = lastError.TextTranslate
	unwrappedError.TextTranslateArgs = lastError.TextTranslateArgs
	unwrappedError.Code = lastError.Code
	unwrappedError.ShowMessage = lastError.ShowMessage

	if lastError.ID == "" {
		unwrappedError.ID = err.ID
	} else {
		unwrappedError.ID = err.ID + "." + lastError.ID
	}

	return unwrappedError
}
