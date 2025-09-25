package model

import (
	"encoding/json"
	"time"
)

type AuthConfig struct {
	Key       string          `json:"key"`
	Value     json.RawMessage `json:"value" gorm:"type:jsonb"`
	UpdatedAt time.Time       `json:"updated_at" gorm:"autoUpdateTime"`
}

func (s AuthConfig) TableName() string {
	return "auth_configs"
}
