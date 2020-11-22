package mock

import (
	"context"
	"testing"

	"github.com/matryer/is"

	"github.com/hakierspejs/long-season/pkg/models"
	"github.com/hakierspejs/long-season/pkg/services/users"
)

func TestNewUsers(t *testing.T) {
	is := is.New(t)

	ctx := context.Background()
	s := New().Users()

	usersData := []models.User{
		{
			UserPublicData: models.UserPublicData{
				Nickname: "user3000",
			},
			Password: []byte("71Hk4Rt2WY8xqgYoKxPm"),
		},
		{
			UserPublicData: models.UserPublicData{
				Nickname: "user3001",
			},
			Password: []byte("u8dXHRi0JNo23JVeHkjh"),
		},
		{
			UserPublicData: models.UserPublicData{
				Nickname: "user3002",
			},
			Password: []byte("6eaOciUcg5EGSTkfQYvL"),
		},
	}

	for i, u := range usersData {
		id, err := s.New(ctx, u)
		if err != nil {
			t.Fatalf("can not add new user: %v", u)
		}
		is.Equal(i, id)
	}

	for i, u := range usersData {
		readUser, err := s.Read(ctx, i)
		if err != nil {
			t.Fatalf("can not read user with 0 ID from mock. error: %s", err)
		}
		is.True(users.Equals(u, *readUser))
	}
}
