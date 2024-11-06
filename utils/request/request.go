package request

import (
	"encoding/json"
	"io"

	"github.com/c2pc/go-pkg/v2/utils/apperr"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

func BindJSON[T any](c *gin.Context) (*T, error) {
	var r T
	if err := c.ShouldBindBodyWith(&r, binding.JSON); err != nil {
		return nil, apperr.ErrValidation.WithError(err)
	}
	return &r, nil
}

func BindJSONValue(c *gin.Context, r interface{}) error {
	if err := c.ShouldBindBodyWith(&r, binding.JSON); err != nil {
		return apperr.ErrValidation.WithError(err)
	}
	return nil
}

func BindQuery[T any](c *gin.Context) (*T, error) {
	var r T
	if err := c.ShouldBind(&r); err != nil {
		return nil, apperr.ErrValidation.WithError(err)
	}
	return &r, nil
}

func Bind[T any](c *gin.Context) (*T, error) {
	var r T
	if err := c.ShouldBind(&r); err != nil {
		return nil, apperr.ErrValidation.WithError(err)
	}
	return &r, nil
}

func BindUri[T any](c *gin.Context) (*T, error) {
	var r T
	if err := c.ShouldBindUri(&r); err != nil {
		return nil, apperr.ErrValidation.WithError(err)
	}
	return &r, nil
}

func DecodeJSON(r io.Reader, obj any) error {
	decoder := json.NewDecoder(r)
	if binding.EnableDecoderUseNumber {
		decoder.UseNumber()
	}
	if binding.EnableDecoderDisallowUnknownFields {
		decoder.DisallowUnknownFields()
	}
	if err := decoder.Decode(obj); err != nil {
		return apperr.ErrValidation.WithError(err)
	}
	err := binding.Validator.ValidateStruct(obj)
	if err != nil {
		return apperr.ErrValidation.WithError(err)
	}
	return nil
}

func BindJsonStruct(obj any) error {
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		err := v.Struct(obj)
		if err != nil {
			return apperr.ErrValidation.WithError(err)
		}
	}
	return nil
}
