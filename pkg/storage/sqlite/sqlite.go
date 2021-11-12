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

	"github.com/hakierspejs/long-season/pkg/models"
	"github.com/hakierspejs/long-season/pkg/models/set"
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

// Factory implements storage factory interface.
type Factory struct {
	UsersStorage     *Users
	DevicesStorage   *Devices
	TwoFactorStorage *TwoFactor
}

// NewFactory returns Factory, database closer for sqlite connection and
// error if something went wrong.
func NewFactory(filename string) (*Factory, func() error, error) {
	db, err := sql.Open("sqlite", filename)
	if err != nil {
		return nil, nil, fmt.Errorf("sql.Open: %w", err)
	}

	if err := migrateWithFS(db, migrations); err != nil {
		return nil, nil, fmt.Errorf("migrateWithFS: %w", err)
	}

	cs := &coreStorage{
		db:         db,
		writeGuard: new(sync.Mutex),
	}

	closer := func() error {
		return db.Close()
	}

	return &Factory{
		UsersStorage: &Users{
			cs: cs,
		},
		DevicesStorage: &Devices{
			cs: cs,
		},
		TwoFactorStorage: &TwoFactor{
			cs: cs,
		},
	}, closer, nil
}

// Users returns sqlite implementation of
// storage Users interface.
func (f *Factory) Users() storage.Users {
	return f.UsersStorage
}

// Devices returns sqlite implementation of
// storage Devices interface.
func (f *Factory) Devices() storage.Devices {
	return f.DevicesStorage
}

// TwoFactor returns sqlite implementation of
// storage TwoFactor interface.
func (f *Factory) TwoFactor() storage.TwoFactor {
	return f.TwoFactorStorage
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

func (cs *coreStorage) newDevice(ctx context.Context, userID string, d models.Device) (string, error) {
	query := pragma(`
	INSERT INTO devices
		(deviceID, deviceOwnerID, deviceTag, deviceMAC)
	VALUES
		($1, $2, $3, $4)
	`)

	cs.writeGuard.Lock()
	defer cs.writeGuard.Unlock()

	_, err := cs.db.ExecContext(
		ctx,
		query,
		d.ID,
		userID,
		d.Tag,
		d.MAC,
	)
	if err != nil {
		return "", fmt.Errorf("cs.db.ExecContext: %w", err)
	}

	return d.ID, nil
}

func (cs *coreStorage) deviceOfUser(ctx context.Context, userID string) ([]models.Device, error) {
	query := `
	SELECT
		deviceID, deviceOwnerID, userNickname, deviceTag, deviceMAC
	FROM
		users INNER JOIN devices
	ON
		users.userID = devices.deviceOwnerID
	WHERE
		users.userID = $1;
	`

	var (
		deviceID      string
		deviceOwnerID string
		userNickname  string
		deviceTag     string
		deviceMAC     []byte
	)

	rows, err := cs.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("cs.db.QueryContext: %w", err)
	}
	defer rows.Close()

	res := []models.Device{}

	for rows.Next() {
		err = rows.Scan(
			&deviceID,
			&deviceOwnerID,
			&userNickname,
			&deviceTag,
			&deviceMAC,
		)
		if err != nil {
			return nil, fmt.Errorf("rows.Scan: %w", err)
		}

		res = append(res, models.Device{
			DevicePublicData: models.DevicePublicData{
				ID:    deviceID,
				Tag:   deviceTag,
				Owner: userNickname,
			},
			OwnerID: deviceOwnerID,
			MAC:     copyBytes(deviceMAC),
		})
	}

	return res, nil
}

func (cs *coreStorage) allDevices(ctx context.Context) ([]models.Device, error) {
	query := `
	SELECT
		deviceID, deviceOwnerID, userNickname, deviceTag, deviceMAC
	FROM
		users INNER JOIN devices
	ON
		users.userID = devices.deviceOwnerID;
	`

	var (
		deviceID      string
		deviceOwnerID string
		userNickname  string
		deviceTag     string
		deviceMAC     []byte
	)

	rows, err := cs.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("cs.db.QueryContext: %w", err)
	}
	defer rows.Close()

	res := []models.Device{}

	for rows.Next() {
		err = rows.Scan(
			&deviceID,
			&deviceOwnerID,
			&userNickname,
			&deviceTag,
			&deviceMAC,
		)
		if err != nil {
			return nil, fmt.Errorf("rows.Scan: %w", err)
		}

		res = append(res, models.Device{
			DevicePublicData: models.DevicePublicData{
				ID:    deviceID,
				Tag:   deviceTag,
				Owner: userNickname,
			},
			OwnerID: deviceOwnerID,
			MAC:     copyBytes(deviceMAC),
		})
	}

	return res, nil
}

