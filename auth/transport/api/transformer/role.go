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
	ID    int           `json:"id"`
	Name  string        `json:"name"`
	Read  []interface{} `json:"read"`
	Write []interface{} `json:"write"`
	Exec  []interface{} `json:"exec"`
}

func RoleTransform(m *model.Role) *RoleTransformer {
	r := &RoleTransformer{
		ID:    m.ID,
		Name:  m.Name,
		Read:  []interface{}{},
		Write: []interface{}{},
		Exec:  []interface{}{},
	}

	r.Write, r.Read, r.Exec = getRolePermissions(m, true)

	return r
}

func RoleWithNameTransform(m *model.Role) *RoleTransformer {
	r := &RoleTransformer{
		ID:    m.ID,
		Name:  m.Name,
		Read:  []interface{}{},
		Write: []interface{}{},
		Exec:  []interface{}{},
	}

	r.Write, r.Read, r.Exec = getRolePermissions(m, false)

	return r
}

type RoleListTransformer struct {
	ID    int           `json:"id"`
	Name  string        `json:"name"`
	Read  []interface{} `json:"read"`
	Write []interface{} `json:"write"`
	Exec  []interface{} `json:"exec"`
}

func RoleListTransform(c *gin.Context, p *model2.Pagination[model.Role]) (r []RoleListTransformer) {
	transformer.PaginationTransform(c, p)

	for _, m := range p.Rows {
		t := RoleListTransformer{
			ID:   m.ID,
			Name: m.Name,
		}
		t.Write, t.Read, t.Exec = getRolePermissions(&m, true)
		r = append(r, t)
	}

	return r
}

func getRolePermissions(m *model.Role, getIDs bool) ([]interface{}, []interface{}, []interface{}) {
	write, read, exec := make([]interface{}, 0), make([]interface{}, 0), make([]interface{}, 0)

	if m.RolePermissions != nil {
		for _, perm := range m.RolePermissions {
			var permID interface{}
			if getIDs {
				permID = perm.PermissionID
			} else {
				permID = perm.Permission.Name
			}

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
