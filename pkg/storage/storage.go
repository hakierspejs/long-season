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

// Status interface provides methods for reading and
// writing numerical information about users and devices
// spending time in hackerspace.
type Status interface {
	// OnlineUsers returns number of people being
	// currently online.
	OnlineUsers(ctx context.Context) (int, error)

	// SetOnlineUsers ovewrites number of people being
	// currently online.
	SetOnlineUsers(ctx context.Context, number int) error

	// UnknownDevices returns number of unknown devices
	// connected to the network.
	UnknownDevices(ctx context.Context) (int, error)

	// SetUnknownDevices overwrites number of unknown devices
	// connected to the network.
	SetUnknownDevices(ctx context.Context, number int) error
}

// StatusTx interface provides methods for reading and
// writing numerical information about users and devices
// spending time in hackerspace in one safe transaction.
//
// Use this interface if you want to omit data races.
type StatusTx interface {
	// DevicesStatus accepts function that manipulates number of
	// unknown devices and online users in single safe transaction.
	DevicesStatus(context.Context, func(context.Context, Status) error) error
}
