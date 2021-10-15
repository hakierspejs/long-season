package storage

import (
	"context"
	"fmt"

	"github.com/hakierspejs/long-season/pkg/models"
)

// UserAdapter implements methods for converting user entry
// data to the format compatible with the rest of the API.
type UserAdapter struct {
	OnlineUsersStorage OnlineUsers
}

// User adapts database user's entry to User model.
func (ua *UserAdapter) User(ctx context.Context, u UserEntry) (*models.User, error) {
	online, err := ua.OnlineUsersStorage.IsOnline(ctx, u.ID)
	if err != nil {
		return nil, fmt.Errorf("ua.OnlineUsersStorage.IsOnline: %w", err)
	}

	return &models.User{
		UserPublicData: models.UserPublicData{
			ID:       u.ID,
			Nickname: u.Nickname,
			Online:   online,
		},
		Password: u.HashedPassword,
		Private:  u.Private,
	}, nil
}

// Users transform slice of user entries to slice of user models.
func (ua *UserAdapter) Users(ctx context.Context, users []UserEntry) ([]models.User, error) {
	res := make([]models.User, len(users), len(users))

	for i, u := range users {
		newUser, err := ua.User(ctx, u)
		if err != nil {
			return nil, fmt.Errorf("ua.User: %w", err)
		}

		res[i] = *newUser
	}

	return res, nil
}
