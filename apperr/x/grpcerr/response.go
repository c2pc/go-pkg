package grpcerr

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/c2pc/go-pkg/apperr"
	"github.com/c2pc/go-pkg/apperr/utils/appErrors"
	"github.com/c2pc/go-pkg/apperr/utils/translator"
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

func Response(ctx context.Context, err apperr.Apperr) error {
	var syntaxError *json.SyntaxError
	var unmarshalTypeError *json.UnmarshalTypeError
	var invalidUnmarshalError *json.InvalidUnmarshalError
	var validationError validator.ValidationErrors
	br := &errdetails.BadRequest{}

	var appErr apperr.Apperr
	if !errors.As(err.GetErr(), &appErr) {
		childError := err.GetErr()

		switch {
		case errors.As(childError, &syntaxError), errors.As(childError, &unmarshalTypeError), errors.As(childError, &invalidUnmarshalError):
			err = appErrors.ErrSyntax.WithError(childError)
		case errors.Is(childError, io.EOF), errors.Is(childError, io.ErrUnexpectedEOF), errors.Is(childError, io.ErrNoProgress):
			err = appErrors.ErrEmptyData.WithError(childError)
		case errors.As(childError, &validationError):
			appErr := apperr.Unwrap(appErrors.ErrValidation.WithError(childError))
			title, text := apperr.Translate(appErr, GetTranslate(ctx))

			errs := []ValidateError{}
			for s, v := range validationError.Translate(getTranslator(ctx)) {
				column := getNamespace(s)
				columnError := strings.ReplaceAll(v, column, "")
				columnError = strings.ReplaceAll(v, "  ", " ")
				errs = append(errs, ValidateError{Column: column, Error: columnError})
			}

			errConvert, _ := json.Marshal(errs)

			st := status.New(codeToGrpc(appErr.GetCode()), appErr.Error())
			v := []*errdetails.BadRequest_FieldViolation{
				{Field: "id", Description: appErr.GetID()},
				{Field: "title", Description: title},
				{Field: "text", Description: text},
				{Field: "context", Description: appErr.GetContext()},
				{Field: "show_message_banner", Description: strconv.FormatBool(appErr.GetShowMessage())},
				{Field: "errors", Description: string(errConvert)},
			}
			br.FieldViolations = append(br.FieldViolations, v...)
			st, _ = st.WithDetails(br)

			return st.Err()
		default:
			err = appErrors.ErrInternal.WithError(childError)
		}
	}

	appErr = apperr.Unwrap(err)
	title, text := apperr.Translate(appErr, GetTranslate(ctx))

	st := status.New(codeToGrpc(appErr.GetCode()), appErr.Error())
	v := []*errdetails.BadRequest_FieldViolation{
		{Field: "id", Description: appErr.GetID()},
		{Field: "title", Description: title},
		{Field: "text", Description: text},
		{Field: "context", Description: appErr.GetContext()},
		{Field: "show_message_banner", Description: strconv.FormatBool(appErr.GetShowMessage())},
	}
	br.FieldViolations = append(br.FieldViolations, v...)
	st, _ = st.WithDetails(br)

	return st.Err()
}
