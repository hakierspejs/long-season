package middleware

import (
	"context"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/hakierspejs/long-season/pkg/models"
	"github.com/hakierspejs/long-season/pkg/services/ctxkey"
	"github.com/hakierspejs/long-season/pkg/services/result"
)

// URLParamInjection injects given chi parameter into request context.
func URLParamInjection(param string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			value := chi.URLParam(r, param)
			ctx := context.WithValue(r.Context(), param, value)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// UserID injects user id into request context.
func UserID(next http.Handler) http.Handler {
	return URLParamInjection("user-id")(next)
}

// DeviceID injects user id into request context.
func DeviceID(next http.Handler) http.Handler {
	return URLParamInjection("device-id")(next)
}

func UpdateAuth(c *models.Config) func(http.Handler) http.Handler {
	// TODO(thinkofher) Add doc.
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token, err := apiExtractor("Status")(r)
			if err != nil {
				result.JSONError(w, &result.JSONErrorBody{
					Code:    http.StatusUnauthorized,
					Message: "invalid authorization header",
					Type:    "unauthorized",
				})
				return
			}

			if token != c.UpdateSecret {
				result.JSONError(w, &result.JSONErrorBody{
					Code:    http.StatusUnauthorized,
					Message: "invalid authorization token",
					Type:    "unauthorized",
				})
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// Debug injects information about application debug mode
// to every http request's context.
func Debug(c models.Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), ctxkey.DebugKey, c.Debug)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
