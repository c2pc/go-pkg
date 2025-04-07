package model

import "time"

type RefreshToken struct {
	ID        int       `json:"id"`
	UserID    int       `json:"user_id"`
	DeviceID  int       `json:"device_id"`
	Token     string    `json:"token"`
	LoggedAt  time.Time `json:"logged_at"`
	UpdatedAt time.Time `json:"updated_at"`
	ExpiresAt time.Time `json:"expires_at"`
	Provider  *string   `json:"provider"`

	User *User `json:"user"`
}

func (m RefreshToken) TableName() string {
	return "auth_tokens"
}
