// Package memory implements storage interfaces for
// bolt key-value store database.
package memory

import (
	"bytes"
	"context"
	"encoding/gob"
	"fmt"
	"strconv"

	bolt "go.etcd.io/bbolt"

	"github.com/hakierspejs/long-season/pkg/models"
	serrors "github.com/hakierspejs/long-season/pkg/storage/errors"
)

const (
	countersBucket       = "ls::counters"
	usersBucket          = "ls::users"
	usersBucketCounter   = "ls::users::counter"
	devicesBucket        = "ls::devices"
	devicesBucketCounter = "ls::devices::counter"
)

// Factory implements storage.Factory interface for
// bolt database.
type Factory struct {
	users   *UsersStorage
	devices *DevicesStorage
}

// Users returns storage interface for manipulating
// users data.
func (f Factory) Users() *UsersStorage {
	return f.users
}

func (f Factory) Devices() *DevicesStorage {
	return f.devices
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
		users:   &UsersStorage{db},
		devices: &DevicesStorage{db},
	}, nil
}

// UsersStorage implements storage.Users interface
// for bolt database.
type UsersStorage struct {
	db *bolt.DB
}

func incr(tx *bolt.Tx, key []byte) (int, error) {
	counterBucket, err := tx.CreateBucketIfNotExists([]byte(countersBucket))
	if err != nil {
		return 0, fmt.Errorf("cannot create new bucket when incrementing %s: %w", key, err)
	}

	val := counterBucket.Get([]byte(key))

	counter := 0

	if val != nil {
		counter, err = strconv.Atoi(string(val))
		if err != nil {
			return counter, fmt.Errorf("conversion %s to int failed: %w", val, err)
		}
	}

	err = counterBucket.Put([]byte(key), []byte(strconv.Itoa(counter+1)))
	if err != nil {
		return counter, fmt.Errorf("cannot put counter into bucket: %w", err)
	}

	return counter, nil
}

const (
	userIDKey       = "ls::user::id"
	userNicknameKey = "ls::user::nickname"
	userPasswordKey = "ls::user::password"
	userOnlineKey   = "ls::user::online"
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

	id, err := strconv.Atoi(string(idBytes))
	if err != nil {
		return nil, err
	}
	result.ID = id

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

	return result, nil
}

func userBucketKey(id int) []byte {
	return []byte(fmt.Sprintf("ls::user::%d", id))
}

func storeUserInBucket(user models.User, b *bolt.Bucket) error {
	userBucket, err := b.CreateBucketIfNotExists(userBucketKey(user.ID))
	if err != nil {
		return err
	}

	id := []byte(strconv.Itoa(user.ID))
	if err := userBucket.Put([]byte(userIDKey), id); err != nil {
		return err
	}

	if err := userBucket.Put([]byte(userNicknameKey), []byte(user.Nickname)); err != nil {
		return err
	}

	if err := userBucket.Put([]byte(userPasswordKey), user.Password); err != nil {
		return err
	}

	return userBucket.Put([]byte(userOnlineKey), boolToBytes(user.Online))
}

// New stores given user data in database and returns
// assigned id.
func (s *UsersStorage) New(ctx context.Context, newUser models.User) (int, error) {
	var id int

	err := s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(usersBucket))

		// Check if there is the user with
		// the same nickname as new user.
		// TODO(thinkofher) Store nicknames in another bucket
		// and change O(n) time of checking if nickname is occupied
		// to O(1).
		err := b.ForEach(func(k, v []byte) error {
			userBucket := b.Bucket(k)

			user, err := userFromBucket(userBucket)
			if err != nil {
				return err
			}

			if user.Nickname == newUser.Nickname {
				return serrors.ErrNicknameTaken
			}

			return nil
		})
		if err != nil {
			return err
		}

		id, err = incr(tx, []byte(usersBucketCounter))
		if err != nil {
			return err
		}
		newUser.ID = id

		return storeUserInBucket(newUser, b)
	})
	if err != nil {
		return 0, err
	}

	return id, nil
}

// Read returns single user data with given ID.
func (s *UsersStorage) Read(ctx context.Context, id int) (*models.User, error) {
	user := new(models.User)

	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(usersBucket))
		bucketKey := userBucketKey(id)
		userBucket := b.Bucket(bucketKey)
		if userBucket == nil {
			return serrors.ErrNoID
		}

		var err error
		user, err = userFromBucket(userBucket)
		return err
	})
	if err != nil {
		return nil, fmt.Errorf("reading user with id=%d failed: %w", id, err)
	}

	return user, nil
}

