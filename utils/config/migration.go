package config

import (
	"embed"
	"errors"
	"fmt"
	"os"
	"strings"

	migrator "github.com/c2pc/config-migrate/driver"
	"github.com/c2pc/config-migrate/driver/yaml"
	_ "github.com/c2pc/config-migrate/driver/yaml"
	_ "github.com/c2pc/config-migrate/replacer/ip"
	_ "github.com/c2pc/config-migrate/replacer/project_name"
	_ "github.com/c2pc/config-migrate/replacer/random"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/source"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

// Migrate выполняет миграции на основе предоставленных файлов миграций и пути к миграции.
// fs - файловая система, содержащая миграции.
// migrationsDir - директория в файловой системе, где хранятся миграции.
// migratePath - путь к файлу миграции.
func Migrate(fs embed.FS, migrationsDir, migratePath string) error {
	fileName := migratePath
	// Если путь к миграции не содержит разделителей путей, добавляем рабочую директорию
	if !strings.Contains(migratePath, "/") && !strings.Contains(migratePath, "\\") {
		wd, err := os.Getwd()
		if err != nil {
			return err
		}
		migratePath = wd + "/" + migratePath
	}

	yamlMigr := yaml.New(migrator.Settings{
		Path:                    migratePath,
		Perm:                    0666,
		UnableToReplaceComments: true,
	})

	// Создание нового источника миграций на основе файловой системы
	d, err := iofs.New(fs, migrationsDir)
	if err != nil {
		return err
	}

	// Создание нового миграционного экземпляра с источником миграций
	m, err := migrate.NewWithInstance("iofs", d, "yaml", yamlMigr)
	if err != nil {
		return err
	}

	yamlMigr.Lock()
	currentVersion, _, _ := yamlMigr.Version()
	yamlMigr.Unlock()
	lastVersion := getLastVersion(d)

	if currentVersion > lastVersion {
		return errors.New(fmt.Sprintf(
			"Файл конфигурации %s имеет версию v%d, которая выше, чем поддерживаемая этой версией сервиса (v%d). Вероятно, сервис был откатан на более раннюю версию. Чтобы избежать ошибок, удалите или переместите этот файл, перезапустите сервис — он создаст новую конфигурацию автоматически, затем настройте конфигурацию заново",
			fileName, currentVersion, lastVersion,
		))
	}

	// Выполнение миграций
	if err := m.Up(); errors.Is(err, migrate.ErrNoChange) {
		return nil // Нет изменений в миграциях, возвращаем nil
	} else if err != nil {
		return errors.New(fmt.Sprintf("Произошла ошибка миграции файла конфигураций %s: %s", fileName, strings.ReplaceAll(err.Error(), "\n", " ")))
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
