package dto

import (
	"strconv"

	"github.com/c2pc/go-pkg/v2/example/internal/service"
	"github.com/c2pc/go-pkg/v2/example/internal/transport/api/request"
	request2 "github.com/c2pc/go-pkg/v2/utils/request"
)

func NewsCreate(input *request.NewsCreateRequest) service.NewsCreateInput {
	return service.NewsCreateInput{
		Title:   input.Title,
		Content: input.Content,
	}
}

func NewsUpdate(input *request.NewsUpdateRequest) service.NewsUpdateInput {
	return service.NewsUpdateInput{
		Title:   input.Title,
		Content: input.Content,
	}
}

func NewsMassUpdate(input request2.MultipleUpdateRequest[request.NewsMassUpdateRequest]) service.NewsMassUpdateInput {
	return service.NewsMassUpdateInput{
		IDs:     input.IDs,
		Content: input.Data.Content,
	}
}

func NewsImport(input []request.NewsImportRequest, errs map[string]string) service.NewsImportInput {
	var data []service.NewsImportDataInput

	for i, v := range input {
		if e, ok := errs[strconv.Itoa(i)]; ok {
			data = append(data, service.NewsImportDataInput{
				Err: &e,
			})
		} else {
			data = append(data, service.NewsImportDataInput{
				Title:   v.Title,
				Content: v.Content,
			})
		}
	}

	return service.NewsImportInput{
		Data: data,
	}
}
