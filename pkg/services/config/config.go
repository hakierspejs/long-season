package config

import (
	"os"

	"github.com/hakierspejs/long-season/pkg/models"
)

const (
	hostEnv = "LS_HOST"
	portEnv = "LS_PORT"
	boltENV = "LS_BOLT_DB"
)

// Env returns pointer to models.Config which is
// parsed from environmental variables. Cannot be nil.
// Unset variables will be
func Env() *models.Config {
	return &models.Config{
		Host:         DefaultEnv(hostEnv, "127.0.0.1"),
		Port:         DefaultEnv(portEnv, "3000"),
		DatabasePath: DefaultEnv(boltENV, "long-season.db"),
	}
}

// DefaultEnv returns content of shell variable
// assigned to given key. If result is empty, returns
// fallback value.
func DefaultEnv(key, fallback string) string {
	res := os.Getenv(key)
	if res == "" {
		return fallback
	}
	return res
}
