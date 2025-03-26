package config

import (
	"embed"
	"errors"
	"os"
	"strings"

	_ "github.com/c2pc/config-migrate/config/yaml"
	_ "github.com/c2pc/config-migrate/replacer/ip"
	_ "github.com/c2pc/config-migrate/replacer/project_name"
	_ "github.com/c2pc/config-migrate/replacer/random"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

// Migrate выполняет миграции на основе предоставленных файлов миграций и пути к миграции.
// fs - файловая система, содержащая миграции.
// migrationsDir - директория в файловой системе, где хранятся миграции.
// migratePath - путь к файлу миграции.
func Migrate(fs embed.FS, migrationsDir, migratePath string) error {
	// Создание нового источника миграций на основе файловой системы
	d, err := iofs.New(fs, migrationsDir)
	if err != nil {
		return err
	}

	// Если путь к миграции не содержит разделителей путей, добавляем рабочую директорию
	if !strings.Contains(migratePath, "/") && !strings.Contains(migratePath, "\\") {
		wd, err := os.Getwd()
		if err != nil {
			return err
		}
		migratePath = wd + "/" + migratePath
	}

	// Создание нового миграционного экземпляра с источником миграций
	m, err := migrate.NewWithSourceInstance("iofs", d, "yaml://"+migratePath)
	if err != nil {
		return err
	}

	// Выполнение миграций
	if err := m.Up(); errors.Is(err, migrate.ErrNoChange) {
		return nil // Нет изменений в миграциях, возвращаем nil
	} else if err != nil {
		return err // Возвращаем ошибку, если что-то пошло не так
	}

	return nil
}
