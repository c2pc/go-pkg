package apperr

import (
	"errors"
	"fmt"
)

type APPError struct {
	ID                string `json:"id"`
	Title             string `json:"title"`
	Text              string `json:"text"`
	Context           string `json:"context"`
	ShowMessageBanner bool   `json:"show_message_banner"`

	MethodID      string        `json:"-"`
	Translate     Translate     `json:"-"`
	TranslateArgs []interface{} `json:"-"`
	Status        Code          `json:"-"`
	Err           error         `json:"-"`
}

func New(status Code, id string) *APPError {
	return &APPError{Status: status, ID: id}
}

func NewForResponse(methodID string, title string, context string) *APPError {
	return &APPError{MethodID: methodID, Title: title, Context: context}
}

func (e APPError) WithApperr(err *APPError) *APPError {
	if e.MethodID != "" {
		e.ID = e.MethodID + "." + err.ID
	} else {
		e.ID = err.ID
	}

	e.Text = err.Text
	e.ShowMessageBanner = err.ShowMessageBanner
	e.Translate = err.Translate
	e.TranslateArgs = err.TranslateArgs
	e.Status = err.Status
	e.Err = err.Err

	return &e
}

func (e APPError) WithError(err error) *APPError {
	e.Err = err
	return &e
}

func (e APPError) WithContext(context string) *APPError {
	e.Context = context
	return &e
}

func (e APPError) WithId(id string) *APPError {
	e.ID = id
	return &e
}

func (e APPError) WithStatus(status Code) *APPError {
	e.Status = status
	return &e
}

func (e APPError) WithTranslate(translate Translate) *APPError {
	e.Translate = translate
	return &e
}

func (e APPError) WithTranslateArgs(args ...interface{}) *APPError {
	e.TranslateArgs = args
	return &e
}

func (e APPError) WithShowMessageBanner(showMessageBanner bool) *APPError {
	e.ShowMessageBanner = showMessageBanner
	return &e
}

func (e APPError) WithTitle(title string) *APPError {
	e.Title = title
	return &e
}

func (e APPError) WithText(text string) *APPError {
	if text != "" {
		e.Text = text
	}
	return &e
}

func (e APPError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s -- %s -- %s -- %s \n %s", e.Context, e.ID, e.Title, e.Text, e.Err.Error())
	}
	return fmt.Sprintf("%s -- %s -- %s -- %s", e.Context, e.ID, e.Title, e.Text)
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
