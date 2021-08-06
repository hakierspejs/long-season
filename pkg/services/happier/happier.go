package happier

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/alioygur/gores"
	"github.com/hakierspejs/long-season/pkg/services/requests"
	"github.com/thinkofher/horror"
)

// Factory contains method for declaring
// http oriented errors.
type Factory struct {
	debug bool
}

// FromRequest returns factory that should be used only with
// given http request.
func FromRequest(r *http.Request) *Factory {
	debug, err := requests.Debug(r)
	if err != nil {
		debug = false
	}
	return &Factory{
		debug: debug,
	}
}

// NotFound implements http NotFound (404) error for horror.Error
// interface to use in long-season REST API.
func (f *Factory) NotFound(err error, message string) horror.Error {
	return &errorHandler{
		message: message,
		wrapped: err,
		code:    http.StatusNotFound,
		debug:   f.debug,
	}
}

// InternalServerError implements http internal server error (500)
// for horror.Error interface to use in long-season REST API.
func (f *Factory) InternalServerError(err error, message string) horror.Error {
	return &errorHandler{
		message: message,
		wrapped: err,
		code:    http.StatusInternalServerError,
		debug:   f.debug,
	}
}

// Conflict implements http conflict (409) error for horror.Error
// interface to use in long-season REST API.
func (f *Factory) Conflict(err error, message string) horror.Error {
	return &errorHandler{
		message: message,
		wrapped: err,
		code:    http.StatusConflict,
		debug:   f.debug,
	}
}

// BadRequest implements http bad request (400) error for horror.Error
// interface to use in long-season REST API.
func (f *Factory) BadRequest(err error, message string) horror.Error {
	return &errorHandler{
		message: message,
		wrapped: err,
		code:    http.StatusBadRequest,
		debug:   f.debug,
	}
}

type errorHandler struct {
	message string
	wrapped error
	code    int
	debug   bool
}

func (e *errorHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	res := &errorResponse{
		Data: &dataResponse{
			Code:    e.code,
			Message: e.message,
		},
	}

	if e.debug {
		res.Debug = &debugResponse{
			ErrorMessage: e.wrapped.Error(),
		}
	}
	gores.JSONIndent(w, e.code, res, defaultPrefix, defaultIndent)
}

type errorResponse struct {
	Data  interface{} `json:"error"`
	Debug interface{} `json:"debug,omitempty"`
}

type dataResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type debugResponse struct {
	ErrorMessage string `json:"err"`
}

func (e *errorHandler) Error() string {
	return fmt.Sprint("apierr: code=%d public-message=%s err=%w", e.code, e.message, e.wrapped)
}

func (e *errorHandler) Is(target error) bool {
	return errors.Is(e.wrapped, target)
}

const (
	defaultPrefix string = ""
	defaultIndent        = "    "
)
