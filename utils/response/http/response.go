package http

import (
	"encoding/json"
	"errors"
	"github.com/c2pc/go-pkg/v2/utils/apperr"
	"github.com/c2pc/go-pkg/v2/utils/translator"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"io"
	"strings"
	"unicode"
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
	var syntaxError *json.SyntaxError
	var unmarshalTypeError *json.UnmarshalTypeError
	var invalidUnmarshalError *json.InvalidUnmarshalError
	var validationError validator.ValidationErrors

	_ = c.Error(err).SetType(gin.ErrorTypePrivate)

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

			text := apperr.Translate(appError, GetTranslate(c))

			errs := []ValidateError{}
			for s, v := range validationError.Translate(getTranslator(c)) {
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

			c.AbortWithStatusJSON(CodeToHttp(appError.Code), gin.H{
				"id":     appError.ID,
				"text":   text,
				"errors": errs,
			})

			return
		}
	}

	text := apperr.Translate(appError, GetTranslate(c))

	c.AbortWithStatusJSON(CodeToHttp(appError.Code), gin.H{
		"id":   appError.ID,
		"text": text,
	})
}
