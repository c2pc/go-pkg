package transformer

import (
	"github.com/c2pc/go-pkg/v2/auth/internal/model"
)

type VersionTransformer struct {
	App string `json:"app"`
	DB  string `json:"db"`
}

func VersionTransform(m *model.Version) *VersionTransformer {
	r := &VersionTransformer{
		App: m.App,
		DB:  m.DB,
	}

	return r
}
