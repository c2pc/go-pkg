package apperr

import (
	"errors"
	"fmt"
	"strings"
)

type APPError struct {
	ID                string `json:"id"`
	Title             string `json:"title"`
	Text              string `json:"text"`
	Context           string `json:"context"`
	ShowMessageBanner bool   `json:"show_message_banner"`

	MethodID           string        `json:"-"`
	TextTranslate      Translate     `json:"-"`
	TextTranslateArgs  []interface{} `json:"-"`
	TitleTranslate     Translate     `json:"-"`
	TitleTranslateArgs []interface{} `json:"-"`
	Status             Code          `json:"-"`
	Err                error         `json:"-"`
}

func New(status Code, id string) *APPError {
	return &APPError{Status: status, ID: id}
}

func NewMethod(methodID string, context string) *APPError {
	return &APPError{MethodID: methodID, Context: context}
}

func (e APPError) Combine(err *APPError) *APPError {
	if e.MethodID == "" {
		e.ID = err.ID
	} else if strings.Contains(e.ID, ".") {
		i := strings.Index(e.ID, ".")
		e.ID = err.MethodID + err.ID[i:]
	} else {
		e.ID = err.ID
	}

	if e.TextTranslate == nil {
		e.TextTranslate = err.TextTranslate
		e.TextTranslateArgs = err.TextTranslateArgs
	}

	if e.TitleTranslate == nil {
		e.TitleTranslate = err.TitleTranslate
		e.TitleTranslateArgs = err.TitleTranslateArgs
	}

	e.Status = err.Status
	e.Err = err.Err

	return &e
}

func (e APPError) WithError(err error) *APPError {
	e.Err = err
	return &e
}

func (e APPError) WithStatus(status Code) *APPError {
	e.Status = status
	return &e
}

func (e APPError) WithTextTranslate(translate Translate) *APPError {
	e.TextTranslate = translate
	return &e
}

func (e APPError) WithTextTranslateArgs(args ...interface{}) *APPError {
	e.TextTranslateArgs = args
	return &e
}

func (e APPError) WithTitleTranslate(translate Translate) *APPError {
	e.TitleTranslate = translate
	return &e
}

func (e APPError) WithTitleTranslateArgs(args ...interface{}) *APPError {
	e.TitleTranslateArgs = args
	return &e
}

func (e APPError) WithShowMessageBanner(showMessageBanner bool) *APPError {
	e.ShowMessageBanner = showMessageBanner
	return &e
}

func (e APPError) Translate(lang string) *APPError {
	e.Title = e.TitleTranslate.Translate(lang, e.TitleTranslateArgs...)
	e.Text = e.TextTranslate.Translate(lang, e.TextTranslateArgs...)
	return &e
}

func (e APPError) Error() string {
	err := e.Translate("ru")
	if e.Err != nil {
		return fmt.Sprintf("%s -- %s -- %s -- %s \n %s", err.Context, err.ID, err.Title, err.Text, err.Err.Error())
	}
	return fmt.Sprintf("%s -- %s -- %s -- %s", err.Context, err.ID, err.Title, err.Text)
}

func Is(err, target error) bool {
	var appError *APPError
	var appTarget *APPError
	if errors.As(err, &appError) && errors.As(target, &appTarget) {
		return appError.Status == appTarget.Status && appError.ID == appTarget.ID
	}

	if errors.As(err, &appError) && target != nil {
		return appError.ID == target.Error()
	}

	if errors.As(target, &appTarget) && err != nil {
		return appTarget.ID == err.Error()
	}

	return errors.Is(err, target)
}
