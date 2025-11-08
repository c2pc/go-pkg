package transformer

import (
	"github.com/c2pc/go-pkg/v2/auth/internal/model"
	"github.com/c2pc/go-pkg/v2/utils/meta"
	"github.com/c2pc/go-pkg/v2/utils/transformer"
	"github.com/gin-gonic/gin"
)

type SimpleRoleTransformer struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	IsSystem bool   `json:"is_system"`
}

func SimpleRoleTransform(m *model.Role) *SimpleRoleTransformer {
	return &SimpleRoleTransformer{
		ID:       m.ID,
		Name:     m.Name,
		IsSystem: model.IsRole(m.Name, model.SuperAdmin),
	}
}

type RoleTransformer struct {
	ID       int           `json:"id"`
	Name     string        `json:"name"`
	IsSystem bool          `json:"is_system"`
	Read     []interface{} `json:"read"`
	Write    []interface{} `json:"write"`
	Exec     []interface{} `json:"exec"`
}

func RoleTransform(m *model.Role) *RoleTransformer {
	r := &RoleTransformer{
		ID:       m.ID,
		Name:     m.Name,
		IsSystem: model.IsRole(m.Name, model.SuperAdmin),
		Read:     []interface{}{},
		Write:    []interface{}{},
		Exec:     []interface{}{},
	}

	r.Write, r.Read, r.Exec = getRolePermissions(m, true)

	return r
}

func RoleWithNameTransform(m *model.Role) *RoleTransformer {
	r := &RoleTransformer{
		ID:       m.ID,
		Name:     m.Name,
		IsSystem: model.IsRole(m.Name, model.SuperAdmin),
		Read:     []interface{}{},
		Write:    []interface{}{},
		Exec:     []interface{}{},
	}

	r.Write, r.Read, r.Exec = getRolePermissions(m, false)

	return r
}

type RoleListTransformer struct {
	ID       int           `json:"id"`
	Name     string        `json:"name"`
	IsSystem bool          `json:"is_system"`
	Read     []interface{} `json:"read"`
	Write    []interface{} `json:"write"`
	Exec     []interface{} `json:"exec"`
}

func RoleListTransform(c *gin.Context, p *meta.Pagination[model.Role]) []RoleListTransformer {
	transformer.PaginationTransform(c, p)

	r := make([]RoleListTransformer, 0)

	for _, m := range p.Rows {
		t := RoleListTransformer{
			ID:       m.ID,
			Name:     m.Name,
			IsSystem: model.IsRole(m.Name, model.SuperAdmin),
		}
		t.Write, t.Read, t.Exec = getRolePermissions(&m, true)
		r = append(r, t)
	}

	return r
}

func UserRoleListTransform(c *gin.Context, p *meta.Pagination[model.UserRole]) []UserListTransformer {
	transformer.PaginationTransform(c, p)

	r := make([]UserListTransformer, 0)

	for _, m := range p.Rows {
		if m.User != nil {
			user := UserListTransformer{
				ID:         m.User.ID,
				Login:      m.User.Login,
				FirstName:  m.User.FirstName,
				SecondName: m.User.SecondName,
				LastName:   m.User.LastName,
				Email:      m.User.Email,
				Blocked:    m.User.Blocked,
				IsDomain:   m.User.IsDomain,
				Roles:      transformer.Array(m.User.Roles, SimpleRoleTransform),
			}

			r = append(r, user)
		}
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
