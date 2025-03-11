package transformer

import (
	"github.com/c2pc/go-pkg/v2/auth/internal/model"
	model2 "github.com/c2pc/go-pkg/v2/utils/model"
	"github.com/c2pc/go-pkg/v2/utils/transformer"
	"github.com/gin-gonic/gin"
)

type FilterTransformer struct {
	ID       int    `json:"id"`
	Endpoint string `json:"endpoint"`
	Name     string `json:"name"`
	Value    string `json:"value"`
}

func FilterTransform(m *model.Filter) *FilterTransformer {
	r := &FilterTransformer{
		ID:       m.ID,
		Endpoint: m.Endpoint,
		Name:     m.Name,
		Value:    string(m.Value),
	}

	return r
}

type FilterListTransformer struct {
	ID       int    `json:"id"`
	Endpoint string `json:"endpoint"`
	Name     string `json:"name"`
	Value    string `json:"value"`
}

func FilterListTransform(c *gin.Context, p *model2.Pagination[model.Filter]) []FilterListTransformer {
	transformer.PaginationTransform(c, p)

	r := make([]FilterListTransformer, 0)

	for _, m := range p.Rows {
		user := FilterListTransformer{
			ID:       m.ID,
			Endpoint: m.Endpoint,
			Name:     m.Name,
			Value:    string(m.Value),
		}

		r = append(r, user)
	}

	return r
}
