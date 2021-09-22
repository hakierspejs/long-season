package session

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/google/uuid"

	"github.com/hakierspejs/long-season/pkg/services/happier"
)

// State holds common data for storing in single
// user session.
type State struct {
	// ID is unique identifier of single session.
	ID string

	// UserID is user ID of session's owner.
	UserID string

	// Nickname is user name of session's owner.
	Nickname string

	// Values are key/value storage for additional
	// session data.
	Values map[string]interface{}
}

// Saver saves data in database and returns public information
// about session to client.
type Saver interface {
	// Save is method for returning session data or
	// session identifier to client.
	Save(context.Context, http.ResponseWriter, State) error
}

type Option func(*State)

// WithOptionsArguments holds arguments for
// WithOptions function. All arguments are
// required.
type WithOptionsArguments struct {
	Saver   Saver
	Writer  http.ResponseWriter
	Options []Option
}

// WithOptions applies options to given state and saves it.
// It does nothing more thatn regular Saver if you won't provide
// options.
func WithOptions(ctx context.Context, state State, args WithOptionsArguments) error {
	for _, op := range args.Options {
		op(&state)
	}

	return args.Saver.Save(ctx, args.Writer, state)
}

// Renewer retrieves session data from http request.
type Renewer interface {
	// Renew is method for restoring session from
	// provided request data from client.
	Renew(*http.Request) (*State, error)
}

// Killer purges session. It can make given session
// implementation expired.
type Killer interface {
	// Kill is method for purging session.
	Kill(context.Context, http.ResponseWriter) error
}

// Builder contains arguments and dependencies
// for building new session for user with given ID.
type Builder struct {
	UserID string

	Nickname string

	Values map[string]interface{}
}

// New returns fresh session for given user and saves it in
// database. You can output returned session to user after
// tokenizing it.
func New(ctx context.Context, b Builder) *State {
	values := map[string]interface{}{}
	if b.Values != nil {
		for k, v := range b.Values {
			values[k] = v
		}
	}

	return &State{
		ID:       uuid.New().String(),
		UserID:   b.UserID,
		Nickname: b.Nickname,
		Values:   values,
	}
}

// Guard returns http middleware which guards from
// clients accessing given handler without valid session.
func Guard(renewer Renewer) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, err := renewer.Renew(r)
			if err != nil {
				happier.FromRequest(r).Unauthorized(
					fmt.Errorf("session.SessionGuard.renewer.Renew: %w", err),
					"Invalid session. Please login in.",
				).ServeHTTP(w, r)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

type renewerFunc func(*http.Request) (*State, error)

func (f renewerFunc) Renew(r *http.Request) (*State, error) {
	return f(r)
}

// ErrNoRenewers is returned by Renewer composed by RenewerComposite
// when there are no renewers to run.
var ErrNoRenewers = errors.New("session: you haven't provided any renewers")

// RenewerComposite compose multiple Renewers into single one.
// Composed renewer will run each renewer till the first will
// successfully return any State.
func RenewerComposite(renewers ...Renewer) Renewer {
	return renewerFunc(func(r *http.Request) (*State, error) {
		var (
			res *State
			err error
		)

		for _, renewer := range renewers {
			res, err = renewer.Renew(r)
			if res != nil && err == nil {
				// Successfully retrieved session, we can
				// early return now.
				return res, nil
			}
		}
		if err != nil {
			return nil, err
		}

		return nil, ErrNoRenewers
	})
}
