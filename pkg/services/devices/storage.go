package devices

import (
	"context"
	"errors"
	"fmt"
	"net"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/hakierspejs/long-season/pkg/models"
	"github.com/hakierspejs/long-season/pkg/services/happier"
	"github.com/hakierspejs/long-season/pkg/storage"
	serrors "github.com/hakierspejs/long-season/pkg/storage/errors"
)

const internalServerErrorResponse = "Internal server error. Please try again later."

// AddDeviceRequest holds arguments for Add service method.
type AddDeviceRequest struct {
	// OwnerID is id of user that owns new device.
	OwnerID string

	// Owner is nickname of new device owner.
	Owner string

	// Tag is a name for new device.
	Tag string

	// MAC is raw hardress address of the device.
	//
	// There is no need to verify MAC before passing to
	// AddDeviceRequest, because it will be verified during
	// adding to storage by Add function.
	MAC string

	// Storage for devices.
	Storage storage.Devices
}

// Add adds new Device to given storage with default options. Returns
// assigned ID if there is no error.
func Add(ctx context.Context, args AddDeviceRequest) (string, error) {
	errFactory := happier.FromContext(ctx)

	newID := uuid.New().String()

	mac, err := net.ParseMAC(args.MAC)
	if err != nil {
		return "", errFactory.BadRequest(
			fmt.Errorf("net.ParseMAC: %w", err),
			fmt.Sprintf("invalid input: invalid mac address %s", mac),
		)
	}

	hashedMac, err := bcrypt.GenerateFromPassword(mac, bcrypt.DefaultCost)
	if err != nil {
		return "", errFactory.InternalServerError(
			fmt.Errorf("bcrypt.GenerateFromPassword: %w", err),
			internalServerErrorResponse,
		)
	}

	_, err = args.Storage.New(ctx, args.OwnerID, models.Device{
		DevicePublicData: models.DevicePublicData{
			ID:    newID,
			Tag:   args.Tag,
			Owner: args.Owner,
		},
		OwnerID: args.OwnerID,
		MAC:     hashedMac,
	})
	if errors.Is(err, serrors.ErrDeviceDuplication) {
		return "", errFactory.Conflict(
			fmt.Errorf("db.New: %w", err),
			fmt.Sprintf("tag already used"),
		)
	}
	if err != nil {
		return "", errFactory.InternalServerError(
			fmt.Errorf("db.New: %w", err),
			internalServerErrorResponse,
		)
	}

	return newID, nil
}
