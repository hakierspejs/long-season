// Package temp implements temporary storage interfaces
// that will be valid only during runtime and will evaporate
// after shutdown of program.
package temp

import (
	"context"
	"errors"
	"sync"
)

var ErrContextDone = errors.New("temp: context is done, performed graceful shutdown")

type OnlineUsers struct {
	set   map[string]struct{}
	guard *sync.RWMutex
}

func NewOnlineUsers() *OnlineUsers {
	return &OnlineUsers{
		set:   map[string]struct{}{},
		guard: new(sync.RWMutex),
	}
}

func (o *OnlineUsers) All(ctx context.Context) ([]string, error) {
	o.guard.RLock()
	defer o.guard.RUnlock()

	resChan := make(chan []string)
	go func() {
		res := []string{}
		for id, _ := range o.set {
			res = append(res, id)
		}
		resChan <- res
	}()

	select {
	case res := <-resChan:
		return res, nil
	case <-ctx.Done():
		return nil, ErrContextDone
	}
}

func (o *OnlineUsers) Update(ctx context.Context, ids []string) error {
	o.guard.Lock()
	defer o.guard.Unlock()

	success := make(chan struct{})
	go func() {
		o.set = map[string]struct{}{}
		for _, id := range ids {
			o.set[id] = struct{}{}
		}
		success <- struct{}{}
	}()

	select {
	case <-success:
		return nil
	case <-ctx.Done():
		return ErrContextDone
	}
}

func (o *OnlineUsers) IsOnline(ctx context.Context, id string) (bool, error) {
	o.guard.RLock()
	defer o.guard.RUnlock()

	resChan := make(chan bool)
	go func() {
		_, res := o.set[id]
		resChan <- res
	}()

	select {
	case res := <-resChan:
		return res, nil
	case <-ctx.Done():
		return false, ErrContextDone
	}
}
