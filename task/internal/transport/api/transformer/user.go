package transformer

import (
	"github.com/c2pc/go-pkg/v2/task/internal/model"
)

type UserSimpleTransformer struct {
	ID         int     `json:"id"`
	Login      string  `json:"login"`
	FirstName  string  `json:"first_name"`
	SecondName *string `json:"second_name"`
	LastName   *string `json:"last_name"`
}

func UserSimpleTransform(m *model.User) *UserSimpleTransformer {
	r := &UserSimpleTransformer{
		ID:         m.ID,
		Login:      m.Login,
		FirstName:  m.FirstName,
		SecondName: m.SecondName,
		LastName:   m.LastName,
	}

	return r
}
