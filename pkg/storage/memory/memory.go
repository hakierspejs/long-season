// Package memory implements storage interfaces for
// bolt key-value store database.
package memory

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	bolt "go.etcd.io/bbolt"

	"github.com/google/uuid"
	"github.com/hakierspejs/long-season/pkg/models"
	"github.com/hakierspejs/long-season/pkg/storage"
	serrors "github.com/hakierspejs/long-season/pkg/storage/errors"
)

const (
	countersBucket       = "ls::counters"
	usersBucket          = "ls::users"
	devicesBucket        = "ls::devices"
	devicesBucketCounter = "ls::devices::counter"
)

// Factory implements storage.Factory interface for
// bolt database.
type Factory struct {
	users           *UsersStorage
	devices         *DevicesStorage
	statusIter      *StatusIterator
	statusStorageTx *StatusStorageTx
}

// Users returns storage interface for manipulating
// users data.
func (f Factory) Users() *UsersStorage {
	return f.users
}

func (f Factory) Devices() *DevicesStorage {
	return f.devices
}

// StatusIterator returns storage interface for
// iterating over users and their devices.
func (f Factory) StatusIterator() *StatusIterator {
	return f.statusIter
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
		statusIter:      &StatusIterator{db},
		statusStorageTx: &StatusStorageTx{db},
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
	userOnlineKey      = "ls::user::online"
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

func userFromBucket(b *bolt.Bucket) (*models.User, error) {
	result := new(models.User)

	fail := func() (*models.User, error) {
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
	result.Password = password

	online := b.Get([]byte(userOnlineKey))
	if online == nil {
		return fail()
	}
	result.Online = bytesToBool(online)

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
func storeUserInBucket(user models.User, b *bolt.Bucket) error {
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
		{[]byte(userPasswordKey), user.Password},
		{[]byte(userOnlineKey), boolToBytes(user.Online)},
		{[]byte(userPrivateModeKey), boolToBytes(user.Private)},
	}

	for _, item := range kvs {
		if err := userBucket.Put(item.key, item.value); err != nil {
			return err
		}
	}

	return nil
}

func readUser(tx *bolt.Tx, userID string) (*models.User, error) {
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
// assigned id.
func (s *UsersStorage) New(ctx context.Context, newUser models.User) (string, error) {
	var id string

	err := s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(usersBucket))

		// Check if there is the user with
		// the same nickname as new user.
		// TODO(thinkofher) Store nicknames in another bucket
		// and change O(n) time of checking if nickname is occupied
		// to O(1).
		err := forEachUser(tx, func(user models.User) error {
			if user.Nickname == newUser.Nickname {
				return serrors.ErrNicknameTaken
			}
			return nil
		})
		if err != nil {
			return err
		}

		id = uuid.New().String()
		newUser.ID = id
		return storeUserInBucket(newUser, b)
	})
	if err != nil {
		return "", err
	}

	return id, nil
}

// Read returns single user data with given ID.
func (s *UsersStorage) Read(ctx context.Context, id string) (*models.User, error) {
	user := new(models.User)

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

func forEachUser(tx *bolt.Tx, f func(models.User) error) error {
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
func (s *UsersStorage) All(ctx context.Context) ([]models.User, error) {
	res := []models.User{}

	err := s.db.View(func(tx *bolt.Tx) error {
		return forEachUser(tx, func(u models.User) error {
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
func updateOneUser(b *bolt.Bucket, u models.User) error {
	// Check if there is user with given id in database.
	if b.Bucket(userBucketKey(u.ID)) == nil {
		return serrors.ErrNoID
	}

	return storeUserInBucket(u, b)
}

// Update overwrites existing user data.
func (s *UsersStorage) Update(ctx context.Context, id string, f func(*models.User) error) error {
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
// assigned id.
func (d *DevicesStorage) New(ctx context.Context, userID string, newDevice models.Device) (string, error) {
	var id string

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

		id = uuid.New().String()
		newDevice.ID = id

		b := tx.Bucket([]byte(devicesBucket))
		return storeDeviceInBucket(newDevice, b)
	})
	if err != nil {
		return "", err
	}

	return id, nil
}

// NewByOwner stores given devices (owned by user with given nickname) data
// in database and returns assigned id.
func (d *DevicesStorage) NewByOwner(ctx context.Context, deviceOwner string, newDevice models.Device) (string, error) {
	var id string

	err := d.db.Update(func(tx *bolt.Tx) error {
		// FIXME(thinkofher) If there is no user with given nickname
		// zero-valued user will be used.
		var targetUser models.User
		err := forEachUser(tx, func(u models.User) error {
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

		id = uuid.New().String()
		newDevice.ID = id
		newDevice.OwnerID = targetUser.ID
		newDevice.Owner = targetUser.Nickname

		return storeDeviceInBucket(newDevice, b)
	})
	if err != nil {
		return "", err
	}

	return id, nil
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

// StatusIterator implements storage.StatusIterator interface.
type StatusIterator struct {
	db *bolt.DB
}

// ForEachUpdate iterates over every user from database and their
// devices. Overwrites current user data with returned one.
func (s *StatusIterator) ForEachUpdate(
	ctx context.Context,
	iterFunc func(models.User, []models.Device) (*models.User, error),
) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		return forEachUser(tx, func(u models.User) error {
			userDevices := []models.Device{}

			err := forDevicesOfUser(tx, u.ID, func(d models.Device) error {
				userDevices = append(userDevices, d)
				return nil
			})
			if err != nil {
				return err
			}

			newUser, err := iterFunc(u, userDevices)
			if err != nil {
				return err
			}

			b := tx.Bucket([]byte(usersBucket))
			return updateOneUser(b, *newUser)
		})
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
