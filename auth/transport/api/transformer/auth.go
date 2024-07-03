package transformer

import (
	"github.com/c2pc/go-pkg/v2/auth/model"
	"github.com/c2pc/go-pkg/v2/utils/transformer"
)

type AuthTokenTransformer struct {
	Token        string  `json:"token"`
	RefreshToken string  `json:"refresh_token"`
	ExpiresAt    float64 `json:"expires_at"`
	TokenType    string  `json:"token_type"`
	UserID       int     `json:"user_id"`

	User *UserTransformer `json:"user"`
}

func AuthTokenTransform(m *model.AuthToken) *AuthTokenTransformer {
	return &AuthTokenTransformer{
		Token:        m.Auth.Token,
		RefreshToken: m.Auth.RefreshToken,
		ExpiresAt:    m.Auth.ExpiresAt,
		TokenType:    m.Auth.TokenType,
		UserID:       m.Auth.UserID,
		User:         UserTransform(&m.User),
	}
}

type AuthAccountTransformer struct {
	ID         int     `json:"id"`
	Login      string  `json:"login"`
	FirstName  string  `json:"name"`
	SecondName *string `json:"second_name"`
	LastName   *string `json:"last_name"`
	Email      *string `json:"email"`
	Phone      *string `json:"phone"`

	Roles []*RoleTransformer `json:"roles"`
}

func AuthAccountTransform(m *model.User) *AuthAccountTransformer {
	r := &AuthAccountTransformer{
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
