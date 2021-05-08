package users

import (
	"fmt"
	"testing"

	"github.com/hakierspejs/long-season/pkg/models"
	"github.com/matryer/is"
)

func TestEquals(t *testing.T) {
	tests := []struct {
		desc        string
		expectedOut bool
		aUser       *models.User
		bUser       *models.User
	}{
		{
			desc:        "Positive test case; when equal",
			expectedOut: true,
			aUser:       &models.User{UserPublicData: models.UserPublicData{Nickname: "nickname1"}, Password: []byte{001}},
			bUser:       &models.User{UserPublicData: models.UserPublicData{Nickname: "nickname1"}, Password: []byte{001}},
		},
		{
			desc:        "Negative test case; when unequal - different nicknames",
			expectedOut: false,
			aUser:       &models.User{UserPublicData: models.UserPublicData{Nickname: "nickname2"}, Password: []byte{001}},
			bUser:       &models.User{UserPublicData: models.UserPublicData{Nickname: "nickname1"}, Password: []byte{001}},
		},
		{
			desc:        "Negative test case; when unequal - different passwords",
			expectedOut: false,
			aUser:       &models.User{UserPublicData: models.UserPublicData{Nickname: "nickname1"}, Password: []byte{001}},
			bUser:       &models.User{UserPublicData: models.UserPublicData{Nickname: "nickname1"}, Password: []byte{002}},
		},
	}
	for i, tc := range tests {
		is := is.New(t)
		t.Run(fmt.Sprintf("%d:%s", i, tc.desc), func(t *testing.T) {
			out := Equals(*tc.aUser, *tc.bUser)
			is.Equal(out, tc.expectedOut)
		})
	}
}

func TestStrictEquals(t *testing.T) {
	tests := []struct {
		desc        string
		expectedOut bool
		aUser       *models.User
		bUser       *models.User
	}{
		{
			desc:        "Positive test case; when equal",
			expectedOut: true,
			aUser:       &models.User{UserPublicData: models.UserPublicData{Nickname: "nickname1", Online: true, ID: 123}, Password: []byte{001}},
			bUser:       &models.User{UserPublicData: models.UserPublicData{Nickname: "nickname1", Online: true, ID: 123}, Password: []byte{001}},
		},
		{
			desc:        "Negative test case; when unequal - different online status",
			expectedOut: false,
			aUser:       &models.User{UserPublicData: models.UserPublicData{Nickname: "nickname1", Online: true, ID: 123}, Password: []byte{001}},
			bUser:       &models.User{UserPublicData: models.UserPublicData{Nickname: "nickname1", Online: false, ID: 123}, Password: []byte{001}},
		},
		{
			desc:        "Negative test case; when unequal - different ID",
			expectedOut: false,
			aUser:       &models.User{UserPublicData: models.UserPublicData{Nickname: "nickname1", Online: true, ID: 121}, Password: []byte{001}},
			bUser:       &models.User{UserPublicData: models.UserPublicData{Nickname: "nickname1", Online: true, ID: 123}, Password: []byte{001}},
		},
	}
	is := is.New(t)
	for i, tc := range tests {
		t.Run(fmt.Sprintf("%d:%s", i, tc.desc), func(t *testing.T) {

			out := StrictEquals(*tc.aUser, *tc.bUser)
			is.Equal(out, tc.expectedOut)
		})
	}
}

func TestUpdate(t *testing.T) {
	tests := []struct {
		desc         string
		onlineStatus bool
		aUser        *models.User
		changes      *Changes
	}{
		{
			desc:    "Positive test case; updating nickname and password",
			aUser:   &models.User{UserPublicData: models.UserPublicData{Nickname: "nickname1", Online: true, ID: 123}, Password: []byte{001}},
			changes: &Changes{Nickname: "nickname2", Password: []byte{002}},
		},
	}
	is := is.New(t)
	for i, tc := range tests {
		t.Run(fmt.Sprintf("%d:%s", i, tc.desc), func(t *testing.T) {
			retUser := Update(*tc.aUser, tc.changes)
			is.Equal(tc.changes.Nickname, retUser.UserPublicData.Nickname)
		})
	}
}

func TestPublicSlice(t *testing.T) {
	tests := []struct {
		desc        string
		expectedOut bool
		aPassword   []byte
		users       *[]models.User
	}{
		{
			desc:        "Positive test case",
			expectedOut: true,
			users:       &[]models.User{{Password: []byte{01}, UserPublicData: models.UserPublicData{ID: 123}}, {Password: []byte{02}, UserPublicData: models.UserPublicData{Nickname: "nickname"}}},
		},
	}
	for i, tc := range tests {
		t.Run(fmt.Sprintf("%d:%s", i, tc.desc), func(t *testing.T) {

			_ = PublicSlice(*tc.users)
		})
	}
}

