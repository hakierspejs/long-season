package storage

import (
	"context"
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

// UpdateStatuses set online user fields, with any device's MAC equal to one
// of addresses from given slice, to true and writes them to database.
func UpdateStatuses(
	ctx context.Context, addresses []net.HardwareAddr, iter StatusIterator,
) error {

	return iter.ForEachUpdate(ctx,
		func(u models.User, devices []models.Device) (*models.User, error) {
			result := u
			result.Online = false

			for _, address := range addresses {
				for _, device := range devices {
					if err := bcrypt.CompareHashAndPassword(device.MAC, address); err == nil {
						result.Online = true
						return &result, nil
					}
				}
			}

			return &result, nil
		})
}
