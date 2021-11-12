// Package abstract implements abstract factory
// for storage Factory interface.
package abstract

import (
	"errors"
	"fmt"

	bolt "go.etcd.io/bbolt"

	"github.com/hakierspejs/long-season/pkg/storage"
	"github.com/hakierspejs/long-season/pkg/storage/memory"
	"github.com/hakierspejs/long-season/pkg/storage/sqlite"
)

// ErrInvalidDBType is returned when database type cannot be recognized.
var ErrInvalidDBType = errors.New("abstract: invalid database type")

// Factory is abstract factory for storage.Factory interface. Returns
// storage factory for given type and uses storage file placed
// at given path.
func Factory(dbPath string, dbType string) (storage.Factory, func(), error) {
	switch dbType {
	case "bolt":
		boltDB, err := bolt.Open(dbPath, 0666, nil)
		if err != nil {
			return nil, nil, fmt.Errorf("bolt.Open: %w", err)
		}
		boltFactory, err := memory.New(boltDB)
		if err != nil {
			return nil, nil, fmt.Errorf("memory.New: %w", err)
		}

		boltCloser := func() {
			boltDB.Close()
		}

		return boltFactory, boltCloser, nil
	case "sqlite":
		sqliteDB, closer, err := sqlite.NewFactory(dbPath)
		if err != nil {
			return nil, nil, fmt.Errorf("sqlite.NewFactory: %w", err)
		}

		sqliteCloser := func() {
			closer()
		}

		return sqliteDB, sqliteCloser, nil
	default:
		return nil, nil, ErrInvalidDBType
	}
}
