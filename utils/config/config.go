package config

import (
	"gopkg.in/yaml.v3"
	"os"
)

func NewConfig[T any](configPath string) (*T, error) {
	var config *T

	file, err := os.Open(configPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	d := yaml.NewDecoder(file)

	if err := d.Decode(&config); err != nil {
		return nil, err
	}

	return config, nil
}
