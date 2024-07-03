package transformer

import (
	"github.com/c2pc/go-pkg/v2/auth/model"
	model2 "github.com/c2pc/go-pkg/v2/utils/model"
	"github.com/c2pc/go-pkg/v2/utils/transformer"
)

type UserTransformer struct {
	ID         int     `json:"id"`
	Login      string  `json:"login"`
	FirstName  string  `json:"name"`
	SecondName *string `json:"second_name"`
	LastName   *string `json:"last_name"`
	Email      *string `json:"email"`
	Phone      *string `json:"phone"`

	Roles []*RoleTransformer `json:"roles"`
}

func UserTransform(m *model.User) *UserTransformer {
	r := &UserTransformer{
		ID:         m.ID,
		Login:      m.Login,
		FirstName:  m.FirstName,
		SecondName: m.SecondName,
		LastName:   m.LastName,
		Email:      m.Email,
		Phone:      m.Phone,
		Roles:      transformer.NillableArray(m.Roles, RoleTransform),
	}

	return r
}

type UserListTransformer struct {
	ID         int     `json:"id"`
	Login      string  `json:"login"`
	FirstName  string  `json:"name"`
	SecondName *string `json:"second_name"`
	LastName   *string `json:"last_name"`
	Email      *string `json:"email"`
	Phone      *string `json:"phone"`

	Roles []*RoleTransformer `json:"roles"`
}

func UserListTransform(p *model2.Pagination[model.User]) (r []UserListTransformer) {
	for _, m := range p.Rows {
		r = append(r, UserListTransformer{
			ID:         m.ID,
			Login:      m.Login,
			FirstName:  m.FirstName,
			SecondName: m.SecondName,
			LastName:   m.LastName,
			Email:      m.Email,
			Phone:      m.Phone,
			Roles:      transformer.NillableArray(m.Roles, RoleTransform),
		})
	}

	return r
}
