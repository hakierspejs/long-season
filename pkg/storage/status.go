package storage

import (
	"context"
	"net"

	"golang.org/x/crypto/bcrypt"
)

// UpdateStatuses set online user fields, with any device's MAC equal to one
// of addresses from given slice, to true and writes them to database.
func UpdateStatuses(ctx context.Context, addresses []string, d Devices, u Users) error {
	users, err := u.All(ctx)
	if err != nil {
		return err
	}

	devices, err := d.All(ctx)
	if err != nil {
		return err
	}

	parsedAddresses := []net.HardwareAddr{}
	var parsedAddress net.HardwareAddr
	for _, address := range addresses {
		parsedAddress, err = net.ParseMAC(address)
		if err != nil {
			return err
		}
		parsedAddresses = append(parsedAddresses, parsedAddress)
	}

	online := map[int]struct{}{}
	for _, address := range parsedAddresses {
		for _, device := range devices {
			if err := bcrypt.CompareHashAndPassword(device.MAC, address); err == nil {
				online[device.OwnerID] = struct{}{}
			}
		}
	}

	for i, user := range users {
		_, isOnline := online[user.ID]
		users[i].Online = isOnline
	}

	return u.UpdateMany(ctx, users)
}
