package migrations

import (
	"errors"

	"github.com/0x0FACED/link-saver-api/internal/wrap"
	"github.com/golang-migrate/migrate"
	_ "github.com/golang-migrate/migrate/database/postgres"
	_ "github.com/golang-migrate/migrate/source/file"
)

var pkg = "migrations"

func Up(url string) error {
	m, err := migrate.New(
		"file://./migrations/",
		url)
	if err != nil {
		return err
	}

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return wrap.E(pkg, "failed to Up()", err)
	}

	return nil
}
