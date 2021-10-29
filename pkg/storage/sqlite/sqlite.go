package sqlite

import (
	"database/sql"
	"embed"
	"fmt"
	"io/fs"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite"
	"github.com/golang-migrate/migrate/v4/source/iofs"

	_ "modernc.org/sqlite"
)

//go:embed migrations
var migrations embed.FS

const migrationsCurrentVersion = 1

type Factory struct {
}

func NewFactory(filename string) (*Factory, error) {
	db, err := sql.Open("sqlite", filename)
	if err != nil {
		return nil, fmt.Errorf("sql.Open: %w", err)
	}

	if err := migrateWithFS(db, migrations); err != nil {
		return nil, fmt.Errorf("migrateWithFS: %w", err)
	}

	return &Factory{}, nil
}

func migrateWithFS(db *sql.DB, fileSystem fs.FS) error {
	sourceInstance, err := iofs.New(fileSystem, "migrations")
	if err != nil {
		return fmt.Errorf("iofs.New: %w", err)
	}

	targetInstance, err := sqlite.WithInstance(db, new(sqlite.Config))
	if err != nil {
		return fmt.Errorf("sqlite.WithInstance: %w", err)
	}

	m, err := migrate.NewWithInstance("iofs", sourceInstance, "sqlite", targetInstance)
	if err != nil {
		return fmt.Errorf("migrate.NewWithInstance: %w", err)
	}

	err = m.Migrate(migrationsCurrentVersion)
	if err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("m.Migrate: %w", err)
	}

	return sourceInstance.Close()
}
