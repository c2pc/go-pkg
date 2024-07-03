package request

import (
	"github.com/c2pc/go-pkg/v2/utils/apperr"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
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
