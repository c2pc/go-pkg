package apperr

import (
	"errors"
	"fmt"
	"github.com/c2pc/go-pkg/v2/utils/apperr/code"
	"github.com/c2pc/go-pkg/v2/utils/translator"
	"strings"
)

type Error struct {
	ID                string
	Code              code.Code
	Text              string
	TextTranslate     translator.Translate
	TextTranslateArgs []interface{}
	Err               error
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

func (e Error) WithErrorText(err string) Error {
	e.Err = errors.New(err)
	return e
}

func (e Error) WithTextArgs(args ...any) Error {
	e.TextTranslateArgs = args
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

func (e Error) GetIDPrefix() string {
	ids := strings.Split(e.ID, ".")
	if len(ids) == 0 {
		return e.ID
	}
	return ids[0]
}

func Translate(err Error, lang string) string {
	text := err.TextTranslate.Translate(lang, err.TextTranslateArgs...)

	if text == "" {
		text = err.Text
	}

	return text
}
