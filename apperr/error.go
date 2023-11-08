package apperr

import (
	"errors"
	"github.com/c2pc/go-pkg/apperr/utils/code"
	"github.com/c2pc/go-pkg/apperr/utils/translate"
	"strings"
)

type Apperr interface {
	Error() string
	WithError(err error) Apperr
	WithTextArgs(args ...any) Apperr
	WithTitleArgs(args ...any) Apperr
	NewID(id string) Apperr
	GetIDSuffix() string
	GetID() string
	GetContext() string
	GetShowMessage() bool
	GetCode() code.Code
	GetTextTranslate() translate.Translate
	GetTextTranslateArgs() []interface{}
	GetTitleTranslate() translate.Translate
	GetTitleTranslateArgs() []interface{}
	GetErr() error
	SetID(id string) Apperr
	SetContext(context string) Apperr
	SetShowMessage(showMessage bool) Apperr
	SetTextTranslate(tr translate.Translate) Apperr
	SetTitleTranslate(tr translate.Translate) Apperr
	SetCode(code code.Code) Apperr
	SetTextTranslateArgs(args ...any) Apperr
	SetTitleTranslateArgs(args ...any) Apperr
	SetErr(err error) Apperr
	SetText(text string) Apperr
	SetTitle(title string) Apperr
}

type Error struct {
	ID                 string
	Context            string
	ShowMessage        bool
	Text               string
	Code               code.Code
	TextTranslate      translate.Translate
	TextTranslateArgs  []interface{}
	Title              string
	TitleTranslate     translate.Translate
	TitleTranslateArgs []interface{}
	Err                error
}

func New(id string, annotators ...Annotator) Apperr {
	err := WithID(id)(&Error{})

	for _, f := range annotators {
		err = f(err)
	}

	return err
}

func Copy(err Apperr, annotators ...Annotator) Apperr {
	for _, f := range annotators {
		err = f(err)
	}

	return err
}

func (e *Error) Error() string {
	if e.Err == nil {
		return ""
	}

	var appError Apperr
	if errors.As(e.Err, &appError) {
		return e.ID + "." + appError.Error()
	}

	return e.ID + "." + strings.ReplaceAll(e.Err.Error(), " ", "_")
}

func (e *Error) WithError(err error) Apperr {
	e.Err = err
	return e
}

func (e *Error) WithTextArgs(args ...any) Apperr {
	e.TextTranslateArgs = args
	return e
}

func (e *Error) WithTitleArgs(args ...any) Apperr {
	e.TitleTranslateArgs = args
	return e
}

func (e *Error) NewID(id string) Apperr {
	e.ID = id
	return e
}

func (e *Error) GetIDSuffix() string {
	ids := strings.Split(e.ID, ".")
	if len(ids) == 0 {
		return e.ID
	}
	return ids[len(ids)-1]
}

func (e *Error) GetID() string {
	return e.ID
}

func (e *Error) GetContext() string {
	return e.Context
}

func (e *Error) GetShowMessage() bool {
	return e.ShowMessage
}

func (e *Error) GetCode() code.Code {
	return e.Code
}

func (e *Error) GetTextTranslate() translate.Translate {
	return e.TextTranslate
}

func (e *Error) GetTextTranslateArgs() []interface{} {
	return e.TextTranslateArgs
}

func (e *Error) GetTitleTranslate() translate.Translate {
	return e.TitleTranslate
}

func (e *Error) GetTitleTranslateArgs() []interface{} {
	return e.TitleTranslateArgs
}

func (e *Error) GetErr() error {
	return e.Err
}

func (e *Error) SetID(id string) Apperr {
	e.ID = id
	return e
}

func (e *Error) SetContext(context string) Apperr {
	e.Context = context
	return e
}

func (e *Error) SetShowMessage(showMessage bool) Apperr {
	e.ShowMessage = showMessage
	return e
}

func (e *Error) SetCode(code code.Code) Apperr {
	e.Code = code
	return e
}

func (e *Error) SetTextTranslate(tr translate.Translate) Apperr {
	e.TextTranslate = tr
	return e
}

func (e *Error) SetTextTranslateArgs(args ...any) Apperr {
	e.TextTranslateArgs = args
	return e
}

func (e *Error) SetTitleTranslate(tr translate.Translate) Apperr {
	e.TitleTranslate = tr
	return e
}

func (e *Error) SetTitleTranslateArgs(args ...any) Apperr {
	e.TitleTranslateArgs = args
	return e
}

func (e *Error) SetErr(err error) Apperr {
	e.Err = err
	return e
}

func (e *Error) SetText(text string) Apperr {
	e.Title = text
	return e
}

func (e *Error) SetTitle(title string) Apperr {
	e.Text = title
	return e
}

func Is(err, target error) bool {
	var appError Apperr
	var appTarget Apperr
	if errors.As(err, &appError) && errors.As(target, &appTarget) {
		return appError.GetID() == appTarget.GetID()
	}

	return errors.Is(err, target)
}

func Translate(err Apperr, lang string) (string, string) {
	return err.GetTitleTranslate().Translate(lang, err.GetTitleTranslateArgs()...),
		err.GetTextTranslate().Translate(lang, err.GetTextTranslateArgs()...)
}

func Unwrap(err Apperr) Apperr {
	unwrappedError := err

	lastError := err
	if err.GetErr() != nil {
		var r Apperr
		if errors.As(err.GetErr(), &r) {
			lastError = r
		}
	}

	if Is(err, lastError) {
		return err
	}

	var id string
	if lastError.GetID() == "" {
		id = err.GetID()
	} else {
		id = err.GetID() + "." + lastError.GetID()
	}

	return Copy(unwrappedError,
		WithTextTranslate(lastError.GetTextTranslate()),
		WithCode(lastError.GetCode()),
		WithShowMessage(lastError.GetShowMessage()),
		WithID(id),
		WithTextTranslateArgs(lastError.GetTextTranslateArgs()),
	)
}
