package config

import (
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"os"
	"testing"
)

func TestYaml_Up(t *testing.T) {
	m, err := migrate.New("file://./examples/json", "yaml://./examples/config.yml")
	if err != nil {
		t.Fatal(err)
	}

	if err := m.Up(); err != nil {
		t.Fatal(err)
	}

	if err := os.Remove("./examples/config.yml"); err != nil {
		t.Fatal(err)
	}
}
