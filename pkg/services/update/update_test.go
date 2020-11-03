package update

import (
	"testing"

	"github.com/matryer/is"
)

func TestString(t *testing.T) {
	is := is.New(t)

	old := "hello world"
	updated := ""

	is.Equal(old, String(old, updated))

	updated = "mariusz"

	is.Equal(updated, String(old, updated))
}

func TestBytes(t *testing.T) {
	is := is.New(t)

	old := []byte("hello world")
	updated := []byte("")

	is.Equal(old, Bytes(old, updated))

	updated = []byte("mariusz")

	is.Equal(updated, Bytes(old, updated))
}

func TestNullableBool(t *testing.T) {
	is := is.New(t)

	old := false
	var updated *bool = nil

	is.Equal(old, NullableBool(old, updated))

	updatedNotNull := true

	is.Equal(updatedNotNull, NullableBool(old, &updatedNotNull))
}