func TestAll(t *testing.T) {
	tests := []struct {
		desc        string
		expectedOut bool
		arg         bool
	}{
		{
			desc:        "Positive test case; True args",
			expectedOut: true,
			arg:         true,
		},
		{
			desc:        "Negative test case; False args",
			expectedOut: false,
			arg:         false,
		},
	}
	is := is.New(t)
	for i, tc := range tests {

		t.Run(fmt.Sprintf("%d:%s", i, tc.desc), func(t *testing.T) {
			out := all(true, true, true, tc.arg)
			is.Equal(out, tc.expectedOut)
		})
	}
}

func TestFilter(t *testing.T) {
	data := []models.User{
		{
			Password: []byte("lol2lol3password"),
			Private:  false,
			UserPublicData: models.UserPublicData{
				ID:       1,
				Nickname: "lolmen",
				Online:   true,
			},
		},
		{
			Password: []byte("lol2lol3wordpass"),
			Private:  true,
			UserPublicData: models.UserPublicData{
				ID:       2,
				Nickname: "mariusz",
				Online:   false,
			},
		},
		{
			Password: []byte("212102121"),
			Private:  false,
			UserPublicData: models.UserPublicData{
				ID:       3,
				Nickname: "patryk",
				Online:   true,
			},
		},
		{
			Password: []byte("2137&2137"),
			Private:  false,
			UserPublicData: models.UserPublicData{
				ID:       4,
				Nickname: "patryka",
				Online:   false,
			},
		},
	}

	same := func(a []models.User, b []models.User) bool {
		for i, v := range a {
			if !StrictEquals(v, b[i]) {
				return false
			}
		}
		return true
	}

	type test struct {
		name    string
		want    []models.User
		filters []FilterFunc
	}

	tests := []test{
		{
			name: "filter only online users",
			want: []models.User{
				{
					Password: []byte("lol2lol3password"),
					Private:  false,
					UserPublicData: models.UserPublicData{
						ID:       1,
						Nickname: "lolmen",
						Online:   true,
					},
				},
				{
					Password: []byte("212102121"),
					Private:  false,
					UserPublicData: models.UserPublicData{
						ID:       3,
						Nickname: "patryk",
						Online:   true,
					},
				},
			},
			filters: []FilterFunc{Online},
		},
		{
			name: "filter only public users",
			want: []models.User{
				{
					Password: []byte("lol2lol3password"),
					Private:  false,
					UserPublicData: models.UserPublicData{
						ID:       1,
						Nickname: "lolmen",
						Online:   true,
					},
				},
				{
					Password: []byte("212102121"),
					Private:  false,
					UserPublicData: models.UserPublicData{
						ID:       3,
						Nickname: "patryk",
						Online:   true,
					},
				},
				{
					Password: []byte("2137&2137"),
					Private:  false,
					UserPublicData: models.UserPublicData{
						ID:       4,
						Nickname: "patryka",
						Online:   false,
					},
				},
			},
			filters: []FilterFunc{Not(Private)},
		},
		{
			name: "filter only private users",
			want: []models.User{
				{
					Password: []byte("lol2lol3wordpass"),
					Private:  true,
					UserPublicData: models.UserPublicData{
						ID:       2,
						Nickname: "mariusz",
						Online:   false,
					},
				},
			},
			filters: []FilterFunc{Private},
		},
		{
			name: "filter only offline users",
			want: []models.User{
				{
					Password: []byte("lol2lol3wordpass"),
					Private:  true,
					UserPublicData: models.UserPublicData{
						ID:       2,
						Nickname: "mariusz",
						Online:   false,
					},
				},
				{
					Password: []byte("2137&2137"),
					Private:  false,
					UserPublicData: models.UserPublicData{
						ID:       4,
						Nickname: "patryka",
						Online:   false,
					},
				},
			},
			filters: []FilterFunc{Not(Online)},
		},
		{
			name: "filter online public users",
			want: []models.User{
				{
					Password: []byte("lol2lol3password"),
					Private:  false,
					UserPublicData: models.UserPublicData{
						ID:       1,
						Nickname: "lolmen",
						Online:   true,
					},
				},
				{
					Password: []byte("212102121"),
					Private:  false,
					UserPublicData: models.UserPublicData{
						ID:       3,
						Nickname: "patryk",
						Online:   true,
					},
				},
			},
			filters: []FilterFunc{Not(Private), Online},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Filter(data, tt.filters...)
			if !same(result, tt.want) {
				t.Errorf("got: %v, want: %v", result, tt.want)
			}
		})
	}

}
