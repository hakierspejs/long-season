package middleware

import (
	"context"
	"net/http"

	jwtmiddleware "github.com/auth0/go-jwt-middleware"
	"github.com/dgrijalva/jwt-go"
	"github.com/go-chi/chi"
	"github.com/hakierspejs/long-season/pkg/models"
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

// JWT checks if user provided valid jwt token and then injects
// its content with jwt-user key into context.
func JWT(config models.Config, optional bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		middleware := jwtmiddleware.New(jwtmiddleware.Options{
			ValidationKeyGetter: func(token *jwt.Token) (interface{}, error) {
				return []byte(config.JWTSecret), nil
			},
			CredentialsOptional: optional,
			SigningMethod:       jwt.SigningMethodHS256,
			UserProperty:        "jwt-user",
		})

		return middleware.Handler(next)
	}
}
