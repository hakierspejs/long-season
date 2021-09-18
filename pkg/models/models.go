package models

import (
	"encoding/gob"
	"fmt"
	"time"

	"github.com/cristalhq/jwt/v3"
)

func init() {
	gob.Register(User{})
	gob.Register(Device{})
}

// User represents single user data stored in storage.
type User struct {
	UserPublicData

	// Password of User hashed with bcrypt algorithm.
	Password []byte

	// Private is flag for enabling private-mode that hides
	// user activity from others.
	Private bool

	// TwoFactor is not nil if the user has any
	// two factor method activated.
	TwoFactor *TwoFactor
}

// UserPublicData is subset of User containing
// only data that can be shown publicly to
// everybody that will interact with API or website.
type UserPublicData struct {
	// ID unique to every user.
	ID string `json:"id"`
	// Nickname represents name that will be exposed to public,
	// to inform people who is in the hackerspace.
	Nickname string `json:"nickname"`
	// Online indicates if player is currently in the hackerspace.
	Online bool `json:"online"`
}

// TwoFactor holds two factor methods with
// data required to verify with one of the
// following methods.
type TwoFactor struct {
	OneTimeCodes []OneTimeCode
}

// OneTimeCode holds data stored in database for two factor
// verification with one time codes.
type OneTimeCode struct {
	// ID is unique id of one time code.
	ID string

	// Name is human readable name of one time code
	// provided by user.
	Name string

	// Secret is used to verify one time code.
	Secret string
}

type TwoFactorMethods struct {
}

type Device struct {
	DevicePublicData

	// OwnerID is id of user that owns this device.
	OwnerID string
	// MAC contains hashed MAC address of the device.
	MAC []byte
}

type DevicePublicData struct {
	ID    string `json:"id"`
	Tag   string `json:"tag"`
	Owner string `json:"owner"`
}

// Config represents configuration that is
// being used by server.
type Config struct {
	// Space is the name of the space
	// where long-season is watching for macs.
	Space string

	// City is name of city, where
	// long-season is watching for macs.
	City string

	Debug         bool
	Host          string
	Port          string
	DatabasePath  string
	JWTSecret     string
	UpdateSecret  string
	AppName       string
	RefreshTime   time.Duration
	SingleAddrTTL time.Duration
}

// Address returns address string that is compatible
// with http.ListenAndServe function.
func (c Config) Address() string {
	return fmt.Sprintf("%s:%s", c.Host, c.Port)
}

// Claims represents custom claims for jwt authentication.
type Claims struct {
	jwt.StandardClaims
	UserID   string                 `json:"id"`
	Nickname string                 `json:"nck"`
	Values   map[string]interface{} `json:"vls,omitempty"`
}
