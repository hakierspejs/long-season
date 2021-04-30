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

// FilterFunc accepts User model and returns
// true or false.
type FilterFunc func(models.User) bool

// Not returns opposite of given FilterFunc
// result.
func Not(f FilterFunc) FilterFunc {
	return func(u models.User) bool {
		return !f(u)
	}
}

// Online returns true if given user is
// currently online.
func Online(u models.User) bool {
	return u.Online
}

// Private returns true if given user has
// private flag set to true.
func Private(u models.User) bool {
	return u.Private
}

// Filter returns slice of users that passed all given
// FilterFunc tests. If no filters given, returns exactly same
// slice of users that has been passed to function.
func Filter(users []models.User, filters ...FilterFunc) []models.User {
	if len(filters) == 0 {
		return users
	}

	filtered := make([]models.User, 0, len(users))
	tests := make([]bool, len(filters), len(filters))
	for _, u := range users {
		for i := 0; i < len(filters); i++ {
			tests[i] = filters[i](u)
		}

		if all(tests...) {
			filtered = append(filtered, u)
		}
	}

	return filtered
}

// DefaultFilters returns slice with convenient collection of
// default filters that can be used for outputting user data.
//
// For example: default filters contains Not(Online) filter
// for hiding private users.
func DefaultFilters() []FilterFunc {
	return []FilterFunc{
		Not(Online),
	}
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
