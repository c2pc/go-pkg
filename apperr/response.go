package apperr

import (
	"encoding/json"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"io"
	"reflect"
	"strings"
	"unicode"
)

func RegisterTagNameFunc(fld reflect.StructField) string {
	fieldName := fld.Tag.Get("key")
	if fieldName == "-" {
		return ""
	}
	return fieldName
}

// Add key to tag Email string `json="email" key="email_address"`
//
//	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
//		v.RegisterTagNameFunc(RegisterTagNameFunc)
//	}
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

// HTTPResponse writes an error response to client
func HTTPResponse(c *gin.Context, err *APPError) {
	_ = c.Error(err).SetType(gin.ErrorTypePrivate)

	var syntaxError *json.SyntaxError
	var unmarshalTypeError *json.UnmarshalTypeError
	var invalidUnmarshalError *json.InvalidUnmarshalError
	var appError *APPError

	switch {
	case errors.As(err.Err, &syntaxError), errors.As(err.Err, &unmarshalTypeError), errors.As(err.Err, &invalidUnmarshalError):
		c.AbortWithStatusJSON(ErrSyntax.Status,
			ErrSyntax.
				WithContext(err.Context).
				WithShowMessageBanner(err.ShowMessageBanner).
				WithTitle(err.Title).
				WithText(ErrSyntax.Translate.Translate(c)),
		)
	case errors.As(err, &ErrValidation):
		if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
			v.RegisterTagNameFunc(RegisterTagNameFunc)
		}

		errors.As(err, &ErrValidation)
		errs := map[string]string{}
		if ErrValidation.Err != nil {
			var validateErrors validator.ValidationErrors
			errors.As(ErrValidation.Err, &validateErrors)
			for s, v := range validateErrors.Translate(getTranslator(c)) {
				errs[getNamespace(s)] = v
			}
		}
		c.AbortWithStatusJSON(ErrValidation.Status, gin.H{
			"id":                  ErrValidation.ID,
			"title":               err.Title,
			"text":                ErrValidation.Translate.Translate(c),
			"context":             err.Context,
			"show_message_banner": err.ShowMessageBanner,
			"errors":              errs,
		})

	case errors.Is(err.Err, io.EOF), errors.Is(err.Err, io.ErrUnexpectedEOF), errors.Is(err.Err, io.ErrNoProgress):
		c.AbortWithStatusJSON(ErrEmptyData.Status,
			ErrEmptyData.
				WithContext(err.Context).
				WithShowMessageBanner(err.ShowMessageBanner).
				WithTitle(err.Title).
				WithText(ErrEmptyData.Translate.Translate(c)),
		)
	case errors.As(err, &appError):
		errors.As(err, &appError)
		c.AbortWithStatusJSON(appError.Status, appError)
	default:
		c.AbortWithStatusJSON(ErrInternal.Status,
			ErrInternal.
				WithContext(err.Context).
				WithShowMessageBanner(err.ShowMessageBanner).
				WithTitle(err.Title).
				WithText(ErrInternal.Translate.Translate(c)),
		)
	}
}
