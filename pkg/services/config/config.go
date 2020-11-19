package config

import (
	"os"

	"github.com/hakierspejs/long-season/pkg/models"
)

// JWTUserKey is key for JWT claims stored in request context.
const JWTUserKey = "jwt-user"

const (
	hostEnv     = "LS_HOST"
	defaultHost = "127.0.0.1"

	portEnv     = "LS_PORT"
	defaultPort = "3000"

	boltENV       = "LS_BOLT_DB"
	defaultBoltDB = "long-season.db"

	jwtSecretEnv     = "LS_JWT_SECRET"
	defaultJWTSecret = "default-super-secret"

	updateSecretEnv     = "LS_UPDATE_SECRET"
	defaultUpdateSecret = "default-super-api-secret"

	appNameEnv     = "LS_APP"
	defaultAppName = "long-season-backend"
)

// Env returns pointer to models.Config which is
// parsed from environmental variables. Cannot be nil.
// Unset variables will be
func Env() *models.Config {
	return &models.Config{
		Host:         DefaultEnv(hostEnv, defaultHost),
		Port:         DefaultEnv(portEnv, defaultPort),
		DatabasePath: DefaultEnv(boltENV, defaultBoltDB),
		JWTSecret:    DefaultEnv(jwtSecretEnv, defaultJWTSecret),
		UpdateSecret: DefaultEnv(updateSecretEnv, defaultUpdateSecret),
		AppName:      DefaultEnv(appNameEnv, defaultAppName),
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
