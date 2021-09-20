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
	New(ctx context.Context, u models.User) (string, error)
	Read(ctx context.Context, id string) (*models.User, error)
	All(ctx context.Context) ([]models.User, error)
	Remove(ctx context.Context, id string) error
	Update(ctx context.Context, id string, f func(*models.User) error) error
}

type Devices interface {
	New(ctx context.Context, userID string, d models.Device) (string, error)
	OfUser(ctx context.Context, userID string) ([]models.Device, error)
	Read(ctx context.Context, id string) (*models.Device, error)
	All(ctx context.Context) ([]models.Device, error)
	Remove(ctx context.Context, id string) error
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

// TwoFactor implements methods for managing given user two factor
// methods. By default every user has two factor entry with empty
// values at the moment of account creation, so there is no need
// for "New" method. If you want to add new two factor method for
// given user you can immediately start with using "Update" method.
type TwoFactor interface {
	// Get returns two factor methods for user with given user ID.
	Get(ctx context.Context, userID string) (*models.TwoFactor, error)

	// Updates apply given function to two factor methods of user
	// with given user ID.
	Update(ctx context.Context, userID string, f func(*models.TwoFactor) error) error

	// Remove deletes all two factor methods of user with given
	// user ID. You can still use Update method after all to start
	// adding more methods.
	Remove(ctx context.Context, userID string) error
}
