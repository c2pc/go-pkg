package transformer

import (
	"github.com/c2pc/go-pkg/v2/auth/model"
	model2 "github.com/c2pc/go-pkg/v2/utils/model"
)

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

func RoleListTransform(p *model2.Pagination[model.Role]) (r []RoleListTransformer) {
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

func getRolePermissions(m *model.Role) (write, read, exec []int) {
	if m.RolePermissions != nil {
		for _, perm := range m.RolePermissions {
			if perm.Write {
				write = append(write, perm.PermissionID)
			}
			if perm.Read {
				read = append(read, perm.PermissionID)
			}
			if perm.Exec {
				exec = append(exec, perm.PermissionID)
			}
		}
	}

	return write, read, exec
}
