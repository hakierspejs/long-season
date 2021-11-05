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

func (tf *TwoFactor) Get(ctx context.Context, userID string) (*models.TwoFactor, error) {
	return tf.cs.getTwoFactor(ctx, userID)
}

func (tf *TwoFactor) Update(ctx context.Context, userID string, f func(*models.TwoFactor) error) error {
	return tf.cs.updateTwoFactor(ctx, userID, f)
}

func (tf *TwoFactor) Remove(ctx context.Context, userID string) error {
	return tf.cs.removeTwoFactor(ctx, userID)
}
