package errors

import (
	"errors"
)

var (
	// ErrDeviceDuplication is returned, when there is already device with given owner and tag.
	ErrDeviceDuplication = errors.New("there is already device with given owner and tag")

	// ErrNoID is returned when there is no resource with given id
	// stored in database.
	ErrNoID = errors.New("resource with given id not found")

	// ErrNicknameTaken is being returned when there is
	// already a user with given username.
	ErrNicknameTaken = errors.New("user with given username is already registered")
)
