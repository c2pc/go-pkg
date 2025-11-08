package transformer

import (
	"encoding/json"
	"time"

	"github.com/c2pc/go-pkg/v2/auth/internal/model"
)

type ConfigTransformer struct {
	Value     json.RawMessage `json:"value"`
	UpdatedAt time.Time       `json:"updated_at"`
}

func ConfigTransform(m *model.Config) *ConfigTransformer {
	r := &ConfigTransformer{
		Value:     m.Value,
		UpdatedAt: m.UpdatedAt,
	}

	return r
}
