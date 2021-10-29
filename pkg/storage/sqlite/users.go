package sqlite

import (
	"context"

	"github.com/hakierspejs/long-season/pkg/storage"
)

// Users storage implements storage.Users interface for
// sqlite database.
type Users struct {
	cs *coreStorage
}

func (s *Users) New(ctx context.Context, u storage.UserEntry) (string, error) {
	return s.cs.newUser(ctx, u)
}

func (s *Users) Read(ctx context.Context, id string) (*storage.UserEntry, error) {
	return s.cs.readUser(ctx, id)
}

func (s *Users) All(ctx context.Context) ([]storage.UserEntry, error) {
	return s.cs.allUsers(ctx)
}

func (s *Users) Remove(ctx context.Context, id string) error {
	return s.cs.removeUser(ctx, id)
}

func (s *Users) Update(ctx context.Context, id string, f func(*storage.UserEntry) error) error {
	return s.cs.updateUser(ctx, id, f)
}
