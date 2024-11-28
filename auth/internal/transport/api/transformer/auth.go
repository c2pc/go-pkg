package transformer

import (
	"strings"

	model2 "github.com/c2pc/go-pkg/v2/auth/internal/model"
	"github.com/c2pc/go-pkg/v2/utils/transformer"
)

type AuthTokenTransformer struct {
	Token        string  `json:"accessToken"`
	RefreshToken string  `json:"refreshToken"`
	ExpiresAt    float64 `json:"expires"`
	TokenType    string  `json:"token_type"`
	UserID       int     `json:"user_id"`

	User *AuthAccountTransformer `json:"user"`
}

func AuthTokenTransform(m *model2.AuthToken) *AuthTokenTransformer {
	return &AuthTokenTransformer{
		Token:        m.Auth.Token,
		RefreshToken: m.Auth.RefreshToken,
		ExpiresAt:    m.Auth.ExpiresAt,
		TokenType:    m.Auth.TokenType,
		UserID:       m.Auth.UserID,
		User:         AuthAccountTransform(&m.User),
	}
}

type AuthAccountTransformer struct {
	ID         int     `json:"id"`
	Login      string  `json:"login"`
	FirstName  string  `json:"first_name"`
	SecondName *string `json:"second_name"`
	LastName   *string `json:"last_name"`
	Email      *string `json:"email"`
	Phone      *string `json:"phone"`

	Roles []*RoleTransformer `json:"roles"`
}

func AuthAccountTransform(m *model2.User) *AuthAccountTransformer {
	r := &AuthAccountTransformer{
		ID:         m.ID,
		Login:      strings.ToLower(m.Login),
		FirstName:  m.FirstName,
		SecondName: m.SecondName,
		LastName:   m.LastName,
		Email:      m.Email,
		Phone:      m.Phone,
		Roles:      transformer.Array(m.Roles, RoleWithNameTransform),
	}

	return r
}