package transformer

import (
	"github.com/c2pc/go-pkg/v2/auth/internal/model"
)

type SettingTransformer struct {
	Settings *string `json:"settings"`
}

func SettingTransform(m *model.Setting) *SettingTransformer {
	return &SettingTransformer{
		Settings: m.Settings,
	}
}
