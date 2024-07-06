package migrate

import (
	"errors"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/golang-migrate/migrate/v4/source/pkger"
	"strings"
)

var (
	ErrNoChange  = migrate.ErrNoChange
	ErrEmptyPath = errors.New("empty path")
)

func New(source, filePath, database string) (*migrate.Migrate, error) {
	if filePath == "" {
		return nil, ErrEmptyPath
	}
	filePath = strings.ReplaceAll(filePath, "\\", "/")

	m, err := migrate.New(source+"://"+filePath, database)
	if err != nil {
		return nil, err
	}

	return m, nil
}

func NewWithSourceInstance(sourceName string, sourceInstance source.Driver, databaseURL string) (*migrate.Migrate, error) {
	m, err := migrate.NewWithSourceInstance(sourceName, sourceInstance, databaseURL)
	if err != nil {
		return nil, err
	}

	return m, nil
}

func NewWithInstance(sourceName string, sourceInstance source.Driver, databaseName string, databaseInstance database.Driver) (*migrate.Migrate, error) {
	m, err := migrate.NewWithInstance(sourceName, sourceInstance, databaseName, databaseInstance)
	if err != nil {
		return nil, err
	}

	return m, nil
}
