package http

import (
	"encoding/json"
	"errors"
	"io"
	"strings"
	"unicode"

	"github.com/c2pc/go-pkg/v2/utils/apperr"
	"github.com/c2pc/go-pkg/v2/utils/translator"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
)

func init() {
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		translator.SetValidateTranslators(v)
	}
}

func removePrefix(str string) string {
	touch := strings.LastIndex(str, ".")
	if touch != -1 {
		return str[touch+1:]
	}
	return str
}

func removeSuffix(str string) string {
	touch := strings.Index(str, ".")
	if touch != -1 {
		return str[touch+1:]
	}
	return str
}

func getNamespace(str string) string {
	str = removeSuffix(str)
	newStr := ""
	for i, s := range str {
		if unicode.IsUpper(s) {
			if i != 0 && string(str[i-1]) != "." && string(str[i-1]) != strings.ToUpper(string(str[i-1])) {
				newStr += "_" + strings.ToLower(string(s))
			} else {
				newStr += strings.ToLower(string(s))
			}
			continue
		}
		newStr += string(s)
	}

	touch := strings.LastIndex(newStr, "].")
	if touch != -1 {
		return newStr[touch+2:]
	}

	return newStr
}

type ValidateError struct {
	Column string `json:"column"`
	Error  string `json:"error"`
}

func Response(c *gin.Context, err error) {
	_ = c.Error(err).SetType(gin.ErrorTypePrivate)

	response := UnwrapError(c, err, "")
	c.AbortWithStatusJSON(response.Code, response)
}

type ErrorResponse struct {
	Code   int         `json:"-"`
	ID     string      `json:"id"`
	Text   string      `json:"text"`
	Errors interface{} `json:"errors,omitempty"`
}

func UnwrapError(c *gin.Context, err error, lang string) ErrorResponse {
	var syntaxError *json.SyntaxError
	var unmarshalTypeError *json.UnmarshalTypeError
	var invalidUnmarshalError *json.InvalidUnmarshalError
	var validationError validator.ValidationErrors
	var importError apperr.ImportError

	var appError apperr.Error
	if !errors.As(err, &appError) {
		appError = apperr.ErrInternal.WithError(err)
	}

	var childError apperr.Error
	lastError := appError.LastError()
	if !errors.As(lastError, &childError) {
		switch {
		case errors.As(lastError, &syntaxError), errors.As(lastError, &unmarshalTypeError), errors.As(lastError, &invalidUnmarshalError):
			err = appError.WithError(apperr.ErrSyntax.WithError(lastError))

		case errors.Is(lastError, io.EOF), errors.Is(lastError, io.ErrUnexpectedEOF), errors.Is(lastError, io.ErrNoProgress):
			err = appError.WithError(apperr.ErrEmptyData.WithError(lastError))

		case errors.As(lastError, &validationError):
			err = appError.WithError(apperr.ErrValidation.WithError(lastError))

			var text string
			if lang != "" {
				text = apperr.Translate(appError, lang)
			} else {
				text = apperr.Translate(appError, GetTranslate(c))
			}

			errs := []ValidateError{}

			var tr ut.Translator
			if lang != "" {
				tr = getTranslatorByLang(lang)
			} else {
				tr = getTranslator(c)
			}

			for s, v := range validationError.Translate(tr) {
				str := removePrefix(s)
				ers := strings.Split(v, "failed on")
				columnError := strings.ReplaceAll(ers[len(ers)-1], str+" ", "")
				if len(ers) > 1 {
					columnError = "failed on" + columnError
				}
				columnError = strings.ReplaceAll(columnError, "  ", " ")
				columnError = strings.ToLower(columnError)

				column := getNamespace(s)
				errs = append(errs, ValidateError{Column: column, Error: columnError})
			}

			return ErrorResponse{
				Code:   CodeToHttp(appError.Code),
				ID:     appError.ID,
				Text:   text,
				Errors: errs,
			}
		case errors.As(lastError, &importError):
			err = appError.WithError(apperr.ErrValidation.WithError(lastError))

			var text string
			if lang != "" {
				text = apperr.Translate(appError, lang)
			} else {
				text = apperr.Translate(appError, GetTranslate(c))
			}

			return ErrorResponse{
				Code:   CodeToHttp(appError.Code),
				ID:     appError.ID,
				Text:   text,
				Errors: importError.Errors,
			}
		}
	}

	var text string
	if lang != "" {
		text = apperr.Translate(appError, lang)
	} else {
		text = apperr.Translate(appError, GetTranslate(c))
	}

	return ErrorResponse{
		Code: CodeToHttp(appError.Code),
		ID:   appError.ID,
		Text: text,
	}
}
