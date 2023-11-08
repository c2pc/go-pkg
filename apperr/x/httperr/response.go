package httperr

import (
	"encoding/json"
	"errors"
	"github.com/c2pc/go-pkg/apperr"
	"github.com/c2pc/go-pkg/apperr/utils/appErrors"
	"github.com/c2pc/go-pkg/apperr/utils/translator"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"io"
	"reflect"
	"strings"
	"unicode"
)

func init() {
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterTagNameFunc(RegisterTagNameFunc)
		translator.SetValidateTranslators(v)
	}
}

func RegisterTagNameFunc(fld reflect.StructField) string {
	fieldName := fld.Tag.Get("key")
	if fieldName == "-" {
		return ""
	}
	return fieldName
}

func getNamespace(str string) string {
	touch := strings.Index(str, ".")
	if touch != -1 {
		str = str[strings.Index(str, ".")+1:]
	}

	newStr := ""
	for i, s := range str {
		if unicode.IsUpper(s) {
			if i != 0 && string(str[i-1]) != "." {
				newStr += "_" + strings.ToLower(string(s))
			} else {
				newStr += strings.ToLower(string(s))
			}
			continue
		}
		newStr += string(s)
	}
	return newStr
}

type ValidateError struct {
	Column string `json:"column"`
	Error  string `json:"error"`
}

func Response(c *gin.Context, err apperr.Error) {
	var syntaxError *json.SyntaxError
	var unmarshalTypeError *json.UnmarshalTypeError
	var invalidUnmarshalError *json.InvalidUnmarshalError
	var validationError validator.ValidationErrors

	var childError apperr.Error
	if !errors.As(err.Err, &childError) {
		switch {
		case errors.As(childError, &syntaxError), errors.As(childError, &unmarshalTypeError), errors.As(childError, &invalidUnmarshalError):
			err = appErrors.ErrSyntax.WithError(childError)
		case errors.Is(childError, io.EOF), errors.Is(childError, io.ErrUnexpectedEOF), errors.Is(childError, io.ErrNoProgress):
			err = appErrors.ErrEmptyData.WithError(childError)
		case errors.As(childError, &validationError):
			appErr := apperr.Unwrap(appErrors.ErrValidation.WithError(childError))
			title, text := apperr.Translate(appErr, GetTranslate(c))

			errs := []ValidateError{}
			for s, v := range validationError.Translate(getTranslator(c)) {
				column := getNamespace(s)
				columnError := strings.ReplaceAll(v, column, "")
				columnError = strings.ReplaceAll(v, "  ", " ")
				errs = append(errs, ValidateError{Column: column, Error: columnError})
			}

			c.AbortWithStatusJSON(codeToHttp(appErr.Code), gin.H{
				"id":                  appErr.ID,
				"title":               title,
				"text":                text,
				"context":             appErr.Context,
				"show_message_banner": appErr.ShowMessage,
				"errors":              errs,
			})

			return
		}
	}

	appErr := apperr.Unwrap(err)
	title, text := apperr.Translate(appErr, GetTranslate(c))

	c.AbortWithStatusJSON(codeToHttp(appErr.Code), gin.H{
		"id":                  appErr.ID,
		"title":               title,
		"text":                text,
		"context":             appErr.Context,
		"show_message_banner": appErr.ShowMessage,
	})
}
