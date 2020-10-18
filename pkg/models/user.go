package models

// User represents single user data stored in storage.
type User struct {
	// ID unique to every user.
	ID int
	// Nickname represents name that will be exposed to public,
	// to inform people who is in the hackerspace.
	Nickname string
	// Online indicates if player is currently in the hackerspace.
	Online bool
	// MAC address of User hashed with bcrypt algorithm.
	MAC []byte
	// Password of User hashed with bcrypt algorithm.
	Password []byte
}
