package db

import (
	"embed"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database"
	"github.com/golang-migrate/migrate/v4/source"
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

	currentVersion, _, _ := driver.Version()
	lastVersion := getLastVersion(d)

	if currentVersion > lastVersion {
		return errors.New(fmt.Sprintf(
			"Версия миграции в БД %s имеет версию v%d, которая выше, чем поддерживаемая этой версией сервиса (v%d). Вероятно, сервис был откатан на более раннюю версию. Чтобы избежать ошибок, нужно откатить изменения.",
			databaseName, currentVersion, lastVersion,
		))
	}

	if err := m.Up(); errors.Is(err, migrate.ErrNoChange) {
		return nil // Нет изменений в миграциях, возвращаем nil
	} else if err != nil {
		return errors.New(fmt.Sprintf("Произошла ошибка миграции в БД %s: %s", databaseName, strings.ReplaceAll(err.Error(), "\n", " ")))
	}

	return nil
}

func getLastVersion(d source.Driver) int {
	firstVersion, err := d.First()
	if err != nil {
		return 0
	}

	var lastVersion = firstVersion
	for {
		next, err := d.Next(lastVersion)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				return int(lastVersion)
			}
			return 0
		}
		lastVersion = next
	}
}
