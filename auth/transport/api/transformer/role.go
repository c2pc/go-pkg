package transformer

import (
	"github.com/c2pc/go-pkg/v2/auth/model"
	model2 "github.com/c2pc/go-pkg/v2/utils/model"
	"github.com/c2pc/go-pkg/v2/utils/transformer"
	"github.com/gin-gonic/gin"
)

type SimpleRoleTransformer struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

func SimpleRoleTransform(m *model.Role) *SimpleRoleTransformer {
	return &SimpleRoleTransformer{
		ID:   m.ID,
		Name: m.Name,
	}
}

type RoleTransformer struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Read  []int  `json:"read"`
	Write []int  `json:"write"`
	Exec  []int  `json:"exec"`
}

func RoleTransform(m *model.Role) *RoleTransformer {
	r := &RoleTransformer{
		ID:    m.ID,
		Name:  m.Name,
		Read:  []int{},
		Write: []int{},
		Exec:  []int{},
	}

	r.Write, r.Read, r.Exec = getRolePermissions(m)

	return r
}

type RoleListTransformer struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Read  []int  `json:"read"`
	Write []int  `json:"write"`
	Exec  []int  `json:"exec"`
}

func RoleListTransform(c *gin.Context, p *model2.Pagination[model.Role]) (r []RoleListTransformer) {
	transformer.PaginationTransform(c, p)

	for _, m := range p.Rows {
		t := RoleListTransformer{
			ID:   m.ID,
			Name: m.Name,
		}
		t.Write, t.Read, t.Exec = getRolePermissions(&m)
		r = append(r, t)
	}

	return r
}

func getRolePermissions(m *model.Role) ([]int, []int, []int) {
	write, read, exec := make([]int, 0), make([]int, 0), make([]int, 0)

	if m.RolePermissions != nil {
		for _, perm := range m.RolePermissions {
			permID := perm.PermissionID
			if perm.Write {
				write = append(write, permID)
			}
			if perm.Read {
				read = append(read, permID)
			}
			if perm.Exec {
				exec = append(exec, permID)
			}
		}
	}

	return write, read, exec
}
