package transformer

import (
	"github.com/c2pc/go-pkg/v2/auth/internal/model"
	"github.com/c2pc/go-pkg/v2/auth/profile"
	model2 "github.com/c2pc/go-pkg/v2/utils/model"
	"github.com/c2pc/go-pkg/v2/utils/transformer"
	"github.com/gin-gonic/gin"
)

type UserSimpleTransformer struct {
	ID         int     `json:"id"`
	Login      string  `json:"login"`
	FirstName  string  `json:"first_name"`
	SecondName *string `json:"second_name"`
	LastName   *string `json:"last_name"`
	IsDomain   bool    `json:"is_domain"`
}

func UserSimpleTransform(m *model.User) *UserSimpleTransformer {
	r := &UserSimpleTransformer{
		ID:         m.ID,
		Login:      m.Login,
		FirstName:  m.FirstName,
		SecondName: m.SecondName,
		LastName:   m.LastName,
		IsDomain:   m.IsDomain,
	}

	return r
}

type UserTransformer struct {
	ID         int     `json:"id"`
	Login      string  `json:"login"`
	FirstName  string  `json:"first_name"`
	SecondName *string `json:"second_name"`
	LastName   *string `json:"last_name"`
	Email      *string `json:"email"`
	Phone      *string `json:"phone"`
	Blocked    bool    `json:"blocked"`
	IsDomain   bool    `json:"is_domain"`

	Roles   []*SimpleRoleTransformer `json:"roles"`
	Profile interface{}              `json:"profile,omitempty"`
}

func UserTransform(m *model.User, profileTransformer profile.ITransformer) *UserTransformer {
	r := &UserTransformer{
		ID:         m.ID,
		Login:      m.Login,
		FirstName:  m.FirstName,
		SecondName: m.SecondName,
		LastName:   m.LastName,
		Email:      m.Email,
		Phone:      m.Phone,
		Blocked:    m.Blocked,
		IsDomain:   m.IsDomain,
		Roles:      transformer.Array(m.Roles, SimpleRoleTransform),
	}

	if profileTransformer != nil && m.Profile != nil {
		if prof, ok := m.Profile.(*profile.IModel); ok {
			r.Profile = profileTransformer.Transform(prof)
		}
	}

	return r
}

type UserListTransformer struct {
	ID         int     `json:"id"`
	Login      string  `json:"login"`
	FirstName  string  `json:"first_name"`
	SecondName *string `json:"second_name"`
	LastName   *string `json:"last_name"`
	Email      *string `json:"email"`
	Phone      *string `json:"phone"`
	Blocked    bool    `json:"blocked"`
	IsDomain   bool    `json:"is_domain"`

	Roles   []*SimpleRoleTransformer `json:"roles"`
	Profile interface{}              `json:"profile,omitempty"`
}

func UserListTransform(c *gin.Context, p *model2.Pagination[model.User], profileTransformer profile.ITransformer) []UserListTransformer {
	transformer.PaginationTransform(c, p)

	r := make([]UserListTransformer, 0)

	for _, m := range p.Rows {
		user := UserListTransformer{
			ID:         m.ID,
			Login:      m.Login,
			FirstName:  m.FirstName,
			SecondName: m.SecondName,
			LastName:   m.LastName,
			Email:      m.Email,
			Phone:      m.Phone,
			Blocked:    m.Blocked,
			IsDomain:   m.IsDomain,
			Roles:      transformer.Array(m.Roles, SimpleRoleTransform),
		}

		if profileTransformer != nil && m.Profile != nil {
			if prof, ok := m.Profile.(*profile.IModel); ok {
				user.Profile = profileTransformer.Transform(prof)
			}
		}

		r = append(r, user)
	}

	return r
}
