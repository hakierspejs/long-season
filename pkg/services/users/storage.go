package users

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/hakierspejs/long-season/pkg/storage"
)

// AddUserRequest contains arguments and dependencies for
// adding new User entry to storage.
type AddUserRequest struct {
	// Nickname represents name that will be exposed to public,
	// to inform people who is in the hackerspace.
	Nickname string

	// Raw password.
	Password []byte

	// Storage for users.
	Storage storage.Users
}

// Add adds new User to given storage with default options. Returns
// new users ID if succeds.
func Add(ctx context.Context, args AddUserRequest) (string, error) {
	newID := uuid.New().String()

	pass, err := bcrypt.GenerateFromPassword(args.Password, bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("bcrypt.GenerateFromPassword: %w", err)
	}

	_, err = args.Storage.New(ctx, storage.UserEntry{
		ID:             newID,
		Nickname:       args.Nickname,
		HashedPassword: pass,
		Private:        false,
	})
	if err != nil {
		return "", fmt.Errorf("args.Storage.New: %w", err)
	}

	return newID, nil
}
