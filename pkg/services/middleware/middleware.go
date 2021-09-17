package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-chi/chi"
	"github.com/hakierspejs/long-season/pkg/models"
	"github.com/hakierspejs/long-season/pkg/services/ctxkey"
	"github.com/hakierspejs/long-season/pkg/services/happier"
	"github.com/hakierspejs/long-season/pkg/services/requests"
	"github.com/hakierspejs/long-season/pkg/services/session"
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
			errFactory := happier.FromRequest(r)

			token, err := apiExtractor("Status")(r)
			if err != nil {
				errFactory.Unauthorized(
					fmt.Errorf("apiExtractor: %w", err),
					fmt.Sprintf("invalid authorization header"),
				).ServeHTTP(w, r)
				return
			}

			if token != c.UpdateSecret {
				errFactory.Unauthorized(
					fmt.Errorf("token != c.UpdateSecret"),
					fmt.Sprintf("invalid authorization token"),
				).ServeHTTP(w, r)
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

func apiExtractor(prefix string) func(r *http.Request) (string, error) {
	return func(r *http.Request) (string, error) {
		header := r.Header.Get("Authorization")
		if header == "" {
			// TODO(thinkofher) replace with generic error for very extractor
			return "", fmt.Errorf("Authorization header is empty.")
		}

		if !strings.HasPrefix(header, prefix+" ") {
			return "", fmt.Errorf("JWT authorization header should has `%s ` prefix.", prefix)
		}

		token := strings.TrimPrefix(header, prefix+" ")
		return token, nil
	}
}

// Private checks if given user id is equal to user id at current
// session state.
func Private(renewer session.Renewer) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fail := func(err error, w http.ResponseWriter, r *http.Request) {
				happier.FromRequest(r).Unauthorized(
					fmt.Errorf("Private middleware fail func: %w", err),
					"You are not allowed to operate at requested resources.",
				).ServeHTTP(w, r)
				return
			}

			userID, err := requests.UserID(r)
			if err != nil {
				fail(err, w, r)
				return
			}

			state, err := renewer.Renew(r)
			if err != nil {
				fail(err, w, r)
				return
			}

			if userID != state.UserID {
				fail(nil, w, r)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RedirectLoggedIn redirects logged in users to homepage.
func RedirectLoggedIn(renewer session.Renewer) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, err := renewer.Renew(r)
			if err == nil {
				http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
