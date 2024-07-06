package config

import (
	"embed"
	"errors"
	"github.com/c2pc/go-pkg/v2/utils/migrate"
	_ "github.com/c2pc/go-pkg/v2/utils/migration/config"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"os"
	"strings"
)

func Migrate(fs embed.FS, migrationsDir, migratePath string) error {
	d, err := iofs.New(fs, migrationsDir)
	if err != nil {
		return err
	}

	if !strings.Contains(migratePath, "/") && !strings.Contains(migratePath, "\\") {
		wd, err := os.Getwd()
		if err != nil {
			return err
		}
		migratePath = wd + "/" + migratePath
	}

	m, err := migrate.NewWithSourceInstance("iofs", d, "yaml://"+migratePath)
	if err != nil {
		return err
	}

	if err := m.Up(); errors.Is(err, migrate.ErrNoChange) {
		return nil
	} else if err != nil {
		return err
	}

	return nil
}
