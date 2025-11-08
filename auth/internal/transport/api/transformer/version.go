package transformer

import (
	"github.com/c2pc/go-pkg/v2/auth/internal/model"
)

type VersionTransformer struct {
	AppName string `json:"app_name"`
	App     string `json:"app"`
	DB      string `json:"db"`
}

func VersionTransform(m *model.Version) *VersionTransformer {
	r := &VersionTransformer{
		AppName: m.AppName,
		App:     m.App,
		DB:      m.DB,
	}

	return r
}
