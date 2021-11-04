package sqlite

import (
	"context"
	"testing"

	"github.com/matryer/is"

	"github.com/hakierspejs/long-season/pkg/storage"
)

func TestUsers(t *testing.T) {
	is := is.New(t)

	ctx := context.Background()

	f, closer, err := NewFactory(":memory:")
	is.NoErr(err)
	defer closer()

	usersData := map[string]storage.UserEntry{
		"1": {
			ID:             "1",
			Nickname:       "user3000",
			HashedPassword: []byte("71Hk4Rt2WY8xqgYoKxPm"),
			Private:        false,
		},
		"2": {
			ID:             "2",
			Nickname:       "user3001",
			HashedPassword: []byte("u8dXHRi0JNo23JVeHkjh"),
			Private:        true,
		},
		"3": {
			ID:             "3",
			Nickname:       "user3002",
			HashedPassword: []byte("6eaOciUcg5EGSTkfQYvL"),
			Private:        false,
		},
	}

	s := f.Users()

	for _, u := range usersData {
		id, err := s.New(ctx, u)
		is.NoErr(err)
		is.True(id == u.ID)
	}

	for _, u := range usersData {
		readUser, err := s.Read(ctx, u.ID)
		is.NoErr(err)
		is.Equal(readUser.Private, u.Private)
		is.Equal(readUser.Nickname, u.Nickname)
		is.Equal(readUser.HashedPassword, u.HashedPassword)
		is.Equal(readUser.ID, u.ID)
	}

	allUsers, err := s.All(ctx)
	is.NoErr(err)
	is.True(len(allUsers) == len(usersData))

	for _, curr := range allUsers {
		u, found := usersData[curr.ID]
		is.True(found)
		is.Equal(u.Private, curr.Private)
		is.Equal(u.ID, curr.ID)
		is.Equal(u.Nickname, curr.Nickname)
		is.Equal(u.HashedPassword, curr.HashedPassword)
	}

	err = s.Update(ctx, "1", func(u *storage.UserEntry) error {
		u.HashedPassword = []byte("new password")
		u.Nickname = "new nickname"
		u.Private = true
		return nil
	})
	is.NoErr(err)

	newUser, err := s.Read(ctx, "1")
	is.NoErr(err)
	is.Equal(newUser.HashedPassword, []byte("new password"))
	is.Equal(newUser.Private, true)
	is.Equal(newUser.ID, "1")
	is.Equal(newUser.Nickname, "new nickname")

	err = s.Remove(ctx, "1")
	is.NoErr(err)

	deletedUser, err := s.Read(ctx, "1")
	is.True(err != nil)
	is.Equal(deletedUser, nil)
}
