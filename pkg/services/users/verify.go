package users

import (
	"errors"
	"regexp"
)

const (
	invalidNicknameMsg = "username should contains from 4 to 32 numerical and alphabetical characters"
	invalidPasswordMsg = "password should contains from 6 to 50 any characters, excluding whitespace characters"
)

var (
	// ErrInvalidNickname error is used for various verifies to
	// signal that user verification failed because of
	// invalid username. Raw message of error is safe to output
	// to client.
	ErrInvalidNickname = errors.New(invalidNicknameMsg)

	// ErrInvaliPassword error is used for various verifies to
	// signal that user verification failed because of
	// invalid password. Raw message of error is safe to output
	// to client.
	ErrInvaliPassword = errors.New(invalidPasswordMsg)
)

var (
	nicknameRegex = regexp.MustCompile(`^[a-zA-Z0-9]{4,32}$`)
	passwordRegex = regexp.MustCompile(`^[^[:space:]]{6,50}$`)
)

// VerifyNickname verifies if given nickname string
// is proper nickname for long-season application.
func VerifyNickname(n string) bool {
	return nicknameRegex.MatchString(n)
}

// VerifyPassword verifies if given password string
// is proper password for long-season application.
func VerifyPassword(p string) bool {
	return passwordRegex.MatchString(p)
}
