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

func (d *Devices) New(ctx context.Context, userID string, m models.Device) (string, error) {
	return d.cs.newDevice(ctx, userID, m)
}

func (d *Devices) OfUser(ctx context.Context, userID string) ([]models.Device, error) {
	return d.cs.deviceOfUser(ctx, userID)
}

func (d *Devices) Read(ctx context.Context, id string) (*models.Device, error) {
	return d.cs.readDevice(ctx, id)
}

func (d *Devices) All(ctx context.Context) ([]models.Device, error) {
	return d.cs.allDevices(ctx)
}

func (d *Devices) Remove(ctx context.Context, id string) error {
	return d.cs.removeDevice(ctx, id)
}
