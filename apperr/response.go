package apperr

import (
	"encoding/json"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/status"
	"io"
	"reflect"
	"strconv"
	"strings"
	"unicode"
)

func init() {
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterTagNameFunc(RegisterTagNameFunc)
		SetTranslators(v)
	}
}

type ValidateError struct {
	Column string `json:"column"`
	Error  string `json:"error"`
}

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
func HTTPResponse(c *gin.Context, err error) {
	var syntaxError *json.SyntaxError
	var unmarshalTypeError *json.UnmarshalTypeError
	var invalidUnmarshalError *json.InvalidUnmarshalError

	var appErr *APPError
	if errors.As(err, &appErr) {
		appErr = appErr.WithText(appErr.Translate.TranslateHttp(c))
		_ = c.Error(appErr).SetType(gin.ErrorTypePrivate)
		switch {
		case errors.As(appErr.Err, &syntaxError), errors.As(appErr.Err, &unmarshalTypeError), errors.As(appErr.Err, &invalidUnmarshalError):
			c.AbortWithStatusJSON(ErrSyntax.Status.HTTP(),
				ErrSyntax.
					WithContext(appErr.Context).
					WithShowMessageBanner(appErr.ShowMessageBanner).
					WithTitle(appErr.Title).
					WithText(ErrSyntax.Translate.TranslateHttp(c)),
			)
			return
		case Is(err, ErrValidation):
			errors.As(err, &ErrValidation)

			var errs []ValidateError
			if ErrValidation.Err != nil {
				var validateErrors validator.ValidationErrors
				errors.As(ErrValidation.Err, &validateErrors)
				for s, v := range validateErrors.Translate(getTranslatorHTTP(c)) {
					errs = append(errs, ValidateError{Column: getNamespace(s), Error: v})
				}
			}
			c.AbortWithStatusJSON(ErrValidation.Status.HTTP(), gin.H{
				"id":                  appErr.ID,
				"title":               appErr.Title,
				"text":                ErrValidation.Translate.TranslateHttp(c),
				"context":             appErr.Context,
				"show_message_banner": appErr.ShowMessageBanner,
				"errors":              errs,
			})
			return

		case errors.Is(appErr.Err, io.EOF), errors.Is(appErr.Err, io.ErrUnexpectedEOF), errors.Is(appErr.Err, io.ErrNoProgress):
			c.AbortWithStatusJSON(ErrEmptyData.Status.HTTP(),
				ErrEmptyData.
					WithContext(appErr.Context).
					WithShowMessageBanner(appErr.ShowMessageBanner).
					WithTitle(appErr.Title).
					WithText(ErrEmptyData.Translate.TranslateHttp(c)),
			)
			return
		default:
			c.AbortWithStatusJSON(appErr.Status.HTTP(), appErr.WithText(appErr.Translate.TranslateHttp(c)))
			return
		}
	} else {
		_ = c.Error(err).SetType(gin.ErrorTypePrivate)
		c.AbortWithStatusJSON(ErrInternal.Status.HTTP(), err.Error())
	}
}

func GRPCResponse(err error) error {
	br := &errdetails.BadRequest{}
	var appError *APPError
	var appErr *APPError
	if errors.As(err, &appErr) {
		switch {
		case Is(err, ErrValidation):
			errors.As(err, &ErrValidation)

			errs := []ValidateError{}
			if ErrValidation.Err != nil {
				var validateErrors validator.ValidationErrors
				errors.As(ErrValidation.Err, &validateErrors)
				for s, v := range validateErrors.Translate(getTranslator("ru")) {
					errs = append(errs, ValidateError{Column: getNamespace(s), Error: v})
				}
			}

			errConvert, _ := json.Marshal(errs)

			st := status.New(ErrValidation.Status.GRPC(), appErr.ID)
			v := []*errdetails.BadRequest_FieldViolation{
				{Field: "id", Description: appErr.ID},
				{Field: "title", Description: appErr.Title},
				{Field: "text", Description: ErrValidation.Translate.Translate("ru")},
				{Field: "context", Description: appErr.Context},
				{Field: "show_message_banner", Description: strconv.FormatBool(appErr.ShowMessageBanner)},
				{Field: "errors", Description: string(errConvert)},
			}
			br.FieldViolations = append(br.FieldViolations, v...)
			st, _ = st.WithDetails(br)

			return st.Err()
		default:
			st := status.New(appError.Status.GRPC(), appErr.ID)
			v := []*errdetails.BadRequest_FieldViolation{
				{Field: "id", Description: appErr.ID},
				{Field: "title", Description: appErr.Title},
				{Field: "text", Description: appError.Translate.Translate("ru")},
				{Field: "context", Description: appErr.Context},
				{Field: "show_message_banner", Description: strconv.FormatBool(appErr.ShowMessageBanner)},
			}
			br.FieldViolations = append(br.FieldViolations, v...)
			st, _ = st.WithDetails(br)
			return st.Err()
		}
	} else {
		st := status.New(ErrInternal.Status.GRPC(), ErrInternal.ID)
		v := []*errdetails.BadRequest_FieldViolation{
			{Field: "id", Description: ErrInternal.ID},
			{Field: "title", Description: ErrInternal.Title},
			{Field: "text", Description: ErrInternal.Translate.Translate("ru")},
			{Field: "context", Description: ErrInternal.Context},
			{Field: "show_message_banner", Description: strconv.FormatBool(ErrInternal.ShowMessageBanner)},
		}
		br.FieldViolations = append(br.FieldViolations, v...)
		st, _ = st.WithDetails(br)
		return st.Err()
	}
}
