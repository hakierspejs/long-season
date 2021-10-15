// Package set implements classic data structure
// for operating with unique values.
package set

import (
	"encoding/json"
	"fmt"
	"sort"
	"sync"
)

// String is implementation of set for strings.
type String struct {
	core map[string]struct{}
	mtx  *sync.RWMutex
}

// NewString returns allocated string set.
func NewString() *String {
	return &String{
		core: map[string]struct{}{},
		mtx:  &sync.RWMutex{},
	}
}

// StringFromSlice returns allocated string set with
// items from given slice.
func StringFromSlice(items []string) *String {
	res := NewString()
	for _, v := range items {
		res.Push(v)
	}
	return res
}

// Push given item into set.
func (s *String) Push(item string) {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	s.core[item] = struct{}{}
}

// Remove given item from set.
func (s *String) Remove(item string) {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	delete(s.core, item)
}

// Contains returns true if given item is in set.
func (s *String) Contains(item string) bool {
	s.mtx.RLock()
	defer s.mtx.RUnlock()
	_, res := s.core[item]
	return res
}

// Items returns slice with all elements of set.
func (s *String) Items() []string {
	s.mtx.RLock()
	defer s.mtx.RUnlock()
	res := make([]string, len(s.core), len(s.core))

	i := 0
	for v := range s.core {
		res[i] = v
		i += 1
	}
	return res
}

// Equals returns true if both sets have the
// same values.
func (s *String) Equals(other String) bool {
	for _, v := range other.Items() {
		if !s.Contains(v) {
			return false
		}
	}
	return true
}

func (s *String) MarshalJSON() ([]byte, error) {
	res := s.Items()
	sort.Slice(res, func(i, j int) bool {
		return res[i] < res[j]
	})
	return json.Marshal(res)
}

func (s *String) UnmarshalJSON(data []byte) error {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	slice := []string{}

	if err := json.Unmarshal(data, &slice); err != nil {
		return fmt.Errorf("json.Unmarshal: %w", err)
	}

	if s.core == nil {
		s.core = map[string]struct{}{}
	}

	for _, v := range slice {
		s.core[v] = struct{}{}
	}

	return nil
}