// All returns slice with all users from storage.
func (s *UsersStorage) All(ctx context.Context) ([]models.User, error) {
	res := []models.User{}

	err := s.db.View(func(tx *bolt.Tx) error {
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

			res = append(res, *user)
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
func (s *UsersStorage) Update(ctx context.Context, u models.User) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(usersBucket))
		return updateOneUser(b, u)
	})
}

// UpdateMany overwrites data of all users in given slice.
func (s *UsersStorage) UpdateMany(ctx context.Context, u []models.User) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(usersBucket))

		for _, user := range u {
			err := updateOneUser(b, user)
			if err != nil {
				return err
			}

		}
		return nil
	})
}

// Remove deletes user with given id from storage.
func (s *UsersStorage) Remove(ctx context.Context, id int) error {
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

// New stores given devices data in database and returns
// assigned id.
func (d *DevicesStorage) New(ctx context.Context, userID int, newDevice models.Device) (int, error) {
	var id int

	// TODO(thinkofher) Check if there is user with given id.
	err := d.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(devicesBucket))

		// Check if there is the device with same owner and tag.
		err := b.ForEach(func(k, v []byte) error {
			buff := bytes.NewBuffer(v)

			var device models.Device

			// Check if given key is an integer.
			if _, err := strconv.Atoi(string(k)); err == nil {
				err := gob.NewDecoder(buff).Decode(&device)
				if err != nil {
					return err
				}

				if device.Owner == newDevice.Owner && device.Tag == newDevice.Tag {
					return serrors.ErrDeviceDuplication
				}
			}

			return nil
		})
		if err != nil {
			return err
		}

		id, err = incr(tx, []byte(devicesBucketCounter))
		if err != nil {
			return err
		}
		newDevice.ID = id

		buff := bytes.NewBuffer([]byte{})
		err = gob.NewEncoder(buff).Encode(&newDevice)
		if err != nil {
			return err
		}

		return b.Put([]byte(strconv.Itoa(id)), buff.Bytes())
	})
	if err != nil {
		return 0, err
	}

	return id, nil
}

// OfUser returns list of models owned by user with given id.
func (d *DevicesStorage) OfUser(ctx context.Context, userID int) ([]models.Device, error) {
	ans := []models.Device{}

	err := d.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(devicesBucket))

		// Check if there is the device with same owner and tag.
		err := b.ForEach(func(k, v []byte) error {
			buff := bytes.NewBuffer(v)

			var device models.Device

			// Check if given key is an integer.
			if _, err := strconv.Atoi(string(k)); err == nil {
				err := gob.NewDecoder(buff).Decode(&device)
				if err != nil {
					return err
				}

				if device.OwnerID == userID {
					ans = append(ans, device)
				}
			}

			return nil
		})

		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return ans, err
	}

	return ans, nil
}

// Read returns single device data with given ID.
func (s *DevicesStorage) Read(ctx context.Context, id int) (*models.Device, error) {
	device := new(models.Device)

	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(devicesBucket))
		data := b.Get([]byte(strconv.Itoa(id)))

		if data == nil {
			return serrors.ErrNoID
		}

		// TODO(thinkofher) You can move process of decoding outside
		// of View function to unlock writing to database.
		buff := bytes.NewBuffer(data)
		return gob.NewDecoder(buff).Decode(device)
	})
	if err != nil {
		return nil, fmt.Errorf("reading device with id=%d failed: %w", id, err)
	}

	return device, nil
}

// All returns slice with all devices from storage.
func (s *DevicesStorage) All(ctx context.Context) ([]models.Device, error) {
	res := []models.Device{}

	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(devicesBucket))

		return b.ForEach(func(k, v []byte) error {
			device := new(models.Device)
			buff := bytes.NewBuffer(v)

			// Check if given key is an integer.
			if _, err := strconv.Atoi(string(k)); err == nil {
				err := gob.NewDecoder(buff).Decode(device)
				if err != nil {
					return fmt.Errorf("decoding device from gob failed: %w", err)
				}

				res = append(res, *device)
			}

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
		if b.Get([]byte(strconv.Itoa(d.ID))) == nil {
			return serrors.ErrNoID
		}

		buff := bytes.NewBuffer([]byte{})
		err := gob.NewEncoder(buff).Encode(&d)
		if err != nil {
			return err
		}

		return b.Put([]byte(strconv.Itoa(d.ID)), buff.Bytes())
	})
}

// Remove deletes device with given id from storage.
func (s *DevicesStorage) Remove(ctx context.Context, id int) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(devicesBucket))

		// Check if there is user with given id in database.
		if b.Get([]byte(strconv.Itoa(id))) == nil {
			return serrors.ErrNoID
		}

		return b.Delete([]byte(strconv.Itoa(id)))
	})
}
