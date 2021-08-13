package happier

import (
	"fmt"
	"log"
	"net/http"

	"github.com/thinkofher/horror"
)

// Adapter adapts horror Handlers for usage with
// chi router, because it accepts only HandlerFuncs
// to its routing methods along `pattern`.
//
// Use NewAdapter as constructor if you want to use Adapter.
type Adapter struct {
	wrapped *horror.Adapter
}

// NewAdapter is the only proper constructor for
// Adapter type.
func NewAdapter() *Adapter {
	return &Adapter{
		wrapped: horror.NewAdapter(&horror.AdapterBuilder{
			BeforeError: func(err error, w http.ResponseWriter, r *http.Request) {
				log.Println("<error>", err)
			},
			InternalHandler: func(err error, w http.ResponseWriter, r *http.Request) {
				FromRequest(r).InternalServerError(
					err,
					fmt.Sprintf("internal server error, please try again later"),
				).ServeHTTP(w, r)
			},
		}),
	}
}

// WithError wraps given horror Handler, adapts it and returns
// http.HandlerFunc that you can use with chi router.
func (a *Adapter) WithError(h horror.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		a.wrapped.WithError(h).ServeHTTP(w, r)
	}
}
