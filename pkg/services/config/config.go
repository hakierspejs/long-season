package config

import (
	"os"
	"strconv"
	"strings"
	"time"

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

	debugEnv     = "LS_DEBUG"
	defaultDebug = "0"

	refreshTimeEnv     = "LS_REFRESH_TIME"
	defaultRefreshTime = time.Duration(60) // seconds

	singleAddrTTLEnv     = "LS_SINGLE_ADDR_TTL"
	defaultSingleAddrTTL = time.Duration(60 * 5) // seconds

	spaceEnv     = "LS_SPACE"
	defaultSpace = "hs"

	cityEnv     = "LS_CITY"
	defaultCity = "lodz"
)

// Env returns pointer to models.Config which is
// parsed from environmental variables. Cannot be nil.
// Unset variables will be
func Env() *models.Config {
	return &models.Config{
		Space:         DefaultEnv(spaceEnv, defaultSpace),
		City:          DefaultEnv(cityEnv, defaultCity),
		Debug:         parseBoolEnv(DefaultEnv(debugEnv, defaultDebug)),
		Host:          DefaultEnv(hostEnv, defaultHost),
		Port:          DefaultEnv(portEnv, defaultPort),
		DatabasePath:  DefaultEnv(boltENV, defaultBoltDB),
		JWTSecret:     DefaultEnv(jwtSecretEnv, defaultJWTSecret),
		UpdateSecret:  DefaultEnv(updateSecretEnv, defaultUpdateSecret),
		AppName:       DefaultEnv(appNameEnv, defaultAppName),
		RefreshTime:   time.Second * DefaultDurationEnv(refreshTimeEnv, defaultRefreshTime),
		SingleAddrTTL: time.Second * DefaultDurationEnv(singleAddrTTLEnv, defaultSingleAddrTTL),
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

// DefaultIntEnv returns content of shell variable
// assigned to given key. If result is empty or
// parsing process failed, returns fallback value.
func DefaultDurationEnv(key string, fallback time.Duration) time.Duration {
	res := os.Getenv(key)
	if res == "" {
		return fallback
	}

	parsed, err := strconv.ParseInt(res, 10, 64)
	if err != nil {
		return fallback
	}

	return time.Duration(parsed)
}

func parseBoolEnv(env string) bool {
	return !(env == "" || env == "0" || strings.ToLower(env) == "false")
}
