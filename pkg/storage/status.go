package storage

import (
	"context"
	"fmt"
	"net"

	"github.com/hakierspejs/long-season/pkg/models"
	"golang.org/x/crypto/bcrypt"
)

// StatusIterator modifies users data based on their devices.
type StatusIterator interface {
	// ForEachUpdate iterates over every user from database and their
	// devices. Overwrites current user data with returned one.
	ForEachUpdate(
		context.Context,
		func(models.User, []models.Device) (*models.User, error),
	) error
}

// UpdateStatusesArgs contains arguments for UpdateStatuses function.
type UpdateStatusesArgs struct {
	Addresses []net.HardwareAddr
	Iter      StatusIterator
	Counters  StatusTx
}

// UpdateStatuses set online user fields, with any device's MAC equal to one
// of addresses from given slice, to true and writes them to database.
func UpdateStatuses(ctx context.Context, args UpdateStatusesArgs) error {

	known, unknown := 0, 0
	err := args.Iter.ForEachUpdate(ctx,
		func(u models.User, devices []models.Device) (*models.User, error) {
			result := u
			result.Online = false

			for _, address := range args.Addresses {
				for _, device := range devices {
					if err := bcrypt.CompareHashAndPassword(device.MAC, address); err == nil {
						known += 1
						result.Online = true
						return &result, nil
					}
				}
			}

			return &result, nil
		})
	if err != nil {
		return fmt.Errorf("failed to update statuses: %w", err)
	}

	unknown = len(args.Addresses) - known

	return args.Counters.DevicesStatus(ctx,
		func(ctx context.Context, s Status) error {
			if err := s.SetOnlineUsers(ctx, known); err != nil {
				return fmt.Errorf("failed to set online users: %w", err)
			}

			if err := s.SetUnknownDevices(ctx, unknown); err != nil {
				return fmt.Errorf("failed to set unknown devices: %w", err)
			}

			return nil
		})
}
