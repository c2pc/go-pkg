package apperr

import (
	"errors"
	"fmt"
	"net/http"
	"testing"
)

func TestNew(t *testing.T) {
	appErr := New(http.StatusUnauthorized, "err")

	if appErr.Status != http.StatusUnauthorized {
		t.Errorf("status must be %d", http.StatusUnauthorized)
	}

	if appErr.ID != "err" {
		t.Errorf("id must be %s", "err")
	}
}

func TestWithError(t *testing.T) {
	appErr := New(http.StatusUnauthorized, "err").WithError(errors.New("error"))

	if appErr.Err == nil {
		t.Errorf("error cannot be nil")
	}

	if errors.Is(appErr.Err, errors.New("error")) {
		t.Errorf("error must be %s", "error")
	}
}

func TestWithContext(t *testing.T) {
	appErr := New(http.StatusUnauthorized, "err").WithContext("context")

	if appErr.Context != "context" {
		t.Errorf("context must be %s", "context")
	}
}

func TestWithTranslate(t *testing.T) {
	appErr := New(http.StatusUnauthorized, "err").WithTranslate(Translate{"ru": "RU"})

	if appErr.Translate == nil {
		t.Errorf("translate cannot be nil")
	}

	if _, found := appErr.Translate["ru"]; !found {
		t.Errorf("translate not found")
	}
}

func TestWithShowMessageBanner(t *testing.T) {
	appErr := New(http.StatusUnauthorized, "err").WithShowMessageBanner(true)

	if appErr.ShowMessageBanner == false {
		t.Errorf("showMessageBanner must be true")
	}
}

func TestWithTitle(t *testing.T) {
	appErr := New(http.StatusUnauthorized, "err").WithTitle("title")

	if appErr.Title != "title" {
		t.Errorf("title must be %s", "title")
	}
}

func TestWithText(t *testing.T) {
	appErr := New(http.StatusUnauthorized, "err").WithText("text")

	if appErr.Text != "text" {
		t.Errorf("text must be %s", "text")
	}
}

func TestError(t *testing.T) {
	appErr := New(http.StatusUnauthorized, "err").
		WithContext("context").
		WithTitle("title").
		WithText("text")

	errText := appErr.Error()
	mustBeErr := fmt.Sprintf("%s -- %s -- %s -- %s", "context", "err", "title", "text")
	if errText != mustBeErr {
		t.Errorf("error must be - %s", mustBeErr)
	}

	appErr = appErr.WithError(errors.New("error"))

	errText = appErr.Error()
	mustBeErr = fmt.Sprintf("%s -- %s -- %s -- %s \n %s", "context", "err", "title", "text", errors.New("error").Error())
	if errText != mustBeErr {
		t.Errorf("error must be - %s", mustBeErr)
	}
}

func TestIs(t *testing.T) {
	appErr := New(http.StatusUnauthorized, "err")
	appErr2 := New(http.StatusUnauthorized, "err")

	if !Is(appErr, appErr2) {
		t.Errorf("the errors must be equal")
	}

	appErr2 = New(http.StatusUnauthorized, "err2")
	if Is(appErr, appErr2) {
		t.Errorf("the errors cannot be equal")
	}

	appErr2 = New(http.StatusNotFound, "err")
	if Is(appErr, appErr2) {
		t.Errorf("the errors cannot be equal")
	}

	err := errors.New("err")
	if !Is(err, err) {
		t.Errorf("the errors must be equal")
	}

	err2 := errors.New("err")
	if Is(err, err2) {
		t.Errorf("the errors cannot be equal")
	}

	if !Is(appErr, err) {
		t.Errorf("the errors must be equal")
	}

	if !Is(err, appErr) {
		t.Errorf("the errors must be equal")
	}

	appErr2 = New(http.StatusNotFound, "err2")
	if Is(appErr2, err) {
		t.Errorf("the errors cannot be equal")
	}

	if Is(err, appErr2) {
		t.Errorf("the errors cannot be equal")
	}
}
