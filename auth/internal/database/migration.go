package database

import (
	"database/sql"
	"embed"

	database "github.com/c2pc/go-pkg/v2/utils/db"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"gorm.io/gorm"
)

//go:embed migrations/postgres/*.sql
var fsPostgre embed.FS

func Migrate(db *gorm.DB) error {
	if db.Dialector.Name() != "postgres" {
		return nil
	}

	sqlDB, err := db.DB()
	if err != nil {
		return err
	}

	return MigratePostgres(sqlDB)
}

func MigratePostgres(db *sql.DB) error {
	driver, err := postgres.WithInstance(db, &postgres.Config{
		MigrationsTable: "schema_auth_migrations",
	})
	if err != nil {
		return err
	}

	err = database.Migrate(fsPostgre, "migrations/postgres", "postgres", driver)
	if err != nil {
		return err
	}
	return nil
}
