package ctxkey

import (
	"context"
	"errors"
)

var ErrValueNotFound = errors.New("ctxkey: value not found in ctx")

// Debug returns value of debug mode stored in context store.
func Debug(ctx context.Context) (bool, error) {
	mode, ok := ctx.Value(DebugKey).(bool)
	if !ok {
		return false, ErrValueNotFound
	}
	return mode, nil
}
