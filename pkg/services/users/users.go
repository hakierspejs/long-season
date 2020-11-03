// Package users implements operations for
// user data manipulation.
package users

import (
	"bytes"
	"sort"

	"github.com/hakierspejs/long-season/pkg/models"
	"github.com/hakierspejs/long-season/pkg/services/update"
)

// Changes represents possible changes that
// can be applied to models.User
type Changes struct {
	Nickname string
	Password []byte
	Online   *bool
}

// Equals returns true if both users
// have same nickname, password and mac address.
func Equals(a, b models.User) bool {
	return all(
		a.Nickname == b.Nickname,
		bytes.Equal(a.Password, b.Password),
	)
}

// StrictEquals returns true if both users
// have same values assigned to every field.
func StrictEquals(a, b models.User) bool {
	return all(
		a.Nickname == b.Nickname,
		bytes.Equal(a.Password, b.Password),
		a.Online == b.Online,
		a.ID == b.ID,
	)
}

// Update applies given changes to given user model
// and returns new user model.
func Update(old models.User, c *Changes) models.User {
	return models.User{
		UserPublicData: models.UserPublicData{
			ID:       old.ID,
			Nickname: update.String(old.Nickname, c.Nickname),
			Online:   update.NullableBool(old.Online, c.Online),
		},
		Password: update.Bytes(old.Password, c.Password),
	}
}

// PublicSlice returns new slice with only public user data,
// created from given slice containing full user data.
func PublicSlice(u []models.User) []models.UserPublicData {
	public := make([]models.UserPublicData, len(u), cap(u))

	for i, old := range u {
		public[i] = old.UserPublicData
	}

	sort.Slice(public, func(i, j int) bool {
		return public[i].ID < public[j].ID
	})

	return public
}

// all returns true if all args are true,
// otherwise returns false.
func all(args ...bool) bool {
	for _, v := range args {
		if !v {
			return false
		}
	}
	return true
}
