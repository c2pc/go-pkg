package request

import (
	"github.com/c2pc/go-pkg/v2/utils/apperr"
	"github.com/c2pc/go-pkg/v2/utils/apperr/code"
	"github.com/c2pc/go-pkg/v2/utils/translator"
	"github.com/gin-gonic/gin"
	"strings"
)

var (
	ErrorValidationId = apperr.New("validation_id",
		apperr.WithTextTranslate(translator.Translate{
			translator.RU: "Неверный ID",
			translator.EN: "Invalid ID",
		}),
		apperr.WithCode(code.InvalidArgument),
	)
	ErrorValidationUUID = apperr.New("validation_uuid",
		apperr.WithTextTranslate(translator.Translate{
			translator.RU: "Неверный uuid",
			translator.EN: "Invalid UUID",
		}),
		apperr.WithCode(code.InvalidArgument),
	)
)

type IdRequest struct {
	Id int `uri:"id" binding:"required"`
}

func Id(c *gin.Context) (int, error) {
	r, err := BindUri[IdRequest](c)
	if err != nil {
		return 0, ErrorValidationId.WithError(err)
	}
	return r.Id, nil
}

type UUIDRequest struct {
	UUID string `uri:"uuid" binding:"required"`
}

func UUID(c *gin.Context) (string, error) {
	r, err := BindUri[UUIDRequest](c)
	if err != nil {
		return "", ErrorValidationUUID.WithError(err)
	}

	uu := strings.Split(r.UUID, ".")
	if len(uu) > 1 {
		return uu[0], nil
	}

	return r.UUID, nil
}
