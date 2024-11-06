package transformer

import (
	"github.com/c2pc/go-pkg/v2/example/internal/model"
	model2 "github.com/c2pc/go-pkg/v2/utils/model"
	"github.com/c2pc/go-pkg/v2/utils/transformer"
	"github.com/gin-gonic/gin"
)

type NewsTransformer struct {
	ID      int     `json:"id"`
	Title   string  `json:"title"`
	Content *string `json:"content"`
}

func NewsTransform(m *model.News) *NewsTransformer {
	r := &NewsTransformer{
		ID:      m.ID,
		Title:   m.Title,
		Content: m.Content,
	}

	return r
}

type NewsListTransformer struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
}

func NewsListTransform(c *gin.Context, p *model2.Pagination[model.News]) []NewsListTransformer {
	transformer.PaginationTransform(c, p)

	r := make([]NewsListTransformer, 0)

	for _, m := range p.Rows {
		user := NewsListTransformer{
			ID:    m.ID,
			Title: m.Title,
		}

		r = append(r, user)
	}

	return r
}
