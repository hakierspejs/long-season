package macs

import "net"

// Set is unordered data structure for storing hardware addresses
// that doesn't allow duplicate values.
type CounterSet struct {
	m     map[string]int
	limit int
}

func NewCounterSet(limit int) CounterSet {
	return CounterSet{
		limit: limit,
		m:     map[string]int{},
	}
}

func (s CounterSet) contains(a net.HardwareAddr) bool {
	_, ok := s.m[a.String()]
	return ok
}

func (s CounterSet) Incr(a net.HardwareAddr) {
	if !s.contains(a) {
		s.m[a.String()] = 1
		return
	}

	if s.m[a.String()] < s.limit {
		s.m[a.String()] += 1
	}
}

func (s CounterSet) Decr(a net.HardwareAddr) {
	if !s.contains(a) {
		return
	}

	if s.m[a.String()] > 1 {
		s.m[a.String()] -= 1
		return
	}

	delete(s.m, a.String())
}

func (s CounterSet) Slice() []net.HardwareAddr {
	res := make([]net.HardwareAddr, len(s.m), len(s.m))
	i := 0
	for k, _ := range s.m {
		res[i] = net.HardwareAddr(k)
		i += 1
	}
	return res
}
