// Package users implements operations for
// user data manipulation.
package users

import (
	"github.com/hakierspejs/long-season/pkg/models"
)

// Changes represents possible changes that
// can be applied to models.User
type Changes struct {
	Nickname string
	MAC      []byte
	Password []byte
	Online   *bool
}

// Update applies given changes to given user model
// and returns new user model.
func Update(old models.User, c *Changes) models.User {
	return models.User{
		ID:       old.ID,
		Nickname: updateString(old.Nickname, c.Nickname),
		MAC:      updateByteSlice(old.MAC, c.MAC),
		Password: updateByteSlice(old.Password, c.Password),
		Online:   updateNullableBool(old.Online, c.Online),
	}
}

func updateByteSlice(old, changes []byte) []byte {
	if len(changes) > 0 {
		return changes
	}
	return old
}

func updateString(old, changes string) string {
	if len(changes) > 0 {
		return changes
	}
	return old
}

func updateNullableBool(old bool, change *bool) bool {
	if change != nil {
		return *change
	}
	return old
}
