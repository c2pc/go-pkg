package apperr

import (
	"errors"
	"fmt"
	"strings"

	"github.com/c2pc/go-pkg/v2/utils/apperr/code"
	"github.com/c2pc/go-pkg/v2/utils/translator"
)

type Error struct {
	ID                string
	Code              code.Code
	Text              string
	TextTranslate     translator.Translator
	TextTranslateArgs []interface{}
	Err               error
}

// New создает новый экземпляр Error и применяет все аннотаторы.
func New(id string, annotators ...Annotator) Error {
	err := Error{}
	WithID(id)(&err)
	for _, annotator := range annotators {
		annotator(&err)
	}
	return err
}

// Replace заменяет существующую ошибку новыми аннотаторами.
func Replace(err Error, annotators ...Annotator) Error {
	for _, annotator := range annotators {
		annotator(&err)
	}
	return err
}

// Error возвращает строковое представление ошибки.
func (e Error) Error() string {
	if e.Err == nil {
		return e.ID
	}

	var appError Error
	if errors.As(e.Err, &appError) {
		return fmt.Sprintf("%s.%s", e.ID, e.Err.Error())
	}

	return fmt.Sprintf("%s.(%s)", e.ID, e.Err.Error())
}

// LastError возвращает последнюю ошибку из цепочки ошибок.
func (e Error) LastError() error {
	var appError Error
	if errors.As(e.Err, &appError) {
		return appError.LastError()
	}
	return e.Err
}

// WithError добавляет внутреннюю ошибку к Error.
func (e Error) WithError(err error) Error {
	e.Err = err
	return e
}

// WithErrorText создает новую ошибку из текстового сообщения.
func (e Error) WithErrorText(err string) Error {
	e.Err = errors.New(err)
	return e
}

// WithTextArgs задает аргументы для перевода текста.
func (e Error) WithTextArgs(args ...interface{}) Error {
	e.TextTranslateArgs = args
	return e
}

// NewID устанавливает новый ID для ошибки.
func (e Error) NewID(id string) Error {
	e.ID = id
	return e
}

// Is проверяет, является ли ошибка `err` такой же, как и `target`.
func Is(err, target error) bool {
	var appError, appTarget Error
	if errors.As(err, &appError) && errors.As(target, &appTarget) {
		return appError.ID == appTarget.ID
	}
	return errors.Is(err, target)
}

// GetIDPrefix возвращает префикс ID до первого разделителя.
func (e Error) GetIDPrefix() string {
	ids := strings.Split(e.ID, ".")
	if len(ids) > 0 {
		return ids[0]
	}
	return e.ID
}

// Translate возвращает переведенное сообщение ошибки или исходное сообщение, если перевод не найден.
func Translate(err error, lang string) string {
	if err == nil {
		err = ErrInternal
	}

	var appError Error
	if !errors.As(err, &appError) {
		err = ErrInternal.WithError(err)
	}

	text := appError.TextTranslate.Translate(lang, appError.TextTranslateArgs...)
	if text == "" {
		text = appError.Text
	}

	return text
}
