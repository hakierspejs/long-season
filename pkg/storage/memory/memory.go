// Package memory implements storage interfaces for
// bolt key-value store database.
package memory

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	bolt "go.etcd.io/bbolt"

	"github.com/hakierspejs/long-season/pkg/models"
	"github.com/hakierspejs/long-season/pkg/storage"
	serrors "github.com/hakierspejs/long-season/pkg/storage/errors"
)

const (
	countersBucket       = "ls::counters"
	usersBucket          = "ls::users"
	devicesBucket        = "ls::devices"
	twoFactorBucket      = "ls::twofactor"
	devicesBucketCounter = "ls::devices::counter"
)

// Factory implements storage.Factory interface for
// bolt database.
type Factory struct {
	users           *UsersStorage
	devices         *DevicesStorage
	statusStorageTx *StatusStorageTx
	twoFactor       *TwoFactorStorage
}

// Users returns storage interface for manipulating
// users data.
func (f Factory) Users() *UsersStorage {
	return f.users
}

func (f Factory) Devices() *DevicesStorage {
	return f.devices
}

func (f Factory) TwoFactor() *TwoFactorStorage {
	return f.twoFactor
}

// StatusTx returns storage interface for
// reading and writing information about numbers
// of online users and unkown devices.
func (f Factory) StatusTx() *StatusStorageTx {
	return f.statusStorageTx
}

// New returns pointer to new memory storage
// Factory.
func New(db *bolt.DB) (*Factory, error) {
	buckets := []string{
		usersBucket,
		devicesBucket,
		twoFactorBucket,
	}
	err := db.Update(func(tx *bolt.Tx) error {
		for _, b := range buckets {
			_, err := tx.CreateBucketIfNotExists([]byte(b))
			if err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return &Factory{
		users:           &UsersStorage{db},
		devices:         &DevicesStorage{db},
		statusStorageTx: &StatusStorageTx{db},
		twoFactor:       &TwoFactorStorage{db},
	}, nil
}

// UsersStorage implements storage.Users interface
// for bolt database.
type UsersStorage struct {
	db *bolt.DB
}

const (
	userIDKey          = "ls::user::id"
	userNicknameKey    = "ls::user::nickname"
	userPasswordKey    = "ls::user::password"
	userPrivateModeKey = "ls::user::private_mode"
)

func boolToBytes(b bool) []byte {
	if b {
		return []byte{1}
	} else {
		return []byte{0}
	}
}

func bytesToBool(b []byte) bool {
	return b[0] > 0
}

func userFromBucket(b *bolt.Bucket) (*storage.UserEntry, error) {
	result := new(storage.UserEntry)

	fail := func() (*storage.UserEntry, error) {
		return nil, serrors.ErrNoID
	}

	idBytes := b.Get([]byte(userIDKey))
	if idBytes == nil {
		return fail()
	}
	result.ID = string(idBytes)

	nickname := b.Get([]byte(userNicknameKey))
	if nickname == nil {
		return fail()
	}
	result.Nickname = string(nickname)

	password := b.Get([]byte(userPasswordKey))
	if password == nil {
		return fail()
	}
	result.HashedPassword = password

	priv := b.Get([]byte(userPrivateModeKey))
	if priv == nil {
		// This is a new field implemented after #43, so
		// we can handle situation when field is nil and
		// assume default value.
		result.Private = false
	} else {
		result.Private = bytesToBool(priv)
	}

	return result, nil
}

func userBucketKey(id string) []byte {
	return []byte(fmt.Sprintf("ls::user::%s", id))
}

type bucketMapping struct {
	key   []byte
	value []byte
}

// storeUserInBucket stores given user model in given bucket, by
// creating sub-bucket with appropriate user bucket key according to
// given user's id.
func storeUserInBucket(user storage.UserEntry, b *bolt.Bucket) error {
	// TODO(thinkofher) Wrap errors with fmt.Errorf and "%w".
	userBucket, err := b.CreateBucketIfNotExists(userBucketKey(user.ID))
	if err != nil {
		return err
	}

	id := []byte(user.ID)

	// keys and values for user data model
	kvs := []bucketMapping{
		{[]byte(userIDKey), id},
		{[]byte(userNicknameKey), []byte(user.Nickname)},
		{[]byte(userPasswordKey), user.HashedPassword},
		{[]byte(userPrivateModeKey), boolToBytes(user.Private)},
	}

	for _, item := range kvs {
		if err := userBucket.Put(item.key, item.value); err != nil {
			return err
		}
	}

	return nil
}

func readUser(tx *bolt.Tx, userID string) (*storage.UserEntry, error) {
	b := tx.Bucket([]byte(usersBucket))
	if b == nil {
		return nil, fmt.Errorf("bucket empty")
	}

	userBucket := b.Bucket([]byte(userBucketKey(userID)))
	if userBucket == nil {
		return nil, fmt.Errorf("there is no user with id=%s, err=%w", userID, serrors.ErrNoID)
	}

	return userFromBucket(userBucket)
}

// New stores given user data in database and returns
// assigned (given) id.
func (s *UsersStorage) New(ctx context.Context, newUser storage.UserEntry) (string, error) {
	err := s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(usersBucket))

		// Check if there is the user with
		// the same nickname as new user.
		// TODO(thinkofher) Store nicknames in another bucket
		// and change O(n) time of checking if nickname is occupied
		// to O(1).
		err := forEachUser(tx, func(user storage.UserEntry) error {
			if user.Nickname == newUser.Nickname {
				return serrors.ErrNicknameTaken
			}
			return nil
		})
		if err != nil {
			return err
		}

		return storeUserInBucket(newUser, b)
	})
	if err != nil {
		return "", err
	}

	return newUser.ID, nil
}

