// Package toussaint implements service logic for manipulating two
// factor methods.
package toussaint

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hakierspejs/long-season/pkg/models"
	"github.com/hakierspejs/long-season/pkg/services/happier"
	"github.com/hakierspejs/long-season/pkg/services/session"
	"github.com/hakierspejs/long-season/pkg/storage"

	"github.com/pquerna/otp/totp"
)

// Method implements method function which returns
// query-able two factor data.
type Method interface {
	Method(userID string) models.TwoFactorMethod
}

// CollectMethods build slice of valeus that implements
// Method interface from multiple slices of Methods.
func CollectMethods(userID string, tf models.TwoFactor) []Method {
	res := []Method{}

	for _, v := range tf.OneTimeCodes {
		res = append(res, v)
	}

	return res
}

// Find returns first two factor method for user with given user id
// which is predictable by given function.
func Find(s []Method, userID string, f func(m models.TwoFactorMethod) bool) *models.TwoFactorMethod {
	methods := make([]models.TwoFactorMethod, len(s), len(s))
	for i, v := range s {
		methods[i] = v.Method(userID)
	}

	for _, v := range methods {
		if f(v) {
			return &v
		}
	}

	return nil
}

// IsTwoFactorEnabled checks whether some user
// has enabled any two factor method.
func IsTwoFactorEnabled(tf models.TwoFactor) bool {
	sum := len(tf.OneTimeCodes)
	return sum > 0
}

const (
	twoFactorRequiredKey = "2fa"
	totpKey              = "totp"
)

// TwoFactorRequired is session's Option. It forces
// session's owner to authenticate with two factor method.
func TwoFactorRequired(required bool) session.Option {
	return func(state *session.State) {
		state.Values[twoFactorRequiredKey] = required
	}
}

// AuthenticationWithTOTP is session's Option. It enables or disables
// two factor authentication with OTP codes.
func AuthenticationWithTOTP(totpEnabled bool) session.Option {
	return func(state *session.State) {
		state.Values[totpKey] = totpEnabled
	}
}

func readBoolFromInterfaceMap(m map[string]interface{}, key string) bool {
	res, ok := m[key].(bool)
	if !ok {
		return false
	}
	return res
}

// IsTwoFactorRequired return true if current session requires
// two factor authentication.
func IsTwoFactorRequired(state session.State) bool {
	return readBoolFromInterfaceMap(state.Values, twoFactorRequiredKey)
}

// IsTOTPEnabled check whether OTP codes are enabled
// for given session.
func IsTOTPEnabled(state session.State) bool {
	return readBoolFromInterfaceMap(state.Values, totpKey)
}

// Guard returns http middleware which guards from
// clients accessing given handler without completed
// two factor authentication.
func Guard(db storage.TwoFactor, renewer session.Renewer) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			errFactory := happier.FromRequest(r)

			s, err := renewer.Renew(r)
			if err != nil {
				errFactory.Unauthorized(
					fmt.Errorf("session.SessionGuard.renewer.Renew: %w", err),
					"Invalid session. Please login in.",
				).ServeHTTP(w, r)
				return
			}

			if IsTwoFactorRequired(*s) {
				errFactory.Unauthorized(
					fmt.Errorf("IsTwoFactorRequired is true"),
					"Two Factor authentication is required.",
				).ServeHTTP(w, r)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// Cleaner purges current session if it has enabled two factor authentication
// and redirects user to given redirect URI.
//
// Sessions with valid two factor authenticate will pass this middleware
// without any effects.
func Cleaner(renewer session.Renewer, killer session.Killer, redirectURI string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			errFactory := happier.FromRequest(r)

			s, err := renewer.Renew(r)
			if err != nil {
				// There is no session, so there is no
				// need to kill sessions without two factor
				// authentication.
				next.ServeHTTP(w, r)
				return
			}

			if IsTwoFactorRequired(*s) {
				err := killer.Kill(r.Context(), w)
				if err != nil {
					errFactory.InternalServerError(
						fmt.Errorf("killer.Kill: %w", err),
						"Failed to kill session.",
					).ServeHTTP(w, r)
					return
				}
				http.Redirect(w, r, redirectURI, http.StatusSeeOther)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// TwoFactorOnly accepts request with required two factor only.
// Otherwise redirects to given redirect URI.
func TwoFactorOnly(renewer session.Renewer, redirectURI string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			s, err := renewer.Renew(r)
			if err != nil {
				happier.FromRequest(r).Unauthorized(
					fmt.Errorf("renewer.Renew: %w", err),
					"Invalid session. Please login in.",
				)
				return
			}

			if !IsTwoFactorRequired(*s) {
				http.Redirect(w, r, redirectURI, http.StatusSeeOther)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// CodeValidator validates given code against implemented
// two factor authentication.
type CodeValidator interface {
	// Validate given code against implemented authentication
	// method.
	Validate(ctx context.Context, code string) bool
}

type funcValidator func(ctx context.Context, code string) bool

// Validate given code against implemented authentication
// method.
func (f funcValidator) Validate(ctx context.Context, code string) bool {
	return f(ctx, code)
}

// ValidatorTOTP returns validator for one time codes stored
// in two factor methods model.
func ValidatorTOTP(tf models.TwoFactor) CodeValidator {
	return funcValidator(func(ctx context.Context, code string) bool {
		for _, method := range tf.OneTimeCodes {
			if totp.Validate(code, method.Secret) {
				return true
			}
		}

		return false
	})
}
