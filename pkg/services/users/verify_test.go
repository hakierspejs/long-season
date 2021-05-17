package users

import (
	"fmt"
	"testing"
)

func TestVerifyNickname(t *testing.T) {
	type test struct {
		in     string
		expect bool
	}

	tests := []test{
		{in: "a?.dsa??;,,", expect: false},
		{in: "a", expect: false},
		{in: "abcd", expect: true},
		{in: "abcdef", expect: true},
		{in: "abcdef?", expect: false},
		{in: "?abcdef", expect: false},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("verify %s nickname", tt.in), func(t *testing.T) {
			if got := VerifyNickname(tt.in); got != tt.expect {
				t.Errorf("want: %v, got: %v", tt.expect, got)
			}
		})
	}
}

func TestVerifyPassword(t *testing.T) {
	type test struct {
		in     string
		expect bool
	}

	tooLongPassword := func() string {
		res := "."
		for len(res) <= 50 {
			res += "."
		}
		return res
	}()

	tests := []test{
		{in: "a?.dsa??;,,", expect: true},
		{in: "a", expect: false},
		{in: "abcd", expect: false},
		{in: "abcdef", expect: true},
		{in: "abcdef?", expect: true},
		{in: "?abcdef", expect: true},
		{in: "         ", expect: false},
		{in: "dsda sdasdad", expect: false},
		{in: ",;.,?><:';", expect: true},
		{in: tooLongPassword, expect: false},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("verify %s password", tt.in), func(t *testing.T) {
			if got := VerifyPassword(tt.in); got != tt.expect {
				t.Errorf("want: %v, got: %v", tt.expect, got)
			}
		})
	}
}
