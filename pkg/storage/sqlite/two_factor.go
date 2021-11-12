package sqlite

import (
	"context"

	"github.com/hakierspejs/long-season/pkg/models"
)

// TwoFactor storage implements storage.TwoFactor interface for
// sqlite database.
type TwoFactor struct {
	cs *coreStorage
}

// Get returns two factor methods for user with given user ID.
func (tf *TwoFactor) Get(ctx context.Context, userID string) (*models.TwoFactor, error) {
	return tf.cs.getTwoFactor(ctx, userID)
}

// Updates apply given function to two factor methods of user
// with given user ID.
func (tf *TwoFactor) Update(ctx context.Context, userID string, f func(*models.TwoFactor) error) error {
	return tf.cs.updateTwoFactor(ctx, userID, f)
}

// Remove deletes all two factor methods of user with given
// user ID. You can still use Update method after all to start
// adding more methods.
func (tf *TwoFactor) Remove(ctx context.Context, userID string) error {
	return tf.cs.removeTwoFactor(ctx, userID)
}
