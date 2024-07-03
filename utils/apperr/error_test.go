package apperr

import (
	"errors"
	"github.com/c2pc/go-pkg/v2/utils/apperr/code"
	"github.com/c2pc/go-pkg/v2/utils/translator"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNew(t *testing.T) {
	err := New("id")

	assert.Equal(t, "id", err.ID)
}

func TestWithError(t *testing.T) {
	err := New("id").WithError(errors.New("new err"))
	assert.Equal(t, "id.(new err)", err.Error())

	err2 := New("id2").WithError(err)
	assert.Equal(t, "id2.id.(new err)", err2.Error())

	err3 := New("id3").WithError(err2)
	assert.Equal(t, "id3.id2.id.(new err)", err3.Error())
}

func TestWithID(t *testing.T) {
	err := New("id", WithID("id2"))

	assert.Equal(t, "id2", err.ID)
}

func TestWithTextTranslate(t *testing.T) {
	textTranslate := translator.Translate{translator.RU: "text"}
	err := New("id", WithTextTranslate(textTranslate))

	assert.Equal(t, textTranslate, err.TextTranslate)
}

func TestWithTextTranslateArgs(t *testing.T) {
	textTranslate := translator.Translate{translator.RU: "text"}
	textTranslateArgs := []interface{}{1, 2, 3}
	err := New("id", WithTextTranslate(textTranslate)).WithTextArgs(textTranslateArgs...)

	assert.Equal(t, textTranslateArgs, err.TextTranslateArgs)
}

func TestIs(t *testing.T) {
	err1 := New("id")
	err11 := New("id")
	err2 := New("id2")
	err21 := New("id2")
	err3 := errors.New("error3")
	err4 := errors.New("error4")

	assert.True(t, Is(err1, err11))
	assert.True(t, Is(err2, err21))
	assert.True(t, Is(err3, err3))
	assert.True(t, Is(err4, err4))

	assert.False(t, Is(err1, err2))
	assert.False(t, Is(err2, err1))
	assert.False(t, Is(err3, err4))
	assert.False(t, Is(err4, err3))
	assert.False(t, Is(err1, err3))
	assert.False(t, Is(err4, err2))

	assert.False(t, Is(err1, nil))
	assert.False(t, Is(nil, err1))

	assert.False(t, Is(err3, nil))
	assert.False(t, Is(nil, err3))
}

func TestWithCode(t *testing.T) {
	err := New("id", WithCode(code.Aborted))

	assert.Equal(t, code.Aborted, err.Code)
}
