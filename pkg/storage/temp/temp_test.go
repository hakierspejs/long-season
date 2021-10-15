package temp

import (
	"context"
	"sort"
	"testing"

	"github.com/matryer/is"
)

func TestOnlineUsers(t *testing.T) {
	t.Run("All", func(t *testing.T) {
		is := is.New(t)
		ctx := context.TODO()

		ou := NewOnlineUsers()
		ou.set["1"] = struct{}{}
		ou.set["2"] = struct{}{}
		ou.set["3"] = struct{}{}

		want := []string{"1", "2", "3"}
		got, err := ou.All(ctx)
		is.NoErr(err)
		sort.Strings(got)

		is.Equal(got, want)
	})

	t.Run("Update", func(t *testing.T) {
		is := is.New(t)
		ctx := context.TODO()

		ou := NewOnlineUsers()
		ou.set["1"] = struct{}{}
		ou.set["2"] = struct{}{}
		ou.set["3"] = struct{}{}
		err := ou.Update(ctx, []string{"5", "4", "6", "8"})
		is.NoErr(err)

		want := map[string]struct{}{
			"4": struct{}{},
			"5": struct{}{},
			"6": struct{}{},
			"8": struct{}{},
		}
		got := ou.set
		is.Equal(got, want)
	})

	t.Run("IsOnline", func(t *testing.T) {
		is := is.New(t)
		ctx := context.TODO()

		ou := NewOnlineUsers()
		ou.set["1"] = struct{}{}
		ou.set["2"] = struct{}{}

		got, err := ou.IsOnline(ctx, "1")
		is.NoErr(err)
		is.True(got)

		got, err = ou.IsOnline(ctx, "2")
		is.NoErr(err)
		is.True(got)

		got, err = ou.IsOnline(ctx, "someid")
		is.NoErr(err)
		is.True(!got)
	})
}
