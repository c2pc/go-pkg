package config

import (
	"embed"
	"github.com/c2pc/go-pkg/v2/utils/config"
	_ "github.com/c2pc/go-pkg/v2/utils/migration/config"
)

//go:embed migrations/*.yml
var fs embed.FS

func Migrate(path string) error {
	return config.Migrate(fs, "migrations", path)
}
