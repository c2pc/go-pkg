package model

import (
	"encoding/json"
	"time"

	"github.com/c2pc/go-pkg/v2/auth/configurator"
)

const (
	AuthConfigKey  = "auth"
	AuditConfigKey = "audit"
	LogConfigKey   = "log"
)

const (
	SAMLMetadataConfigFileKey = "saml_metadata"
	SAMLCertConfigFileKey     = "saml_cert"
	SAMLKeyConfigFileKey      = "saml_key"
)

type Config struct {
	Key       string          `json:"key"`
	Value     json.RawMessage `json:"value" gorm:"type:jsonb"`
	UpdatedAt time.Time       `json:"updated_at" gorm:"autoUpdateTime"`

	Files []ConfigFile `json:"files" gorm:"foreignKey:ConfigKey;references:Key"`
}

func (s Config) TableName() string {
	return "auth_configs"
}

func (s Config) GetKey() string {
	return s.Key
}

func (s Config) GetValue() []byte {
	return s.Value
}

func (s Config) GetFiles() []configurator.File {
	files := make([]configurator.File, len(s.Files))
	for i, f := range s.Files {
		files[i] = f
	}

	return files
}

type ConfigFile struct {
	Key       string    `json:"key"`
	ConfigKey string    `json:"config_key"`
	Value     []byte    `json:"value"`
	FileName  string    `json:"file_name"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`

	Config *Config `json:"config" gorm:"foreignKey:ConfigKey;references:Key"`
}

func (s ConfigFile) TableName() string {
	return "auth_config_files"
}

func (s ConfigFile) GetKey() string {
	return s.Key
}

func (s ConfigFile) GetData() []byte {
	return s.Value
}

func (s ConfigFile) GetName() string {
	return s.FileName
}
