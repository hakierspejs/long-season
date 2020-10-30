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
	"github.com/hakierspejs/long-season/pkg/services/result"
)

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

	// Function for extracting JWT token in form of string.
	// Error message should be readable for users and do not
	// contains sensitive data.
	Extractor func(r *http.Request) (string, error)

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

			tokenStr, err := ops.Extractor(r)
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

// ApiAuth is middleware for authorization of long-season REST api.
func ApiAuth(config models.Config, optional bool) func(next http.Handler) http.Handler {
	return JWT(&JWTOptions{
		Optional:  optional,
		Secret:    []byte(config.JWTSecret),
		Algorithm: jwt.HS256,
		Extractor: func(r *http.Request) (string, error) {
			header := r.Header.Get("Authorization")
			if header == "" {
				return "", fmt.Errorf("Authorization header is empty.")
			}

			if !strings.HasPrefix(header, "Bearer") {
				return "", fmt.Errorf("JWT authorization header should has `Bearer ` preffix.")
			}

			token := strings.TrimPrefix(header, "Bearer ")
			return token, nil
		},
		ContextKey: "jwt-user",
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
	})
}
