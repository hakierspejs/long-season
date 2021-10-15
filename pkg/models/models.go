package models

import (
	"encoding/gob"
	"fmt"
	"time"

	"github.com/cristalhq/jwt/v3"
	"github.com/hakierspejs/long-season/pkg/models/set"
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

// TwoFactorType describes type of two factor for
// inspecting user's two factor enabled methods.
type TwoFactorType string

const (
	// OneTimeCodes are time-based one time passwords.
	OneTimeCodes TwoFactorType = "totp"

	// RecoveryCodes are codes for one time use as additional
	// two factor method.
	RecoveryCodes TwoFactorType = "recovery codes"
)

// TwoFactorMethod holds private data of enabled
// two factor methods for given user. It allows
// users to discover their two factor methods.
type TwoFactorMethod struct {
	// ID is unique identifier of given two factor method.
	ID string `json:"id"`

	// Name is human readable name given by user
	// to its two factor method.
	Name string `json:"name"`

	// Type is two factor type.
	Type TwoFactorType `json:"type"`

	// Locations is URI where given two factor method
	// is stored. It can be used to disable (delete)
	// given two factor method.
	Location string `json:"location"`
}

// TwoFactor holds two factor methods with
// data required to verify with one of the
// following methods.
type TwoFactor struct {
	// OneTimeCodes is map of one time codes entry with
	// theirs IDs as maps keys.
	OneTimeCodes map[string]OneTimeCode `json:"oneTimeCodes,omitempty"`

	// RecoveryCodes is map of recovery codes entries with
	// theirs IDs as maps keys.
	RecoveryCodes map[string]Recovery `json:"recoveryCodes,omitempty"`
}

// OneTimeCode holds data stored in database for two factor
// verification with one time codes.
type OneTimeCode struct {
	// ID is unique id of one time code.
	ID string `json:"id,omitempty"`

	// Name is human readable name of one time code
	// provided by user.
	Name string `json:"name,omitempty"`

	// Secret is used to verify one time code.
	Secret string `json:"secret,omitempty"`
}

// Method is adapter of OneTimeCode for TwoFactorMethod type.
func (o OneTimeCode) Method(userID string) TwoFactorMethod {
	return TwoFactorMethod{
		ID:       o.ID,
		Name:     o.Name,
		Type:     OneTimeCodes,
		Location: fmt.Sprintf("/api/v1/users/%s/twofactor/%s", userID, o.ID),
	}
}

// Recovery holds data stored in database for two
// factor verification with recovery codes.
type Recovery struct {
	// ID is unique id of one time code.
	ID string `json:"id,omitempty"`

	// Name is human readable name of one time code
	// provided by user.
	Name string `json:"name,omitempty"`

	// Codes holds set with recovery codes.
	Codes *set.String `json:"codes,omitempty"`
}

// Method is adapter of RecoveryCodes for TwoFactorMethod type.
func (r Recovery) Method(userID string) TwoFactorMethod {
	return TwoFactorMethod{
		ID:       r.ID,
		Name:     r.Name,
		Type:     RecoveryCodes,
		Location: fmt.Sprintf("/api/v1/users/%s/twofactor/%s", userID, r.ID),
	}
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
