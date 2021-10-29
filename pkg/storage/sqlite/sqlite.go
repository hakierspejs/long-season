package sqlite

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
	"sync"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite"
	"github.com/golang-migrate/migrate/v4/source/iofs"

	_ "modernc.org/sqlite"

	"github.com/hakierspejs/long-season/pkg/storage"
)

//go:embed migrations
var migrations embed.FS

const migrationsCurrentVersion = 1

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

type Factory struct {
	UsersStorage *Users
}

func NewFactory(filename string) (*Factory, error) {
	db, err := sql.Open("sqlite", filename)
	if err != nil {
		return nil, fmt.Errorf("sql.Open: %w", err)
	}

	if err := migrateWithFS(db, migrations); err != nil {
		return nil, fmt.Errorf("migrateWithFS: %w", err)
	}

	cs := &coreStorage{
		db:         db,
		writeGuard: new(sync.Mutex),
	}

	return &Factory{
		UsersStorage: &Users{
			cs: cs,
		},
	}, nil
}

func (f *Factory) Users() storage.Users {
	return f.UsersStorage
}

func pragma(query string) string {
	res := ""
	res += "PRAGMA foreign_keys = ON;"
	res += query
	return res

}

func sqliteBoolean(v bool) int {
	if !v {
		return 0
	}
	return 1
}

type coreStorage struct {
	db *sql.DB

	// writeGuard synchronizes write operations to
	// sqlite database.
	writeGuard *sync.Mutex
}

func (cs *coreStorage) newUser(ctx context.Context, u storage.UserEntry) (string, error) {
	query := pragma(`
	INSERT INTO users
		(userID, userNickname, userPassword, userPrivate)
	VALUES
		($1, $2, $3, $4);
	`)

	cs.writeGuard.Lock()
	defer cs.writeGuard.Unlock()

	_, err := cs.db.ExecContext(
		ctx,
		query,
		u.ID,
		u.Nickname,
		u.HashedPassword,
		sqliteBoolean(u.Private),
	)
	if err != nil {
		return "", fmt.Errorf("cs.db.ExecContext: %w", err)
	}

	return u.ID, nil
}

func (cs *coreStorage) readUser(ctx context.Context, id string) (*storage.UserEntry, error) {
	query := `
	SELECT
		userNickname, userPassword, userPrivate
	FROM
		users
	WHERE
		userID = $1
	`
	var (
		userNickname string
		userPassword []byte
		userPrivate  int
	)
	err := cs.db.QueryRowContext(ctx, query, id).Scan(
		&userNickname,
		&userPassword,
		&userPrivate,
	)
	if err != nil {
		return nil, fmt.Errorf("cs.db.QueryRowContext: %w", err)
	}

	return &storage.UserEntry{
		ID:             id,
		Nickname:       userNickname,
		HashedPassword: userPassword,
		Private:        userPrivate >= 1,
	}, nil
}

func copyBytes(src []byte) []byte {
	dst := make([]byte, len(src), cap(src))
	copy(dst, src)
	return dst
}

func (cs *coreStorage) allUsers(ctx context.Context) ([]storage.UserEntry, error) {
	query := `
	SELECT
		userID, userNickname, userPassword, userPrivate
	FROM
		users
	`

	var (
		userID       string
		userNickname string
		userPassword []byte
		userPrivate  int
	)

	rows, err := cs.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("cs.db.QueryContext: %w", err)
	}
	defer rows.Close()

	res := []storage.UserEntry{}

	for rows.Next() {
		err = rows.Scan(
			&userID,
			&userNickname,
			&userPassword,
			&userPrivate,
		)
		if err != nil {
			return nil, fmt.Errorf("rows.Scan: %w", err)
		}

		res = append(res, storage.UserEntry{
			ID:             userID,
			Nickname:       userNickname,
			HashedPassword: copyBytes(userPassword),
			Private:        userPrivate >= 1,
		})
	}

	return res, nil
}

func (cs *coreStorage) removeUser(ctx context.Context, id string) error {
	query := `
	DELETE FROM
		users
	WHERE
		userID = $1;
	`

	cs.writeGuard.Lock()
	defer cs.writeGuard.Unlock()

	_, err := cs.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("cs.db.ExecContext: %w", err)
	}

	return nil
}

func (cs *coreStorage) updateUser(ctx context.Context, id string, f func(*storage.UserEntry) error) error {
	cs.writeGuard.Lock()
	defer cs.writeGuard.Unlock()

	tx, err := cs.db.Begin()
	if err != nil {
		return fmt.Errorf("cs.db.Begin: %w", err)
	}

	selectUserQuery := `
	SELECT
		userNickname, userPassword, userPrivate
	FROM
		users
	WHERE
		userID = $1
	`

	var (
		userNickname string
		userPassword []byte
		userPrivate  int
	)

	err = tx.QueryRowContext(ctx, selectUserQuery, id).Scan(
		&userNickname,
		&userPassword,
		&userPrivate,
	)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("tx.QueryRowContext: %w", err)
	}

	entry := &storage.UserEntry{
		ID:             id,
		Nickname:       userNickname,
		HashedPassword: userPassword,
		Private:        userPrivate >= 1,
	}

	err = f(entry)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("f: %w", err)
	}

	updateQuery := pragma(`
	UPDATE
		users
	SET
		userNickname = $2, userPassword = $3, userPrivate = $4
	WHERE
		userID = $1;
	`)

	_, err = tx.ExecContext(
		ctx,
		updateQuery,
		entry.ID,
		entry.Nickname,
		entry.HashedPassword,
		sqliteBoolean(entry.Private),
	)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("tx.ExecContext: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("tx.Commit: %w", err)
	}

	return nil
}
