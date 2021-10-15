package storage

import (
	"context"
	"fmt"
	"net"

	"golang.org/x/crypto/bcrypt"
)

// UpdateStatusesArgs contains arguments for UpdateStatuses function.
type UpdateStatusesArgs struct {
	Addresses          []net.HardwareAddr
	DevicesStorage     Devices
	OnlineUsersStorage OnlineUsers
	Counters           StatusTx
}

// UpdateStatuses set online user fields, with any device's MAC equal to one
// of addresses from given slice, to true and writes them to database.
func UpdateStatuses(ctx context.Context, args UpdateStatusesArgs) error {

	known, unknown := 0, 0
	onlineIDs := []string{}

	devices, err := args.DevicesStorage.All(ctx)
	if err != nil {
		return fmt.Errorf("args.DevicesStorage.All: %w", err)
	}

	for _, address := range args.Addresses {
		for _, device := range devices {
			if err := bcrypt.CompareHashAndPassword(device.MAC, address); err == nil {
				known += 1
				onlineIDs = append(onlineIDs, device.OwnerID)
			}
		}
	}

	if err := args.OnlineUsersStorage.Update(ctx, onlineIDs); err != nil {
		return fmt.Errorf("args.OnlineUsersStorage.Update: %w", err)
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
