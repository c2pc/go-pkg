package transformer

import (
	"github.com/c2pc/go-pkg/v2/auth/model"
	"github.com/c2pc/go-pkg/v2/auth/profile"
	model2 "github.com/c2pc/go-pkg/v2/utils/model"
	"github.com/c2pc/go-pkg/v2/utils/transformer"
	"github.com/gin-gonic/gin"
)

type UserTransformer struct {
	ID         int     `json:"id"`
	Login      string  `json:"login"`
	FirstName  string  `json:"first_name"`
	SecondName *string `json:"second_name"`
	LastName   *string `json:"last_name"`
	Email      *string `json:"email"`
	Phone      *string `json:"phone"`
	Blocked    bool    `json:"blocked"`

	Roles   []*SimpleRoleTransformer `json:"roles"`
	Profile interface{}              `json:"profile,omitempty"`
}

func UserTransform[Model any](m *model.User, profileTransformer profile.ITransformer[Model]) *UserTransformer {
	r := &UserTransformer{
		ID:         m.ID,
		Login:      m.Login,
		FirstName:  m.FirstName,
		SecondName: m.SecondName,
		LastName:   m.LastName,
		Email:      m.Email,
		Phone:      m.Phone,
		Blocked:    m.Blocked,
		Roles:      transformer.Array(m.Roles, SimpleRoleTransform),
	}

	if profileTransformer != nil && m.Profile != nil {
		if prof, ok := m.Profile.(*Model); ok {
			r.Profile = profileTransformer.TransformProfile(prof)
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

	Roles   []*SimpleRoleTransformer `json:"roles"`
	Profile interface{}              `json:"profile,omitempty"`
}

func UserListTransform[Model any](c *gin.Context, p *model2.Pagination[model.User], profileTransformer profile.ITransformer[Model]) []UserListTransformer {
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
			Roles:      transformer.Array(m.Roles, SimpleRoleTransform),
		}

		if profileTransformer != nil && m.Profile != nil {
			if prof, ok := m.Profile.(*Model); ok {
				user.Profile = profileTransformer.TransformProfile(prof)
			}
		}

		r = append(r, user)
	}

	return r
}
