package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/cristalhq/jwt/v3"
	"github.com/hakierspejs/long-season/pkg/models"
	"github.com/hakierspejs/long-season/pkg/services/config"
	"github.com/hakierspejs/long-season/pkg/services/requests"
	"github.com/hakierspejs/long-season/pkg/services/result"
)

// Extractor is function for extracting JWT token in form of
// string. Error message should be readable for users and do not
// contains sensitive data.
type Extractor func(r *http.Request) (string, error)

// JWTOptions contains different fields intended for creating new
// JWT middleware.
type JWTOptions struct {
	// Optional is flag, that will enable ignoring failed attempts
	// to login.
	Optional bool

	// Secret is jwt secret. The longer secret is, the better.
	Secret []byte

	// Algorithm is JWT algorithm used for signing
	Algorithm jwt.Algorithm

	// Unauthorized is handling unauthorized sessions with custom error
	// message.
	Unauthorized func(msg string, w http.ResponseWriter, r *http.Request)

	// InternalServerError handles cases when server fails during authorization.
	InternalServerError http.HandlerFunc

	// Functions for extracting JWT token in form of string.
	Extractors []Extractor

	// ContextKey will be set to JWT claims at request Context
	// after successful authorization.
	ContextKey string
}

func handlerToFunc(h http.Handler) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h.ServeHTTP(w, r)
	})
}

// JWT checks if user has provided valid jwt token and then injects
// its content with jwt-user key into context.
func JWT(ops *JWTOptions) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			verifier, err := jwt.NewVerifierHS(ops.Algorithm, ops.Secret)

			// If check if optional, just replace error handlers
			// with next handler to handle
			if ops.Optional {
				ops.InternalServerError = handlerToFunc(next)
				ops.Unauthorized = func(_ string, w http.ResponseWriter, r *http.Request) {
					next.ServeHTTP(w, r)
				}
			}

			if err != nil {
				ops.InternalServerError(w, r)
				return
			}

			tokenStr := ""
			for _, f := range ops.Extractors {
				tokenStr, err = f(r)
				if err == nil {
					break
				}
			}
			if err != nil {
				ops.Unauthorized(err.Error(), w, r)
				return
			}

			newToken, err := jwt.ParseAndVerifyString(tokenStr, verifier)
			if err != nil {
				ops.Unauthorized("Invalid token format", w, r)
				return
			}

			newClaims := new(models.Claims)
			err = json.Unmarshal(newToken.RawClaims(), newClaims)
			if err != nil {
				ops.InternalServerError(w, r)
				return
			}

			now := time.Now()
			if !newClaims.IsValidAt(now) {
				ops.Unauthorized("Token expired.", w, r)
				return
			}

			newCtx := context.WithValue(r.Context(), ops.ContextKey, newClaims)
			next.ServeHTTP(w, r.WithContext(newCtx))
		})
	}
}

func apiExtractor(r *http.Request) (string, error) {
	header := r.Header.Get("Authorization")
	if header == "" {
		return "", fmt.Errorf("Authorization header is empty.")
	}

	if !strings.HasPrefix(header, "Bearer") {
		return "", fmt.Errorf("JWT authorization header should has `Bearer ` preffix.")
	}

	token := strings.TrimPrefix(header, "Bearer ")
	return token, nil
}

func viewExtractor(r *http.Request) (string, error) {
	cookie, err := r.Cookie("jwt-token")
	if err != nil {
		return "", err
	}

	valid := cookie.Expires.Before(time.Now())
	if !valid {
		return "", fmt.Errorf("cookie is expired")
	}

	fmt.Println("jwt-token", cookie.Value)
	return cookie.Value, nil
}

func defaultJwtOptions(c models.Config, optional bool) JWTOptions {
	return JWTOptions{
		Optional:   optional,
		Secret:     []byte(c.JWTSecret),
		Algorithm:  jwt.HS256,
		Extractors: []Extractor{apiExtractor, viewExtractor},
		ContextKey: config.JWTUserKey,
		InternalServerError: func(w http.ResponseWriter, r *http.Request) {
			result.JSONError(w, &result.JSONErrorBody{
				Code:    http.StatusInternalServerError,
				Message: "Authorization failed due to internal error. Please try again.",
				Type:    "internal-server-error",
			})
			return
		},
		Unauthorized: func(msg string, w http.ResponseWriter, r *http.Request) {
			result.JSONError(w, &result.JSONErrorBody{
				Code:    http.StatusUnauthorized,
				Message: msg,
				Type:    "unauthorized",
			})
			return
		},
	}
}

// ApiAuth is middleware for authorization of long-season REST api.
func ApiAuth(c models.Config, optional bool) func(next http.Handler) http.Handler {
	options := defaultJwtOptions(c, optional)
	return JWT(&options)
}

// Private checks if given user id is equal to user id at JWT claims.
func Private(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fail := func(w http.ResponseWriter) {
			result.JSONError(w, &result.JSONErrorBody{
				Code:    http.StatusUnauthorized,
				Message: "You are not allowed to operate at requested resources.",
				Type:    "unauthorized",
			})
		}

		userID, err := requests.UserID(r)
		if err != nil {
			fail(w)
			return
		}

		claims, err := requests.JWTClaims(r)
		if err != nil {
			fail(w)
			return
		}

		if userID != claims.UserID {
			fail(w)
			return
		}

		next.ServeHTTP(w, r)
	})
}
