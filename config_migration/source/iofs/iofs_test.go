package iofs_test

import (
	"testing"

	"github.com/c2pc/go-pkg/config_migration/source/iofs"
	st "github.com/c2pc/go-pkg/config_migration/source/testing"
)

func Test(t *testing.T) {
	// reuse the embed.FS set in example_test.go
	d, err := iofs.New(fs, "testdata/migrations")
	if err != nil {
		t.Fatal(err)
	}

	st.Test(t, d)
}
