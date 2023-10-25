package iofs_test

import (
	"embed"
	"log"

	"github.com/c2pc/go-pkg/migration"
	_ "github.com/c2pc/go-pkg/migration/config/yaml"
	"github.com/c2pc/go-pkg/migration/source/iofs"
)

//go:embed testdata/migrations/*.sql
var fs embed.FS

func Example() {
	d, err := iofs.New(fs, "testdata/migrations")
	if err != nil {
		log.Fatal(err)
	}
	m, err := migrate.NewWithSourceInstance("iofs", d, "yaml://./testdata/config.yml")
	if err != nil {
		log.Fatal(err)
	}
	err = m.Up()
	if err != nil {
		// ...
	}
	// ...
}
