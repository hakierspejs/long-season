// Package exim exports and imports data from storage.
//
// Exim is shortcut for "ex"port and "im"port.
package exim

import "github.com/hakierspejs/long-season/pkg/models/set"

// Data holds whole dump from storage.
type Data struct {
	Users map[string]User `json:"users"`
}

// User holds information about user and
// its devices.
type User struct {
	ID        string     `json:"id"`
	Nickname  string     `json:"nickname"`
	Password  []byte     `json:"password"`
	Devices   []Device   `json:"devices"`
	TwoFactor *TwoFactor `json:"twoFactor,omitempty"`
}

// Device represents single users device.
type Device struct {
	ID  string `json:"id"`
	Tag string `json:"tag"`
	MAC []byte `json:"mac"`
}

// TwoFactor holds storage data for two factor methods.
type TwoFactor struct {
	OneTimeCodes  []OneTimeCode `json:"oneTimeCodes,omitempty"`
	RecoveryCodes []Recovery    `json:"recoveryCodes,omitempty"`
}

// OneTimeCode holds storage data for one time passwords.
type OneTimeCode struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Secret string `json:"secret"`
}

// Recovery holds storage data for recovery codes.
type Recovery struct {
	// ID is unique id of one time code.
	ID string `json:"id"`

	// Name is human readable name of one time code
	// provided by user.
	Name string `json:"name"`

	// Codes holds set with recovery codes.
	Codes *set.String `json:"codes,omitempty"`
}
