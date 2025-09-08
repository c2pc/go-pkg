package model

import "encoding/json"

type AuthConfig struct {
	Key   string
	Value json.RawMessage `gorm:"type:jsonb"`
}

func (s AuthConfig) TableName() string {
	return "auth_configs"
}