func (cs *coreStorage) readDevice(ctx context.Context, id string) (*models.Device, error) {
	query := `
	SELECT
		deviceID, deviceOwnerID, userNickname, deviceTag, deviceMAC
	FROM
		users INNER JOIN devices
	ON
		users.userID = devices.deviceOwnerID
	WHERE
		devices.deviceID = $1;
	`

	var (
		deviceID      string
		deviceOwnerID string
		userNickname  string
		deviceTag     string
		deviceMAC     []byte
	)

	err := cs.db.QueryRowContext(ctx, query, id).Scan(
		&deviceID,
		&deviceOwnerID,
		&userNickname,
		&deviceTag,
		&deviceMAC,
	)
	if err != nil {
		return nil, fmt.Errorf("cs.db.QueryRowContext: %w", err)
	}

	return &models.Device{
		DevicePublicData: models.DevicePublicData{
			ID:    deviceID,
			Tag:   deviceTag,
			Owner: userNickname,
		},
		OwnerID: deviceOwnerID,
		MAC:     deviceMAC,
	}, nil
}

func (cs *coreStorage) removeDevice(ctx context.Context, id string) error {
	query := `
	DELETE FROM
		devices
	WHERE
		deviceID = $1;
	`
	cs.writeGuard.Lock()
	defer cs.writeGuard.Unlock()

	_, err := cs.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("cs.db.ExecContext: %w", err)
	}

	return nil
}

func getTwoFactorFromTx(ctx context.Context, tx *sql.Tx, userID string) (*models.TwoFactor, error) {
	res := models.TwoFactor{
		OneTimeCodes:  map[string]models.OneTimeCode{},
		RecoveryCodes: map[string]models.Recovery{},
	}

	query := `
	SELECT
		otpID, otpName, otpSecret
	FROM
		otp
	WHERE
		otpOwnerID = $1;
	`
	rows, err := tx.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("otp/tx.QueryContext: %w", err)
	}

	var (
		otpID     string
		otpName   string
		otpSecret string
	)

	for rows.Next() {
		err = rows.Scan(&otpID, &otpName, &otpSecret)
		if err != nil {
			return nil, fmt.Errorf("otp/rows.Scan: %w", err)
		}
		res.OneTimeCodes[otpID] = models.OneTimeCode{
			ID:     otpID,
			Name:   otpName,
			Secret: otpSecret,
		}
	}
	if rows.Err(); err != nil {
		return nil, fmt.Errorf("otp/rows.Err: %w", err)
	}

	query = `
	SELECT
		recoveryID, recoveryName
	FROM
		recovery
	WHERE
		recoveryOwnerID = $1;
	`
	rows, err = tx.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("recovery/tx.QueryContext: %w", err)
	}

	var (
		recoveryID   string
		recoveryName string
	)

	for rows.Next() {
		err = rows.Scan(&recoveryID, &recoveryName)
		if err != nil {
			return nil, fmt.Errorf("recovery/rows.Scan: %w", err)
		}
		res.RecoveryCodes[recoveryID] = models.Recovery{
			ID:    recoveryID,
			Name:  recoveryName,
			Codes: set.NewString(),
		}
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("recovery/rows.Err: %w", err)
	}

	for _, recovery := range res.RecoveryCodes {
		query := `
		SELECT
			recoveryCodesCode
		FROM
			recoveryCodes
		WHERE
			recoveryCodesID = $1;
		`
		var recoveryCode string

		rows, err = tx.QueryContext(ctx, query, recovery.ID)
		if err != nil {
			return nil, fmt.Errorf("recoveryCodes/tx.QueryContext: %w", err)
		}

		for rows.Next() {
			err = rows.Scan(&recoveryCode)
			if err != nil {
				return nil, fmt.Errorf("recoveryCodes/rows.Scan: %w", err)
			}
			res.RecoveryCodes[recovery.ID].Codes.Push(recoveryCode)
		}
		if err := rows.Err(); err != nil {
			return nil, fmt.Errorf("recoveryCodes/rows.Err: %w", err)
		}
	}

	return &res, nil
}

