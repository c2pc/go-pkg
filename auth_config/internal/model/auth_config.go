package model

import "encoding/json"

type AuthConfig struct {
	Key   string
	Value json.RawMessage `gorm:"type:jsonb"`
}

func (s AuthConfig) TableName() string {
	return "auth_configs"
}

type CleanerConfig struct {
	HistoryAccDays int `json:"history_acc_days"`
	MaxRows        int `json:"max_rows"`
}
