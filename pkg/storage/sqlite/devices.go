package sqlite

import (
	"context"

	"github.com/hakierspejs/long-season/pkg/models"
)

// Devices storage implements storage.Devices interface for
// sqlite database.
type Devices struct {
	cs *coreStorage
}

// New stores given devices data in database and returns
// id of new device (given with input model).
func (d *Devices) New(ctx context.Context, userID string, m models.Device) (string, error) {
	return d.cs.newDevice(ctx, userID, m)
}

// OfUser returns list of models owned by user with given id.
func (d *Devices) OfUser(ctx context.Context, userID string) ([]models.Device, error) {
	return d.cs.deviceOfUser(ctx, userID)
}

// Read returns single device data with given ID.
func (d *Devices) Read(ctx context.Context, id string) (*models.Device, error) {
	return d.cs.readDevice(ctx, id)
}

// All returns slice with all devices from storage.
func (d *Devices) All(ctx context.Context) ([]models.Device, error) {
	return d.cs.allDevices(ctx)
}

// Remove deletes device with given id from storage.
func (d *Devices) Remove(ctx context.Context, id string) error {
	return d.cs.removeDevice(ctx, id)
}
