package exim

import (
	"context"
	"fmt"

	"github.com/hakierspejs/long-season/pkg/models"
	"github.com/hakierspejs/long-season/pkg/storage"
)

// ImportRequest holds dependencies for Import procedure.
type ImportRequest struct {
	Dump             Data
	UsersStorage     storage.Users
	DevicesStorage   storage.Devices
	TwoFactorStorage storage.TwoFactor
}

// Import parsed database dump into database storage.
func Import(ctx context.Context, req ImportRequest) error {
	for _, user := range req.Dump.Users {
		_, err := req.UsersStorage.New(ctx, storage.UserEntry{
			ID:             user.ID,
			Nickname:       user.Nickname,
			HashedPassword: user.Password,
			Private:        false,
		})
		if err != nil {
			return fmt.Errorf("req.UsersStorage.New: %w", err)
		}

		for _, device := range user.Devices {
			_, err = req.DevicesStorage.New(ctx, user.ID, models.Device{
				DevicePublicData: models.DevicePublicData{
					ID:    device.ID,
					Tag:   device.Tag,
					Owner: user.Nickname,
				},
				OwnerID: user.ID,
				MAC:     device.MAC,
			})
			if err != nil {
				return fmt.Errorf("req.DevicesStorage.New: %w", err)
			}
		}

		if user.TwoFactor != nil {
			for _, oneTimeCode := range user.TwoFactor.OneTimeCodes {
				err := req.TwoFactorStorage.Update(ctx, user.ID, func(tf *models.TwoFactor) error {
					tf.OneTimeCodes[oneTimeCode.ID] = models.OneTimeCode{
						ID:     oneTimeCode.ID,
						Name:   oneTimeCode.Name,
						Secret: oneTimeCode.Secret,
					}
					return nil
				})
				if err != nil {
					return fmt.Errorf("req.TwoFactorStorage.Update(OneTimeCode): %w", err)
				}
			}

			for _, recovery := range user.TwoFactor.RecoveryCodes {
				err := req.TwoFactorStorage.Update(ctx, user.ID, func(tf *models.TwoFactor) error {
					tf.RecoveryCodes[recovery.ID] = models.Recovery{
						ID:    recovery.ID,
						Name:  recovery.Name,
						Codes: recovery.Codes,
					}
					return nil
				})
				if err != nil {
					return fmt.Errorf("req.TwoFactorStorage.Update(Recovery): %w", err)
				}
			}
		}
	}

	return nil
}
