package postgres

import (
	"errors"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jmoiron/sqlx"
)

const (
	baseMigrationDir = "file://migrations/"
)

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

	m, err := migrate.NewWithDatabaseInstance(
		fmt.Sprintf("%v/%v", dir, opts.name.String()),
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
