// Package jojo implements session interfaces for jwt
// tokens for long-season application.
package jojo

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/cristalhq/jwt/v3"

	"github.com/hakierspejs/long-season/pkg/models"
	"github.com/hakierspejs/long-season/pkg/services/happier"
	"github.com/hakierspejs/long-season/pkg/services/session"
)

// JWT implements session's interfaces for stateless session
// management.
type JWT struct {
	// Secret is jwt secret. The longer secret is, the better.
	Secret []byte

	// Algorithm is JWT algorithm used for signing
	Algorithm jwt.Algorithm

	// AppName is issuer application name.
	AppName string

	// CookieKey is key used to store API token in cookies.
	CookieKey string
}

const internalServerErrorResponse = "Internal server error. Please try again later."

// Tokenize outputs JWT token representation of given state.
func (j *JWT) Tokenize(ctx context.Context, s session.State) (string, error) {
	errFactory := happier.FromContext(ctx)

	signer, err := jwt.NewSignerHS(j.Algorithm, []byte(j.Secret))
	if err != nil {
		return "", errFactory.InternalServerError(
			fmt.Errorf("jwt.NewSignerHS: %w", err),
			internalServerErrorResponse,
		)
	}

	builder := jwt.NewBuilder(signer)
	now := time.Now()
	token, err := builder.Build(&models.Claims{
		StandardClaims: jwt.StandardClaims{
			Issuer:    j.AppName,
			Audience:  []string{"ls-apiv1"},
			Subject:   "auth",
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Hour * 48)),
			IssuedAt:  jwt.NewNumericDate(now),
			ID:        s.ID,
		},
		Nickname: s.Nickname,
		UserID:   s.UserID,
		Values:   s.Values,
	})
	if err != nil {
		return "", errFactory.InternalServerError(
			fmt.Errorf("builder.Build: %w", err),
			internalServerErrorResponse,
		)
	}

	return token.String(), nil
}

// Save is method for returning session data or session identifier to client.
func (j *JWT) Save(ctx context.Context, w http.ResponseWriter, s session.State) error {
	token, err := j.Tokenize(ctx, s)
	if err != nil {
		return err
	}

	http.SetCookie(w, &http.Cookie{
		Name:     j.CookieKey,
		Expires:  time.Now().Add(time.Hour * 48),
		Value:    token,
		HttpOnly: true,
		Path:     "/",
	})

	return nil
}

func (j *JWT) parseToken(ctx context.Context, token string) (*session.State, error) {
	fail := func(err error) (*session.State, error) {
		return nil, happier.FromContext(ctx).Unauthorized(
			fmt.Errorf("jojo: failed to parse token: %w", err),
			"Failed to parse authorization token.",
		)
	}

	verifier, err := jwt.NewVerifierHS(j.Algorithm, j.Secret)
	if err != nil {
		return fail(err)
	}

	newToken, err := jwt.ParseAndVerifyString(token, verifier)
	if err != nil {
		return fail(err)
	}

	newClaims := new(models.Claims)
	err = json.Unmarshal(newToken.RawClaims(), newClaims)
	if err != nil {
		return fail(err)
	}

	now := time.Now()
	if !newClaims.IsValidAt(now) {
		return fail(fmt.Errorf("token had expired"))
	}

	return &session.State{
		ID:       newClaims.ID,
		UserID:   newClaims.UserID,
		Nickname: newClaims.Nickname,
		Values:   newClaims.Values,
	}, nil
}

// RenewStrategy is functional implementation of session's
// Renewer interface.
type RenewStrategy func(*http.Request) (*session.State, error)

// Renew is method for restoring session from provided request
// data from client.
func (f RenewStrategy) Renew(r *http.Request) (*session.State, error) {
	return f(r)
}

// RenewFromCookies returns Renewer for retrieving session from
// http cookies.
func (j *JWT) RenewFromCookies() RenewStrategy {
	return func(r *http.Request) (*session.State, error) {
		errFactory := happier.FromRequest(r)

		cookie, err := r.Cookie(j.CookieKey)
		if err != nil {
			return nil, errFactory.Unauthorized(
				fmt.Errorf("r.Cookie: %w", err),
				"Failed to read authorization cookie.",
			)
		}

		valid := cookie.Expires.Before(time.Now())
		now := time.Now()
		if !valid {
			return nil, errFactory.Unauthorized(
				fmt.Errorf("cookie.Expires.Before: %v", now),
				"Failed to read authorization cookie.",
			)
		}

		return j.parseToken(r.Context(), cookie.Value)
	}
}

// RenewFromHeaderToken returns Renewer for retrieving session from
// given header with given prefix. For example for header equal to
// "Authorization" and prefix equal to "Bearer" returned RenewStrategy
// will retrieve JWT token from "Authorization" header in the form of
// "Bearer $TOKEN" where "$TOKEN" is tokenized JWT session State.
func (j *JWT) RenewFromHeaderToken(header, prefix string) RenewStrategy {
	return func(r *http.Request) (*session.State, error) {
		errFactory := happier.FromRequest(r)

		header := r.Header.Get(header)
		if header == "" {
			return nil, errFactory.Unauthorized(
				fmt.Errorf("jojo: %s header is empty", header),
				"Failed to read authorization header.",
			)
		}

		if !strings.HasPrefix(header, prefix+" ") {
			return nil, errFactory.Unauthorized(
				fmt.Errorf("jojo: %s header should has '%s ' prefix", header, prefix),
				"Failed to parse authorization header.",
			)
		}

		token := strings.TrimPrefix(header, prefix+" ")

		return j.parseToken(r.Context(), token)
	}
}

// Kill is method for purging JWT session by setting
// http cookie max age to -1.
func (j *JWT) Kill(_ context.Context, w http.ResponseWriter) error {
	http.SetCookie(w, &http.Cookie{
		Name:     j.CookieKey,
		Value:    "",
		HttpOnly: true,
		MaxAge:   -1,
		Path:     "/",
	})
	return nil
}
