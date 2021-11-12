package temp

import (
	"context"
	"fmt"
	"sync"

	"github.com/hakierspejs/long-season/pkg/storage"
)

type status struct {
	onlineUsers    int
	unknownDevices int
}

// OnlineUsers returns number of people being
// currently online.
func (s status) OnlineUsers(ctx context.Context) (int, error) {
	return s.onlineUsers, nil
}

// SetOnlineUsers ovewrites number of people being
// currently online.
func (s *status) SetOnlineUsers(ctx context.Context, number int) error {
	s.onlineUsers = number
	return nil
}

// UnknownDevices returns number of unknown devices
// connected to the network.
func (s status) UnknownDevices(ctx context.Context) (int, error) {
	return s.unknownDevices, nil
}

// SetUnknownDevices overwrites number of unknown devices
// connected to the network.
func (s *status) SetUnknownDevices(ctx context.Context, number int) error {
	s.unknownDevices = number
	return nil
}

// StatusTx implements storage.StatusTx interface for temporary in
// memory storage.
type StatusTx struct {
	status *status
	guard  *sync.Mutex
}

// NewStatusTx is the only one safe constructor for StatusTx.
func NewStatusTx() *StatusTx {
	return &StatusTx{
		status: new(status),
		guard:  new(sync.Mutex),
	}
}

// DevicesStatus accepts function that manipulates number of
// unknown devices and online users in single safe transaction.
func (s *StatusTx) DevicesStatus(ctx context.Context, f func(context.Context, storage.Status) error) error {
	s.guard.Lock()
	defer s.guard.Unlock()

	if err := f(ctx, s.status); err != nil {
		return fmt.Errorf("f(ctx, s.status) update func error: %w", err)
	}
	return nil
}
