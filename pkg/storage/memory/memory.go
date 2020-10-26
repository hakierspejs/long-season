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
	countersBucket     = "ls::counters"
	usersBucket        = "ls::users"
	usersBucketCounter = "ls::users::counter"
)

// Factory implements storage.Factory interface for
// bolt database.
type Factory struct {
	users *UsersStorage
}

// Users returns storage interface for manipulating
// users data.
func (f Factory) Users() *UsersStorage {
	return f.users
}

// New returns pointer to new memory storage
// Factory.
func New(db *bolt.DB) (*Factory, error) {

	err := db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(usersBucket))
		return err
	})
	if err != nil {
		return nil, err
	}

	return &Factory{
		users: &UsersStorage{db},
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
			buff := bytes.NewBuffer(v)

			var user models.User

			// Check if given key is an integer.
			if _, err := strconv.Atoi(string(k)); err == nil {
				err := gob.NewDecoder(buff).Decode(&user)
				if err != nil {
					return err
				}

				if user.Nickname == newUser.Nickname {
					return serrors.ErrNicknameTaken(newUser.Nickname)
				}
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

		buff := bytes.NewBuffer([]byte{})
		err = gob.NewEncoder(buff).Encode(&newUser)
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

// Read returns single user data with given ID.
func (s *UsersStorage) Read(ctx context.Context, id int) (*models.User, error) {
	user := new(models.User)

	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(usersBucket))
		data := b.Get([]byte(strconv.Itoa(id)))

		if data == nil {
			return serrors.ErrNoID(id)
		}

		// TODO(thinkofher) You can move process of decoding outside
		// of View function to unlock writing to database.
		buff := bytes.NewBuffer(data)
		return gob.NewDecoder(buff).Decode(user)
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
			user := new(models.User)
			buff := bytes.NewBuffer(v)

			// Check if given key is an integer.
			if _, err := strconv.Atoi(string(k)); err == nil {
				err := gob.NewDecoder(buff).Decode(user)
				if err != nil {
					return fmt.Errorf("decoding user from gob failed: %w", err)
				}

				res = append(res, *user)
			}

			return nil
		})
	})
	if err != nil {
		return nil, fmt.Errorf("reading all users failed: %w", err)
	}

	return res, nil
}

// Update overwrites existing user data.
func (s *UsersStorage) Update(ctx context.Context, u models.User) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(usersBucket))

		// Check if there is user with given id in database.
		if b.Get([]byte(strconv.Itoa(u.ID))) == nil {
			return serrors.ErrNoID(u.ID)
		}

		buff := bytes.NewBuffer([]byte{})
		err := gob.NewEncoder(buff).Encode(&u)
		if err != nil {
			return err
		}

		return b.Put([]byte(strconv.Itoa(u.ID)), buff.Bytes())
	})
}

// Remove deletes user with given id from storage.
func (s *UsersStorage) Remove(ctx context.Context, id int) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(usersBucket))

		// Check if there is user with given id in database.
		if b.Get([]byte(strconv.Itoa(id))) == nil {
			return serrors.ErrNoID(id)
		}

		return b.Delete([]byte(strconv.Itoa(id)))
	})
}
