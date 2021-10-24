package set

import (
	"encoding/json"
	"sort"
	"testing"

	"github.com/matryer/is"
)

func TestStringMarshal(t *testing.T) {
	is := is.New(t)

	s := StringFromSlice([]string{"a", "b", "c", "d"})

	b, err := json.Marshal(s)
	is.NoErr(err)

	expected := []byte(`["a","b","c","d"]`)
	is.Equal(expected, b)
}

func TestStringUnmarshal(t *testing.T) {
	is := is.New(t)

	b := []byte(`["a", "b", "c", "d", "eeee"]`)
	parsed := NewString()

	err := json.Unmarshal(b, parsed)
	is.NoErr(err)

	for _, v := range []string{"a", "b", "c", "d", "eeee"} {
		is.True(parsed.Contains(v))
	}
}

func TestNewString(t *testing.T) {
	is := is.New(t)

	s := NewString()
	is.True(s.core != nil)
	is.True(len(s.core) == 0)
}

func TestStringPushAndContains(t *testing.T) {
	is := is.New(t)

	s := NewString()

	items := []string{"a", "c", "d", "eeee"}
	for _, v := range items {
		s.Push(v)
	}

	for _, v := range items {
		is.True(s.Contains(v))
	}
}

func TestStringFromSlice(t *testing.T) {
	is := is.New(t)

	items := []string{"a", "c", "d", "eeee"}
	s := StringFromSlice(items)

	for _, v := range items {
		is.True(s.Contains(v))
	}
}

func TestStringRemove(t *testing.T) {
	is := is.New(t)

	items := []string{"a", "c", "d", "eeee"}
	s := StringFromSlice(items)

	for _, v := range items {
		s.Remove(v)
	}

	for _, v := range items {
		is.True(!s.Contains(v))
	}
}

func TestStrangeEquation(t *testing.T) {
	is := is.New(t)
	is.True("eeee" > "a")
}

func TestStringItems(t *testing.T) {
	is := is.New(t)

	items := []string{"a", "c", "d", "eeee"}
	s := StringFromSlice(items)

	newItems := s.Items()

	sort.Strings(newItems)
	is.Equal(items, newItems)
}

func TestStringEquals(t *testing.T) {
	is := is.New(t)

	first := StringFromSlice([]string{"a", "c", "d", "lol", "amazing"})
	second := NewString()
	for _, v := range []string{"a", "c", "d", "lol", "amazing"} {
		second.Push(v)
	}

	is.True(first.Equals(*second))
}
