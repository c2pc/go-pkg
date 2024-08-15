package transformer

import (
	"github.com/c2pc/go-pkg/v2/auth/model"
	"github.com/c2pc/go-pkg/v2/auth/profile"
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

func AuthTokenTransform[Model any](m *model.AuthToken, profileTransformer profile.ITransformer[Model]) *AuthTokenTransformer {
	return &AuthTokenTransformer{
		Token:        m.Auth.Token,
		RefreshToken: m.Auth.RefreshToken,
		ExpiresAt:    m.Auth.ExpiresAt,
		TokenType:    m.Auth.TokenType,
		UserID:       m.Auth.UserID,
		User:         AuthAccountTransform(&m.User, profileTransformer),
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

	Roles   []*RoleTransformer `json:"roles"`
	Profile interface{}        `json:"profile,omitempty"`
}

func AuthAccountTransform[Model any](m *model.User, profileTransformer profile.ITransformer[Model]) *AuthAccountTransformer {
	r := &AuthAccountTransformer{
		ID:         m.ID,
		Login:      m.Login,
		FirstName:  m.FirstName,
		SecondName: m.SecondName,
		LastName:   m.LastName,
		Email:      m.Email,
		Phone:      m.Phone,
		Roles:      transformer.Array(m.Roles, RoleWithNameTransform),
	}

	if profileTransformer != nil && m.Profile != nil {
		if prof, ok := m.Profile.(*Model); ok {
			r.Profile = profileTransformer.TransformProfile(prof)
		}
	}

	return r
}