// Read returns single user data with given ID.
func (s *UsersStorage) Read(ctx context.Context, id string) (*storage.UserEntry, error) {
	user := new(storage.UserEntry)

	err := s.db.View(func(tx *bolt.Tx) error {
		var err error
		user, err = readUser(tx, id)
		return err
	})
	if err != nil {
		return nil, fmt.Errorf("reading user with id=%s failed: %w", id, err)
	}

	return user, nil
}

func forEachUser(tx *bolt.Tx, f func(storage.UserEntry) error) error {
	b := tx.Bucket([]byte(usersBucket))
	return b.ForEach(func(k, v []byte) error {
		userBucket := b.Bucket(k)
		if userBucket == nil {
			return serrors.ErrNoID
		}

		user, err := userFromBucket(userBucket)
		if err != nil {
			return err
		}

		return f(*user)
	})

}

// All returns slice with all users from storage.
func (s *UsersStorage) All(ctx context.Context) ([]storage.UserEntry, error) {
	res := []storage.UserEntry{}

	err := s.db.View(func(tx *bolt.Tx) error {
		return forEachUser(tx, func(u storage.UserEntry) error {
			res = append(res, u)
			return nil
		})
	})
	if err != nil {
		return nil, fmt.Errorf("reading all users failed: %w", err)
	}

	return res, nil
}

// updateOne updates one user in bucket.
func updateOneUser(b *bolt.Bucket, u storage.UserEntry) error {
	// Check if there is user with given id in database.
	if b.Bucket(userBucketKey(u.ID)) == nil {
		return serrors.ErrNoID
	}

	return storeUserInBucket(u, b)
}

// Update overwrites existing user data.
func (s *UsersStorage) Update(ctx context.Context, id string, f func(*storage.UserEntry) error) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		u, err := readUser(tx, id)
		if err != nil {
			return fmt.Errorf("readUser: %w", err)
		}

		if err := f(u); err != nil {
			return err
		}

		return updateOneUser(tx.Bucket([]byte(usersBucket)), *u)
	})
}

// Remove deletes user with given id from storage.
func (s *UsersStorage) Remove(ctx context.Context, id string) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(usersBucket))

		key := userBucketKey(id)
		// Check if there is user with given id in database.
		if b.Bucket(key) == nil {
			return serrors.ErrNoID
		}

		return b.DeleteBucket(key)
	})
}

// DevicesStorage implements storage.Devices interface
// for bolt database.
type DevicesStorage struct {
	db *bolt.DB
}

const (
	deviceIDKey      = "ls::device::id"
	deviceTagKey     = "ls::device::tag"
	deviceOwnerKey   = "ls::device::owner"
	deviceOwnerIDKey = "ls::device::owner::id"
	deviceMACKey     = "ls::device::mac"
)

func deviceBucketKey(id string) []byte {
	return []byte(fmt.Sprintf("ls::device::%s", id))
}

