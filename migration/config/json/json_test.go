package json

import (
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"os"
	"testing"
)

func TestJson_Up(t *testing.T) {
	m, err := migrate.New("file://./examples/migrations", "json://./examples/config.json")
	if err != nil {
		t.Fatal(err)
	}

	if err := m.Up(); err != nil {
		t.Fatal(err)
	}

	if err := os.Remove("./examples/config.json"); err != nil {
		t.Fatal(err)
	}
}
