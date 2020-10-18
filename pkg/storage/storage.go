package storage

import (
	"context"

	"github.com/hakierspejs/long-season/pkg/models"
)

// Factory returns interfaces specific to
// stored data.
type Factory interface {
	Users(ctx context.Context) Users
}

// Users interface handles generic create, read,
// update and delete operations on users data.
type Users interface {
	New(ctx context.Context, u models.User) error
	Read(ctx context.Context, id int) (*models.User, error)
	All(ctx context.Context) ([]*models.User, error)
	Update(ctx context.Context, u models.User) error
	Remove(ctx context.Context, id int) error
}