func deviceFromBucket(b *bolt.Bucket) (*models.Device, error) {
	result := new(models.Device)

	fail := func() (*models.Device, error) {
		return nil, serrors.ErrNoID
	}

	deviceIDBytes := b.Get([]byte(deviceIDKey))
	if deviceIDBytes == nil {
		return fail()
	}
	result.ID = string(deviceIDBytes)

	ownerIDBytes := b.Get([]byte(deviceOwnerIDKey))
	if ownerIDBytes == nil {
		return fail()
	}
	result.OwnerID = string(ownerIDBytes)

	tag := b.Get([]byte(deviceTagKey))
	if tag == nil {
		return fail()
	}
	result.Tag = string(tag)

	mac := b.Get([]byte(deviceMACKey))
	if mac == nil {
		return fail()
	}
	result.MAC = mac

	owner := b.Get([]byte(deviceOwnerKey))
	if owner == nil {
		return fail()
	}
	result.Owner = string(owner)

	return result, nil
}

func storeDeviceInBucket(device models.Device, b *bolt.Bucket) error {
	// TODO(thinkofher) Wrap errors with fmt.Errorf and "%w".
	deviceBucket, err := b.CreateBucketIfNotExists(deviceBucketKey(device.ID))
	if err != nil {
		return err
	}

	deviceID := []byte(device.ID)
	ownerID := []byte(device.OwnerID)

	// keys and values for device data model
	kvs := []bucketMapping{
		{[]byte(deviceIDKey), deviceID},
		{[]byte(deviceOwnerIDKey), ownerID},
		{[]byte(deviceOwnerKey), []byte(device.Owner)},
		{[]byte(deviceTagKey), []byte(device.Tag)},
		{[]byte(deviceMACKey), device.MAC},
	}

	for _, item := range kvs {
		if err := deviceBucket.Put(item.key, item.value); err != nil {
			return err
		}
	}

	return nil
}

func forEachDevice(tx *bolt.Tx, f func(models.Device) error) error {
	b := tx.Bucket([]byte(devicesBucket))
	return b.ForEach(func(k, v []byte) error {
		deviceBucket := b.Bucket(k)
		if deviceBucket == nil {
			return serrors.ErrNoID
		}

		device, err := deviceFromBucket(deviceBucket)
		if err != nil {
			return err
		}

		return f(*device)
	})

}

func sameDevice(a, b models.Device) bool {
	return a.Owner == b.Owner && a.Tag == b.Tag
}

// userExists checks whether user with given id is stored in database.
func userExists(tx *bolt.Tx, userID string) (bool, error) {
	user, err := readUser(tx, userID)
	if err != nil && !errors.Is(err, serrors.ErrNoID) {
		return false, err
	}
	return user != nil, nil
}

// New stores given devices data in database and returns
// new device (given in new model) id.
func (d *DevicesStorage) New(ctx context.Context, userID string, newDevice models.Device) (string, error) {
	err := d.db.Update(func(tx *bolt.Tx) error {
		// Check if there is user with given id.
		exists, err := userExists(tx, userID)
		if err != nil {
			return err
		}
		if !exists {
			return fmt.Errorf("there is no user with id=%s, err=%w", userID, serrors.ErrNoID)
		}

		// Check if there is the device with same owner and tag.
		err = forEachDevice(tx, func(device models.Device) error {
			if sameDevice(newDevice, device) {
				return serrors.ErrDeviceDuplication
			}
			return nil
		})
		if err != nil {
			return err
		}

		b := tx.Bucket([]byte(devicesBucket))
		return storeDeviceInBucket(newDevice, b)
	})
	if err != nil {
		return "", err
	}

	return newDevice.ID, nil
}

// NewByOwner stores given devices (owned by user with given nickname) data
// in database and returns its id.
func (d *DevicesStorage) NewByOwner(ctx context.Context, deviceOwner string, newDevice models.Device) (string, error) {
	err := d.db.Update(func(tx *bolt.Tx) error {
		// FIXME(thinkofher) If there is no user with given nickname
		// zero-valued user will be used.
		var targetUser storage.UserEntry
		err := forEachUser(tx, func(u storage.UserEntry) error {
			if u.Nickname == deviceOwner {
				targetUser = u
			}
			return nil
		})
		if err != nil {
			return err
		}

		// Check if there is the device with same owner and tag.
		b := tx.Bucket([]byte(devicesBucket))
		err = forEachDevice(tx, func(device models.Device) error {
			if sameDevice(device, newDevice) {
				return serrors.ErrDeviceDuplication
			}
			return nil
		})
		if err != nil {
			return err
		}

		newDevice.OwnerID = targetUser.ID
		newDevice.Owner = targetUser.Nickname

		return storeDeviceInBucket(newDevice, b)
	})
	if err != nil {
		return "", err
	}

	return newDevice.ID, nil
}

