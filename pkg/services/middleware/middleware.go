package middleware

import (
	"context"
	"net/http"

	"github.com/go-chi/chi"
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
