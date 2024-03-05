package postgres

import (
	"errors"
	"fmt"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jmoiron/sqlx"
)

const baseMigrationDir = "file://migrations/"

func RunMigrations(db *sqlx.DB, opts *options) {
	driver, err := postgres.WithInstance(db.DB, &postgres.Config{
		MigrationsTable: opts.migrationsTableName,
	})
	if err != nil {
		panic(err)
	}

	dir := baseMigrationDir
	if opts.migrationDir != "" {
		dir = opts.migrationDir
	}

	if opts.migrationUseDBName {
		dir = strings.TrimRight(fmt.Sprintf("%v/%v", dir, opts.name.String()), "/")
	}
	m, err := migrate.NewWithDatabaseInstance(
		dir,
		"postgres", driver)
	if err != nil {
		panic(err)
	}

	err = m.Up()
	if err == nil || errors.Is(err, migrate.ErrNoChange) {
		return
	}

	panic(err)
}