func forDevicesOfUser(tx *bolt.Tx, userID string, f func(models.Device) error) error {
	// Check if there is the device with same owner and tag.
	return forEachDevice(tx, func(device models.Device) error {
		if device.OwnerID == userID {
			return f(device)
		}
		return nil
	})
}

// OfUser returns list of models owned by user with given id.
func (d *DevicesStorage) OfUser(ctx context.Context, userID string) ([]models.Device, error) {
	ans := []models.Device{}

	err := d.db.View(func(tx *bolt.Tx) error {
		return forDevicesOfUser(tx, userID, func(d models.Device) error {
			ans = append(ans, d)
			return nil
		})
	})

	if err != nil {
		return ans, err
	}

	return ans, nil
}

// Read returns single device data with given ID.
func (s *DevicesStorage) Read(ctx context.Context, id string) (*models.Device, error) {
	device := new(models.Device)

	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(devicesBucket))
		bucketKey := deviceBucketKey(id)
		deviceBucket := b.Bucket(bucketKey)
		if deviceBucket == nil {
			return serrors.ErrNoID
		}

		var err error
		device, err = deviceFromBucket(deviceBucket)
		return err
	})
	if err != nil {
		return nil, fmt.Errorf("reading device with id=%s failed: %w", id, err)
	}

	return device, nil
}

// All returns slice with all devices from storage.
func (s *DevicesStorage) All(ctx context.Context) ([]models.Device, error) {
	res := []models.Device{}

	err := s.db.View(func(tx *bolt.Tx) error {
		return forEachDevice(tx, func(device models.Device) error {
			res = append(res, device)
			return nil
		})
	})
	if err != nil {
		return nil, fmt.Errorf("reading all devices failed: %w", err)
	}

	return res, nil
}

// Update overwrites existing device data.
func (s *DevicesStorage) Update(ctx context.Context, d models.Device) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(devicesBucket))

		// Check if there is device with given id in database.
		if b.Bucket(deviceBucketKey(d.ID)) == nil {
			return serrors.ErrNoID
		}

		return storeDeviceInBucket(d, b)
	})
}

// Remove deletes device with given id from storage.
func (s *DevicesStorage) Remove(ctx context.Context, id string) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(devicesBucket))

		key := deviceBucketKey(id)
		// Check if there is device with given id in database.
		if b.Bucket(key) == nil {
			return serrors.ErrNoID
		}

		return b.DeleteBucket(key)
	})
}

const (
	onlineUsersCounter    = "ls::users::online::counter"
	unknownDevicesCounter = "ls::devices::unknown::counter"
)

// status implements storage.Status interface.
//
// status is able to perform multiple operation, but only in
// single transaction, because it holds pointer to bolt.Tx instead
// of pointer to bolt.DB as other types in this package.
type status struct {
	tx *bolt.Tx
}

// OnlineUsers returns number of people being currently online.
func (s *status) OnlineUsers(ctx context.Context) (int, error) {
	b, err := s.tx.CreateBucketIfNotExists([]byte(countersBucket))
	if err != nil {
		return 0, fmt.Errorf("failed to create %s bucket: %w", countersBucket, err)
	}

	onlineUsers := b.Get([]byte(onlineUsersCounter))
	if onlineUsers == nil {
		return 0, fmt.Errorf("failed to retrieve online users counter")
	}

	parsedOnlineUsers, err := strconv.Atoi(string(onlineUsers))
	if err != nil {
		return 0, fmt.Errorf(
			"failed to parse slice bytes: %s into integer: %w",
			onlineUsers, err,
		)
	}

	return parsedOnlineUsers, nil
}

// SetOnlineUsers ovewrites number of people being currently online.
func (s *status) SetOnlineUsers(ctx context.Context, number int) error {
	b, err := s.tx.CreateBucketIfNotExists([]byte(countersBucket))
	if err != nil {
		return fmt.Errorf("failed to create %s bucket: %w", countersBucket, err)
	}

	return b.Put([]byte(onlineUsersCounter), []byte(strconv.Itoa(number)))
}

