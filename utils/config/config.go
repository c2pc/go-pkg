package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

// NewConfig создает и возвращает конфигурацию типа T, считывая данные из файла YAML.
// configPath - путь к файлу конфигурации YAML.
func NewConfig[T any](configPath string) (*T, error) {
	var config *T

	// Открытие файла конфигурации
	file, err := os.Open(configPath)
	if err != nil {
		return nil, err
	}
	defer file.Close() // Закрытие файла после завершения работы функции

	// Создание нового декодера YAML
	d := yaml.NewDecoder(file)

	// Декодирование содержимого файла YAML в переменную config
	if err := d.Decode(&config); err != nil {
		return nil, err
	}

	return config, nil
}
