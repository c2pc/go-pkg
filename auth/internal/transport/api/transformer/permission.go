package transformer

import (
	"github.com/c2pc/go-pkg/v2/auth/internal/model"
	"github.com/c2pc/go-pkg/v2/utils/response/http"
	"github.com/gin-gonic/gin"
)

type PermissionTransformer struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Desc string `json:"desc"`
}

func PermissionTransform(c *gin.Context, m *model.Permission) *PermissionTransformer {
	return &PermissionTransformer{
		ID:   m.ID,
		Name: m.Name,
		Desc: model.GetPermission(m.Name).Translate(http.GetTranslate(c)),
	}
}

type PermissionListTransformer struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Desc string `json:"desc"`
}

func PermissionListTransform(c *gin.Context, p []model.Permission) (r []PermissionListTransformer) {
	for _, m := range p {
		r = append(r, PermissionListTransformer{
			ID:   m.ID,
			Name: m.Name,
			Desc: model.GetPermission(m.Name).Translate(http.GetTranslate(c)),
		})
	}
	return r
}