// UnknownDevices returns number of unknown devices connected to the network.
func (s *status) UnknownDevices(ctx context.Context) (int, error) {
	b, err := s.tx.CreateBucketIfNotExists([]byte(countersBucket))
	if err != nil {
		return 0, fmt.Errorf("failed to create %s bucket: %w", countersBucket, err)
	}

	unknownDevices := b.Get([]byte(unknownDevicesCounter))
	if unknownDevices == nil {
		return 0, fmt.Errorf("failed to retrieve unknown devices counter")
	}

	parsedUnknownDevices, err := strconv.Atoi(string(unknownDevices))
	if err != nil {
		return 0, fmt.Errorf(
			"failed to parse slice bytes: %s into integer: %w",
			unknownDevices, err,
		)
	}

	return parsedUnknownDevices, nil
}

// SetUnknownDevices overwrites number of unknown devices connected to the network.
func (s *status) SetUnknownDevices(ctx context.Context, number int) error {
	b, err := s.tx.CreateBucketIfNotExists([]byte(countersBucket))
	if err != nil {
		return fmt.Errorf("failed to create %s bucket: %w", countersBucket, err)
	}

	return b.Put([]byte(unknownDevicesCounter), []byte(strconv.Itoa(number)))
}

// StatusStorageTx implements storage.StatusTx interface.
type StatusStorageTx struct {
	db *bolt.DB
}

// DevicesStatus accepts function that manipulates number of
// unknown devices and online users in single safe transaction.
func (s *StatusStorageTx) DevicesStatus(
	ctx context.Context,
	f func(context.Context, storage.Status) error,
) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		statusStorage := &status{tx}
		return f(ctx, statusStorage)
	})
}

// TwoFactorStorage implements TwoFactor storage interface
// for bolt database.
type TwoFactorStorage struct {
	db *bolt.DB
}

func twoFactorKey(userID string) []byte {
	return []byte(fmt.Sprintf("%s::%s", twoFactorBucket, userID))
}

func getTwoFactorBucket(tx *bolt.Tx) *bolt.Bucket {
	return tx.Bucket([]byte(twoFactorBucket))
}

func getTwoFactorMethods(tx *bolt.Tx, userID string) *models.TwoFactor {
	bucket := getTwoFactorBucket(tx)
	dat := bucket.Get(twoFactorKey(userID))
	res := &models.TwoFactor{
		OneTimeCodes:  map[string]models.OneTimeCode{},
		RecoveryCodes: map[string]models.Recovery{},
	}
	if dat == nil {
		// User has not two factor methods so we can
		// return empty entry of TwoFactor.
		return res
	}
	if err := json.Unmarshal(dat, res); err != nil {
		return &models.TwoFactor{
			OneTimeCodes:  map[string]models.OneTimeCode{},
			RecoveryCodes: map[string]models.Recovery{},
		}
	}
	return res
}

func setTwoFactorMethods(tx *bolt.Tx, userID string, tf models.TwoFactor) error {
	bucket := getTwoFactorBucket(tx)
	dat, err := json.Marshal(tf)
	if err != nil {
		return fmt.Errorf("json.Marshal: %w", err)
	}
	return bucket.Put(twoFactorKey(userID), dat)
}

// Get returns two factor methods for user with given user ID.
func (t *TwoFactorStorage) Get(ctx context.Context, userID string) (*models.TwoFactor, error) {
	res := new(models.TwoFactor)
	err := t.db.View(func(tx *bolt.Tx) error {
		_, err := readUser(tx, userID)
		if err != nil {
			return serrors.ErrNoID
		}
		res = getTwoFactorMethods(tx, userID)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return res, nil
}

// Updates apply given function to two factor methods of user
// with given user ID.
func (t *TwoFactorStorage) Update(ctx context.Context, userID string, f func(*models.TwoFactor) error) error {
	return t.db.Update(func(tx *bolt.Tx) error {
		_, err := readUser(tx, userID)
		if err != nil {
			return serrors.ErrNoID
		}
		res := getTwoFactorMethods(tx, userID)
		if err = f(res); err != nil {
			return err
		}
		return setTwoFactorMethods(tx, userID, *res)
	})
}

// Remove deletes all two factor methods of user with given
// user ID. You can still use Update method after all to start
// adding more methods.
func (t *TwoFactorStorage) Remove(ctx context.Context, userID string) error {
	return t.db.Update(func(tx *bolt.Tx) error {
		_, err := readUser(tx, userID)
		if err != nil {
			return serrors.ErrNoID
		}
		bucket := getTwoFactorBucket(tx)
		return bucket.Delete(twoFactorKey(userID))
	})
}
