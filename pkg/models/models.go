package models

import (
	"encoding/gob"
	"fmt"
)

func init() {
	gob.Register(User{})
}

// User represents single user data stored in storage.
type User struct {
	UserPublicData

	// MAC contains the MAC address of the user's device.
	MAC []byte
	// Password of User hashed with bcrypt algorithm.
	Password []byte
}

// UserPublicData is subset of User containing
// only data that can be shown publicly to
// everybody that will interact with API or website.
type UserPublicData struct {
	// ID unique to every user.
	ID int `json:"id"`
	// Nickname represents name that will be exposed to public,
	// to inform people who is in the hackerspace.
	Nickname string `json:"nickname"`
	// Online indicates if player is currently in the hackerspace.
	Online bool `json:"online"`
}

// Config represents configuration that is
// being used by server.
type Config struct {
	Host         string
	Port         string
	DatabasePath string
	JWTSecret    string
}

// Address returns address string that is compatible
// with http.ListenAndServe function.
func (c Config) Address() string {
	return fmt.Sprintf("%s:%s", c.Host, c.Port)
}
