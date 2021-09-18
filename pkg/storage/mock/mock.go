// Package mock implements storage interfaces that
// can be used in unit tests or for running long-season
// with volatile data storage.
package mock

import (
	"context"
	"sync"

	"github.com/google/uuid"
	"github.com/hakierspejs/long-season/pkg/models"
	serrors "github.com/hakierspejs/long-season/pkg/storage/errors"
)

// Factory returns mock interfaces specific
// to stored data. Implements storage.Factory interface.
type Factory struct {
	users *UsersStorage
}

// New returns new mock factory.
func New() *Factory {
	return &Factory{
		users: newUserStorage(),
	}
}

// Users returns storage interface for manipulating
// users data.
func (f Factory) Users() *UsersStorage {
	return f.users
}

// UsersStorage implements storage.Users interface
// for mocking purposes.
type UsersStorage struct {
	data  map[string]models.User
	mutex *sync.Mutex
}

func newUserStorage() *UsersStorage {
	return &UsersStorage{
		data:  make(map[string]models.User),
		mutex: new(sync.Mutex),
	}
}

// New stores given user data in database and returns
// assigned id.
func (s *UsersStorage) New(ctx context.Context, newUser models.User) (string, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	for _, u := range s.data {
		if u.Nickname == newUser.Nickname {
			return "", serrors.ErrNicknameTaken
		}
	}

	newUser.ID = uuid.New().String()

	s.data[newUser.ID] = newUser

	return newUser.ID, nil
}

// Read returns single user data with given ID.
func (s *UsersStorage) Read(ctx context.Context, id string) (*models.User, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	user, ok := s.data[id]
	if !ok {
		return nil, serrors.ErrNoID
	}

	return &user, nil
}

// All returns slice with all users from storage.
func (s *UsersStorage) All(ctx context.Context) ([]models.User, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	res := []models.User{}
	for _, u := range s.data {
		res = append(res, u)
	}

	return res, nil
}

// Update overwrites existing user data.
func (s *UsersStorage) Update(ctx context.Context, u models.User) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	_, ok := s.data[u.ID]
	if !ok {
		return serrors.ErrNoID
	}

	s.data[u.ID] = u
	return nil
}

// UpdateMany overwrites data of all users in given slice.
func (s *UsersStorage) UpdateMany(ctx context.Context, u []models.User) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Loop for the first time to ensure that all users
	// have valid ID to make this operation atomic.
	for _, user := range u {
		_, ok := s.data[user.ID]
		if !ok {
			return serrors.ErrNoID
		}
	}

	// Loop one more time to add users.
	for _, user := range u {
		s.data[user.ID] = user
	}

	return nil
}

// Remove deletes user with given id from storage.
func (s *UsersStorage) Remove(ctx context.Context, id string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	_, ok := s.data[id]
	if !ok {
		return serrors.ErrNoID
	}

	delete(s.data, id)

	return nil

}
