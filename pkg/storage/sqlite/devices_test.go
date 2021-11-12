package sqlite

import (
	"context"
	"testing"

	"github.com/matryer/is"

	"github.com/hakierspejs/long-season/pkg/models"
	"github.com/hakierspejs/long-season/pkg/storage"
)

func TestDevices(t *testing.T) {
	is := is.New(t)
	ctx := context.Background()

	f, closer, err := NewFactory(":memory:")
	is.NoErr(err)
	defer closer()

	devicesData := map[string]models.Device{
		"1": {
			DevicePublicData: models.DevicePublicData{
				ID:    "1",
				Tag:   "one",
				Owner: "johnny",
			},
			OwnerID: "1",
			MAC:     []byte("11:11:11:11:11:11"),
		},
		"2": {
			DevicePublicData: models.DevicePublicData{
				ID:    "2",
				Tag:   "two",
				Owner: "johnny",
			},
			OwnerID: "1",
			MAC:     []byte("22:22:22:22:22:22"),
		},
		"3": {
			DevicePublicData: models.DevicePublicData{
				ID:    "3",
				Tag:   "three",
				Owner: "marco",
			},
			OwnerID: "2",
			MAC:     []byte("33:33:33:33:33:33"),
		},
	}

	usersData := map[string]storage.UserEntry{
		"1": {
			ID:             "1",
			Nickname:       "johnny",
			HashedPassword: []byte("71Hk4Rt2WY8xqgYoKxPm"),
			Private:        false,
		},
		"2": {
			ID:             "2",
			Nickname:       "marco",
			HashedPassword: []byte("u8dXHRi0JNo23JVeHkjh"),
			Private:        true,
		},
	}

	su := f.Users()
	for _, u := range usersData {
		id, err := su.New(ctx, u)
		is.NoErr(err)
		is.Equal(id, u.ID)
	}

	sd := f.Devices()
	for _, d := range devicesData {
		id, err := sd.New(ctx, d.OwnerID, d)
		is.NoErr(err)
		is.Equal(id, d.ID)
	}

	allDevices, err := sd.All(ctx)
	is.NoErr(err)

	for _, d := range allDevices {
		currentDevice, ok := devicesData[d.ID]
		is.True(ok)
		is.Equal(currentDevice.ID, d.ID)
		is.Equal(currentDevice.MAC, d.MAC)
		is.Equal(currentDevice.OwnerID, d.OwnerID)
		is.Equal(currentDevice.Owner, d.Owner)
		is.Equal(currentDevice.Tag, d.Tag)
	}

	johnnyDevices, err := sd.OfUser(ctx, "1")
	is.NoErr(err)

	for _, d := range johnnyDevices {
		currentDevice, ok := devicesData[d.ID]
		is.True(ok)
		is.Equal(currentDevice.ID, d.ID)
		is.Equal(currentDevice.MAC, d.MAC)
		is.Equal(currentDevice.OwnerID, "1")
		is.Equal(currentDevice.Owner, "johnny")
		is.Equal(currentDevice.Tag, d.Tag)
	}

	marcoDevices, err := sd.OfUser(ctx, "2")
	for _, d := range marcoDevices {
		currentDevice, ok := devicesData[d.ID]
		is.True(ok)
		is.Equal(currentDevice.ID, d.ID)
		is.Equal(currentDevice.MAC, d.MAC)
		is.Equal(currentDevice.OwnerID, "2")
		is.Equal(currentDevice.Owner, "marco")
		is.Equal(currentDevice.Tag, d.Tag)
	}

	for _, d := range devicesData {
		current, err := sd.Read(ctx, d.ID)
		is.NoErr(err)
		is.Equal(current.ID, d.ID)
		is.Equal(current.MAC, d.MAC)
		is.Equal(current.OwnerID, d.OwnerID)
		is.Equal(current.Owner, d.Owner)
		is.Equal(current.Tag, d.Tag)
	}

	err = sd.Remove(ctx, "2")
	is.NoErr(err)

	d, err := sd.Read(ctx, "2")
	is.True(d == nil)
	is.True(err != nil)

	// Try to add invalid device.
	_, err = sd.New(ctx, "5", models.Device{})
	is.True(err != nil)
}
