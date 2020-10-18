package errors

import "fmt"

// UsernameTaken is being returned when there is
// already a user with given username.
type NicknameTaken string

// Error method, which implements error interface.
func (u NicknameTaken) Error() string {
	return fmt.Sprintf(
		"long-season storage: user with \"%s\" username is already registered.",
		string(u),
	)
}

// NoID is returned when there is no user with given id
// stored in database.
type NoID int

// Error method, which implements error interface.
func (id NoID) Error() string {
	return fmt.Sprintf(
		"long-season storage: there is no user with \"%d\" id.", int(id),
	)
}

// ErrUsernameTaken is handy facade for UsernameTaken error.
func ErrNicknameTaken(username string) error {
	return NicknameTaken(username)
}

// ErrNoID is handy facade for NoID error.
func ErrNoID(id int) error {
	return NoID(id)
}
