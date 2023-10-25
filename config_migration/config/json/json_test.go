package json

import (
	"errors"
	migrate "github.com/c2pc/go-pkg/config_migration"
	_ "github.com/c2pc/go-pkg/config_migration/source/file"
	"testing"
)

func TestJson_Version(t *testing.T) {
	m, err := migrate.New("file://./examples/migrations", "json://./examples/config.json")
	if err != nil {
		t.Fatal(err)
	}

	version, err := m.Version()
	if err != nil {
		if !errors.Is(err, migrate.ErrNilVersion) {
			t.Fatal(err)
		}
		return
	}

	if version <= 0 {
		t.Error("Invalid version", version)
	}
}

func TestJson_Run(t *testing.T) {
	m, err := migrate.New("file://./examples/migrations", "json://./examples/config.json")
	if err != nil {
		t.Fatal(err)
	}

	if err := m.Up(); err != nil {
		t.Fatal(err)
	}
}
