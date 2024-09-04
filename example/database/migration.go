package database

import (
	"database/sql"
	"embed"
	"errors"
	"github.com/c2pc/go-pkg/v2/utils/migrate"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

const PATH = "migrations"

//go:embed migrations/*.sql
var fs embed.FS

func Migrate(db *sql.DB, databaseName string) error {
	d, err := iofs.New(fs, PATH)
	if err != nil {
		return err
	}

	driver, err := postgres.WithInstance(db, &postgres.Config{})
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
