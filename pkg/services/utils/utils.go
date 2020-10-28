package utils

import "math/rand"

var chars = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

// RandString generates random string with length equal to n from given
// random source.
func RandString(n int, rnd *rand.Rand) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = chars[rnd.Intn(len(chars))]
	}
	return string(b)
}
