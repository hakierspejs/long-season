package storage

import (
	"context"

	"github.com/hakierspejs/long-season/pkg/models"
)

// Factory returns interfaces specific to
// stored data.
type Factory interface {
	Users() Users
	Devices() Devices
}

// Users interface handles generic create, read,
// update and delete operations on users data.
type Users interface {
	// New stores given user data in database and returns
	// assigned id.
	New(ctx context.Context, u models.User) (int, error)
	Read(ctx context.Context, id int) (*models.User, error)
	All(ctx context.Context) ([]models.User, error)
	Update(ctx context.Context, u models.User) error
	UpdateMany(ctx context.Context, u []models.User) error
	Remove(ctx context.Context, id int) error
}

type Devices interface {
	New(ctx context.Context, userID int, d models.Device) (int, error)
	OfUser(ctx context.Context, userID int) ([]models.Device, error)
	Read(ctx context.Context, id int) (*models.Device, error)
	All(ctx context.Context) ([]models.Device, error)
	Update(ctx context.Context, d models.Device) error
	Remove(ctx context.Context, id int) error
}
