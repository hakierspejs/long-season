package exim

import (
	"context"
	"fmt"

	"github.com/hakierspejs/long-season/pkg/storage"
)

// ExportRequest holds dependencies for Export procedure.
type ExportRequest struct {
	UsersStorage     storage.Users
	DevicesStorage   storage.Devices
	TwoFactorStorage storage.TwoFactor
}

// Export all data from storage to the single data structure.
func Export(ctx context.Context, req ExportRequest) (*Data, error) {
	res := &Data{
		Users: map[string]User{},
	}

	users, err := req.UsersStorage.All(ctx)
	if err != nil {
		return nil, fmt.Errorf("req.UsersStorage.All: %w", err)
	}

	for _, u := range users {
		res.Users[u.ID] = User{
			ID:       u.ID,
			Nickname: u.Nickname,
			Password: u.HashedPassword,
			Devices:  []Device{},
			TwoFactor: &TwoFactor{
				OneTimeCodes:  []OneTimeCode{},
				RecoveryCodes: []Recovery{},
			},
		}
	}

	devices, err := req.DevicesStorage.All(ctx)
	if err != nil {
		return nil, fmt.Errorf("req.DevicesStorage.All: %w", err)
	}

	for _, d := range devices {
		currUser, ok := res.Users[d.OwnerID]
		if !ok {
			// There is no user with such user ID.
			// It is possible that user data was deleted
			// but its devices data wasn't garbage-collected.
			//
			// This scenario is not positive at all, but
			// we can pass it by, because it won't affect
			// database dump and garbage data won't be exported,
			// which is welcome.
			continue
		}

		currUser.Devices = append(currUser.Devices, Device{
			ID:  d.ID,
			Tag: d.Tag,
			MAC: d.MAC,
		})
		res.Users[d.OwnerID] = currUser
	}

	// It associates user id with one time codes of particular user.
	allOneTimeCodes := map[string][]OneTimeCode{}

	// It associates user id with recovery codes of particular user.
	allRecoveryCodes := map[string][]Recovery{}

	for _, u := range res.Users {
		methods, err := req.TwoFactorStorage.Get(ctx, u.ID)
		if err != nil {
			return nil, fmt.Errorf("req.TwoFactorStorage.Get(\"%s\"): %w", u.ID, err)
		}

		if methods == nil {
			continue
		}

		for _, otp := range methods.OneTimeCodes {
			currOTPs := allOneTimeCodes[u.ID]
			currOTPs = append(currOTPs, OneTimeCode{
				ID:     otp.ID,
				Name:   otp.Name,
				Secret: otp.Secret,
			})
			allOneTimeCodes[u.ID] = currOTPs
		}

		for _, recovery := range methods.RecoveryCodes {
			currRecoveryCodes := allRecoveryCodes[u.ID]
			currRecoveryCodes = append(currRecoveryCodes, Recovery{
				ID:    recovery.ID,
				Name:  recovery.Name,
				Codes: recovery.Codes,
			})
			allRecoveryCodes[u.ID] = currRecoveryCodes
		}
	}

	for userID, otps := range allOneTimeCodes {
		// Retrieve user with associated user id from
		// map with all one time codes.
		currUser, ok := res.Users[userID]
		if !ok {
			// If there is no user with given user id,
			// skip it.
			continue
		}

		// Overwrite empty slice with one time codes from
		// database dump associated with current user id.
		currUser.TwoFactor.OneTimeCodes = otps

		// Overwrite updated user data.
		res.Users[userID] = currUser
	}

	for userID, recoveryCodes := range allRecoveryCodes {
		// Retrieve user with associated user id from
		// map with all recovery codes.
		currUser, ok := res.Users[userID]
		if !ok {
			// If there is no user with given user id,
			// skip it.
			continue
		}

		// Overwrite empty slice with one time codes from
		// database dump associated with current user id.
		currUser.TwoFactor.RecoveryCodes = recoveryCodes

		// Overwrite updated user data.
		res.Users[userID] = currUser
	}

	return res, nil
}