func (cs *coreStorage) getTwoFactor(ctx context.Context, userID string) (*models.TwoFactor, error) {
	tx, err := cs.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("cs.db.Begin: %w", err)
	}

	res, err := getTwoFactorFromTx(ctx, tx, userID)
	if err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("getTwoFactorFromTx: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("tx.Commit: %w", err)
	}

	return res, nil
}

func removeTwoFactorFromTx(ctx context.Context, tx *sql.Tx, userID string) error {
	query := pragma(`
	DELETE FROM
		otp
	WHERE
		otpOwnerID = $1;
	`)
	_, err := tx.ExecContext(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("tx.ExecContext: %w", err)
	}

	query = pragma(`
	DELETE FROM
		recovery
	WHERE
		recoveryOwnerID = $1;
	`)
	_, err = tx.ExecContext(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("tx.ExecContext: %w", err)
	}

	return nil
}

func (cs *coreStorage) removeTwoFactor(ctx context.Context, userID string) error {
	cs.writeGuard.Lock()
	defer cs.writeGuard.Unlock()

	tx, err := cs.db.Begin()
	if err != nil {
		return fmt.Errorf("cs.db.Begin: %w", err)
	}

	err = removeTwoFactorFromTx(ctx, tx, userID)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("removeTwoFactorFromTx: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("tx.Commit: %w", err)

	}
	return nil
}

func insertTwoFactorWithTx(ctx context.Context, tx *sql.Tx, userID string, tf models.TwoFactor) error {
	otpInsertStmt, err := tx.PrepareContext(ctx, pragma(`
	INSERT INTO otp
		(otpID, otpName, otpSecret, otpOwnerID)
	VALUES
		($1, $2, $3, $4);
	`))
	if err != nil {
		return fmt.Errorf("tx.PrepareContext: %w", err)
	}

	for _, otp := range tf.OneTimeCodes {
		_, err = otpInsertStmt.ExecContext(ctx,
			otp.ID,
			otp.Name,
			otp.Secret,
			userID,
		)
		if err != nil {
			return fmt.Errorf("otpInsertStmt.ExecContext: %w", err)
		}
	}

	recoveryInsertStmt, err := tx.PrepareContext(ctx, pragma(`
	INSERT INTO recovery
		(recoveryID, recoveryName, recoveryOwnerID)
	VALUES
		($1, $2, $3);
	`))
	if err != nil {
		return fmt.Errorf("tx.PrepareContext: %w", err)
	}

	recoveryCodesInsertStmt, err := tx.PrepareContext(ctx, pragma(`
	INSERT INTO recoveryCodes
		(recoveryCodesCode, recoveryCodesID)
	VALUES
		($1, $2);
	`))
	if err != nil {
		return fmt.Errorf("tx.PrepareContext: %w", err)
	}

	for _, recovery := range tf.RecoveryCodes {
		_, err = recoveryInsertStmt.ExecContext(ctx,
			recovery.ID,
			recovery.Name,
			userID,
		)
		if err != nil {
			return fmt.Errorf("recoveryInsertStmt.ExecContext: %w", err)
		}

		for _, code := range recovery.Codes.Items() {
			_, err = recoveryCodesInsertStmt.ExecContext(ctx, code, recovery.ID)
			if err != nil {
				return fmt.Errorf("recoveryCodesInsertStmt.ExecContext: %w", err)
			}
		}
	}

	return nil
}

func (cs *coreStorage) updateTwoFactor(ctx context.Context, userID string, f func(*models.TwoFactor) error) error {
	cs.writeGuard.Lock()
	defer cs.writeGuard.Unlock()

	tx, err := cs.db.Begin()
	if err != nil {
		return fmt.Errorf("cs.db.Begin: %w", err)
	}

	tf, err := getTwoFactorFromTx(ctx, tx, userID)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("getTwoFactorFromTx: %w", err)
	}

	err = removeTwoFactorFromTx(ctx, tx, userID)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("removeTwoFactorFromTx: %w", err)
	}

	err = f(tf)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("f(tf) update function error: %w", err)
	}

	err = insertTwoFactorWithTx(ctx, tx, userID, *tf)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("insertTwoFactorWithTx: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("tx.Commit: %w", err)

	}

	return nil
}
