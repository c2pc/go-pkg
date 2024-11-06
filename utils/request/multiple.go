package request

import (
	"mime/multipart"
	"strconv"
	"strings"

	"github.com/c2pc/go-pkg/v2/utils/csv"
	response "github.com/c2pc/go-pkg/v2/utils/response/http"
	"github.com/c2pc/go-pkg/v2/utils/translator"
	"github.com/gin-gonic/gin"
)

type MultipleUpdateRequest[T any] struct {
	IDs  []int `json:"ids" binding:"required,min=1,max=500000,unique"`
	Data T     `json:"data" binding:"required"`
}

type MultipleDeleteRequest struct {
	IDs []int `json:"ids" binding:"required,min=1,max=500000,unique"`
}

type ImportRequest struct {
	CSV *multipart.FileHeader `form:"csv" binding:"required"`
}

func BindImportFileRequest[T any](c *gin.Context) ([]T, map[string]string, error) {
	cred, err := Bind[ImportRequest](c)
	if err != nil {
		return nil, nil, err
	}

	data, err := csv.UnMarshalCSVFromFile[T](cred.CSV)
	if err != nil {
		return nil, nil, err
	}

	errs := make(map[string]string)
	var inputs []T
	for i, d := range data {
		err := BindJsonStruct(d)
		if err != nil {
			resp := response.UnwrapError(c, err, translator.EN.String())

			var text []string
			for _, e := range resp.Errors {
				text = append(text, e.Column+" - "+e.Error)
			}
			errs[strconv.Itoa(i)] = strings.Join(text, ", ")

			var t T
			inputs = append(inputs, t)
			continue
		}
		inputs = append(inputs, d)
	}

	return inputs, errs, nil
}
