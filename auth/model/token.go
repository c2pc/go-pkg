package model

import "time"

type RefreshToken struct {
	UserID    int       `json:"user_id"`
	DeviceID  int       `json:"device_id"`
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`

	User *User `json:"user"`
}

func (m RefreshToken) TableName() string {
	return "tokens"
}
