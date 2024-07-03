package transformer

import (
	"github.com/c2pc/go-pkg/v2/auth/model"
	model2 "github.com/c2pc/go-pkg/v2/utils/model"
)

type PermissionTransformer struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

func PermissionTransform(m *model.Permission) *PermissionTransformer {
	return &PermissionTransformer{
		ID:   m.ID,
		Name: m.Name,
	}
}

type PermissionListTransformer struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

func PermissionListTransform(p *model2.Pagination[model.Permission]) (r []PermissionListTransformer) {
	for _, m := range p.Rows {
		r = append(r, PermissionListTransformer{
			ID:   m.ID,
			Name: m.Name,
		})
	}
	return r
}
