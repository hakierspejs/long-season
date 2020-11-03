// Package devices implements operations for
// device data manipulation.
package devices

import (
	"sort"

	"github.com/hakierspejs/long-season/pkg/models"
	"github.com/hakierspejs/long-season/pkg/services/update"
)

// Changes represents possible changes that
// can be applied to models.Device
type Changes struct {
	MAC []byte
	Tag string
}

// Update applies given changes to given device model
// and returns new device model.
func Update(old models.Device, c *Changes) models.Device {
	return models.Device{
		DevicePublicData: models.DevicePublicData{
			ID:    old.ID,
			Owner: old.Owner,
			Tag:   update.String(old.Tag, c.Tag),
		},
		OwnerID: old.OwnerID,
		MAC:     update.Bytes(old.MAC, c.MAC),
	}
}

// PublicSlice returns new slice with only public device data,
// created from given slice containing full device data.
func PublicSlice(d []models.Device) []models.DevicePublicData {
	public := make([]models.DevicePublicData, len(d), cap(d))

	for i, old := range d {
		public[i] = old.DevicePublicData
	}

	sort.Slice(public, func(i, j int) bool {
		return public[i].ID < public[j].ID
	})

	return public
}

func Filter(d []models.Device, f func(models.Device) bool) []models.Device {
	res := []models.Device{}
	for _, device := range d {
		if f(device) {
			res = append(res, device)
		}
	}
	return res
}
