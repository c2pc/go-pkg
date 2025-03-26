package db

import (
	"embed"
	"errors"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

func Migrate(fs embed.FS, migrationsDir, databaseName string, driver database.Driver) error {
	d, err := iofs.New(fs, migrationsDir)
	if err != nil {
		return err
	}

	m, err := migrate.NewWithInstance("iofs", d, databaseName, driver)
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
