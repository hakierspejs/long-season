package macs

import (
	"context"
	"net"
	"time"
)

type SetTTL struct {
	m        map[string]*time.Timer
	toAdd    chan setItem
	toDel    chan string
	retrieve chan struct{}
	macArr   chan []net.HardwareAddr
}

type setItem struct {
	value string
	ttl   time.Duration
}

func NewSetTTL(ctx context.Context) *SetTTL {
	res := &SetTTL{
		m:        map[string]*time.Timer{},
		toAdd:    make(chan setItem),
		toDel:    make(chan string),
		retrieve: make(chan struct{}),
		macArr:   make(chan []net.HardwareAddr),
	}

	// start daemon in new goroutine
	go res.daemon(ctx)

	return res
}

func (s *SetTTL) Push(addr net.HardwareAddr, ttl time.Duration) {
	s.toAdd <- setItem{
		value: string(addr),
		ttl:   ttl,
	}
}

func (s *SetTTL) Slice() []net.HardwareAddr {
	s.retrieve <- struct{}{}
	return <-s.macArr
}

func delMac(val string, c chan string) func() {
	return func() {
		c <- val
	}
}

func (s *SetTTL) daemon(ctx context.Context) {
	for {
		select {
		case newMac := <-s.toAdd:
			if timer, contains := s.m[newMac.value]; contains {
				timer.Reset(newMac.ttl)
			} else {
				s.m[newMac.value] = time.AfterFunc(
					newMac.ttl,
					delMac(newMac.value, s.toDel),
				)
			}
		case toDel := <-s.toDel:
			delete(s.m, toDel)
		case <-s.retrieve:
			res := make([]net.HardwareAddr, len(s.m), len(s.m))
			index := 0
			for k, _ := range s.m {
				res[index] = net.HardwareAddr(k)
				index += 1
			}
			s.macArr <- res
		case <-ctx.Done():
			close(s.toAdd)
			close(s.toDel)
			return
		}
	}

}
