package session

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
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
}

// Store can save and delete session with given ID.
type Store interface {
	Save(ctx context.Context, state State) error
	Delete(ctx context.Context, token string) error
}

// Renewer can turn valid token session string into
// session State.
type Renewer interface {
	Renew(ctx context.Context, token string) (*State, error)
}

// Tokenizer turns given State into string that
// can be used to read from state with help of Store in
// the future.
type Tokenizer interface {
	// Tokenize turns given State into string. Returned
	// string has to be safe to output to user (for example:
	// as a HTTP cookie).
	//
	// Token can be a full and signed or encrypted information
	// about session in the form of single string (for example JWT)
	// or can only holds information for identifying this particular
	// session (for example session ID).
	Tokenize(ctx context.Context, state State) string
}

type funcTokenizer func(s State) string

func (f funcTokenizer) Tokenize(ctx context.Context, state State) string {
	return state.ID
}

func TokenizeBydID() Tokenizer {
	return funcTokenizer(func(s State) string {
		return s.ID
	})
}

// Builder contains arguments and dependencies
// for building new session for user with given ID.
type Builder struct {
	UserID string

	Nickname string

	Store
}

// New returns fresh session for given user and saves it in
// database. You can output returned session to user after
// tokenizing it.
func New(ctx context.Context, b Builder) (*State, error) {
	res := State{
		ID:       uuid.New().String(),
		UserID:   b.UserID,
		Nickname: b.Nickname,
	}

	if err := b.Save(ctx, res); err != nil {
		return nil, fmt.Errorf("b.Save: %w", err)
	}

	return &res, nil
}

type cookieOption func(*http.Cookie)

// Clock knows current time.
type Clock interface {
	Now() time.Time
}

// Pusher outputs given session to User as http Cookie.
// Usage of Pusher is optional. See: Push method for more
// information about the whole process.
type Pusher struct {
	tokenizer  Tokenizer
	name       string
	cookieOpts []cookieOption
	clock      Clock
	store      Store
}

// Option represents optional argument for constructing
// new Pusher.
type Option func(*Pusher)

// NewPusher is safe constructor for Pusher type.
func NewPusher(t Tokenizer, s Store, c Clock, opts ...Option) *Pusher {
	res := &Pusher{
		name:      "ls-session", // default session name
		tokenizer: t,
		clock:     c,
		store:     s,
	}

	for _, op := range opts {
		op(res)
	}

	return res
}

func (p *Pusher) applyOptions(c *http.Cookie) {
	for _, op := range p.cookieOpts {
		op(c)
	}
}

// Push outputs given session to User as http Cookie.
// Usage of push is optional and can be handy if you're
// building SSR application.
//
// You can also output tokenized State to client on your own.
//
// By default saved cookie has name "ls-session" and expires after
// 4 hours. It is http-only and is applied to path equal to "/".
func (p *Pusher) Push(ctx context.Context, w http.ResponseWriter, s State) {
	token := p.tokenizer.Tokenize(ctx, s)

	cookie := &http.Cookie{
		Name:     p.name,
		Expires:  p.clock.Now().Add(time.Hour * 4),
		Value:    token,
		HttpOnly: true,
		Path:     "/",
	}

	p.applyOptions(cookie)
	http.SetCookie(w, cookie)
}

// Kill purges session associated with client of given response
// writer.
func (p *Pusher) Kill(ctx context.Context, w http.ResponseWriter) {
	cookie := &http.Cookie{
		Name:     p.name,
		Value:    "",
		HttpOnly: true,
		Path:     "/",
	}
	p.applyOptions(cookie)

	// Expire current session by setting its expiration
	// time to current time subtracted by one year.
	cookie.Expires = p.clock.Now().Add(-1 * time.Hour * 24 * 365)
	http.SetCookie(w, cookie)
}

// WithName changes name of session cookie.
func WithName(name string) Option {
	return func(p *Pusher) {
		p.name = name
		p.cookieOpts = append(p.cookieOpts, func(c *http.Cookie) {
			c.Name = name
		})
	}
}

// WithHTTPOnly changes HTTP only flag of session cookie.
func WithHTTPOnly(httpOnly bool) Option {
	return func(p *Pusher) {
		p.cookieOpts = append(p.cookieOpts, func(c *http.Cookie) {
			c.HttpOnly = httpOnly
		})
	}
}

// WithExpiration changes value of time after which
// session cookie expires. For example if you set
// expiration time to 4 hours, cookie pushed to client
// with Push method will expire in 4 hours after
// time of its creation.
func WithExpiration(at time.Duration) Option {
	return func(p *Pusher) {
		p.cookieOpts = append(p.cookieOpts, func(c *http.Cookie) {
			c.Expires = p.clock.Now().Add(at)
		})
	}
}

// WithPath changes path value of session cookie.
func WithPath(path string) Option {
	return func(p *Pusher) {
		p.cookieOpts = append(p.cookieOpts, func(c *http.Cookie) {
			c.Path = path
		})
	}
}

// WithSecure changes secure flag of session cookie.
func WithSecure(secure bool) Option {
	return func(p *Pusher) {
		p.cookieOpts = append(p.cookieOpts, func(c *http.Cookie) {
			c.Secure = secure
		})
	}
}
