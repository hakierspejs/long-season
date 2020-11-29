package config

import (
	"io/ioutil"
	"os"
	"strings"

	"github.com/hakierspejs/long-season/pkg/models"
	"github.com/hakierspejs/long-season/pkg/static"
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

	debugEnv     = "LS_DEBUG"
	defaultDebug = "0"
)

// Env returns pointer to models.Config which is
// parsed from environmental variables. Cannot be nil.
// Unset variables will be
func Env() *models.Config {
	return &models.Config{
		Debug:        parseBoolEnv(DefaultEnv(debugEnv, defaultDebug)),
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

// Opener is generic function for reading files.
// For example: Opener can open files from filesystem or
// files embedded withing binary. Given string could be a
// path or other indicator.
type Opener func(string) ([]byte, error)

// MakeOpener returns opener depending on given config.
func MakeOpener(c *models.Config) Opener {
	if c.Debug {
		return ioutil.ReadFile
	}
	return static.Open
}

func parseBoolEnv(env string) bool {
	return !(env == "" || env == "0" || strings.ToLower(env) == "false")
}
