package macs

import (
	"context"
	"net"
	"time"
)

// SetTTL is set for mac addresses with special daemon running
// in the background that will delete given mac addresses after
// specified amount of time (TTL).
//
// SetTTL is completely thread safe. Probably.
type SetTTL struct {
	m              map[string]*time.Timer
	toAdd          chan setItem
	toDel          chan string
	retrieveSignal chan struct{}
	macSlice       chan []net.HardwareAddr
}

type setItem struct {
	value string
	ttl   time.Duration
}

// NewSetTTL returns initialised pointer to SetTTL.
func NewSetTTL(ctx context.Context) *SetTTL {
	res := &SetTTL{
		m:              map[string]*time.Timer{},
		toAdd:          make(chan setItem),
		toDel:          make(chan string),
		retrieveSignal: make(chan struct{}),
		macSlice:       make(chan []net.HardwareAddr),
	}

	// start daemon in new goroutine
	go res.daemon(ctx)

	return res
}

// Push adds given HardwareAddr to set and setup its TTL.
// If given HardwareAddr is already in the set, it's reset
// its TTL to given amount of time duration.
func (s *SetTTL) Push(addr net.HardwareAddr, ttl time.Duration) {
	s.toAdd <- setItem{
		value: string(addr),
		ttl:   ttl,
	}
}

// Slice returns slice of current Hardware addresses.
func (s *SetTTL) Slice() []net.HardwareAddr {
	s.retrieveSignal <- struct{}{}
	return <-s.macSlice
}

func delMac(val string, c chan string) func() {
	return func() {
		c <- val
	}
}

// daemon runs infinite loop that will end when
// given context will be done.
func (s *SetTTL) daemon(ctx context.Context) {
	// oh boy, here starts fun: infinite select loop
	// with different branches
	for {
		// lets select from multiple channels
		// attached to given set, this way
		// we can ensure that everything
		// will be synced together
		select {
		case newMac := <-s.toAdd:
			// first scenario, client want to
			// ad new mac address to set

			// lets check if new mac address is already in the
			// map
			if timer, contains := s.m[newMac.value]; contains {
				// if it is, reset timer with given ttl value
				timer.Reset(newMac.ttl)
			} else {
				// if given mac address is not present
				// at the map, lets create new timer
				// that will send delete signal to our
				// daemon
				s.m[newMac.value] = time.AfterFunc(
					newMac.ttl,
					delMac(newMac.value, s.toDel),
				)
			}
		case toDel := <-s.toDel:
			// simple scenario: delete received mac
			// from our map and go on
			delete(s.m, toDel)
		case <-s.retrieveSignal:
			// we've just received retrieveSignal signal!
			// lets allocate new slice that we will
			// send to our client
			res := make([]net.HardwareAddr, len(s.m), len(s.m))

			// loop over collection of addresses.
			index := 0
			for k, _ := range s.m {
				// we have to cast every key to net.HardwareAddr
				// we can omit net.ParseMAC, because SetTTL permits
				// only net.HardwareAddr to Push, so we assume here
				// that client properly parsed HardwareAddr
				res[index] = net.HardwareAddr(k)
				index += 1
			}

			// send result to client
			s.macSlice <- res
		case <-ctx.Done():
			// context is Done, so we're closing channel and
			// return to escape from loop
			close(s.toAdd)
			close(s.toDel)
			return
		}

		// there are no mutexes lol
	}
}
