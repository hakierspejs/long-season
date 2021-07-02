package apierr

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/alioygur/gores"
	"github.com/thinkofher/horror"
)

type errorHandler struct {
	message string
	wrapped error
	code    int
	debug   bool
}

type errorResponse struct {
	Data interface{} `json:"error"`
}

type responseData struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
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

func (e *errorHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	gores.JSONIndent(w, e.code, &errorResponse{
		Data: &responseData{
			Code:    e.code,
			Message: e.message,
		},
	}, defaultPrefix, defaultIndent)
}

type Factory struct {
	debug bool
}

func FromRequest(r *http.Request) *Factory {
	return &Factory{
		debug: false,
	}
}

func (f *Factory) NotFound(err error, message string) horror.Error {
	return &errorHandler{
		message: message,
		wrapped: err,
		code:    http.StatusNotFound,
		debug:   f.debug,
	}
}

func (f *Factory) InternalServerError(err error, message string) horror.Error {
	return &errorHandler{
		message: message,
		wrapped: err,
		code:    http.StatusInternalServerError,
		debug:   f.debug,
	}
}

func (f *Factory) Conflict(err error, message string) horror.Error {
	return &errorHandler{
		message: message,
		wrapped: err,
		code:    http.StatusConflict,
		debug:   f.debug,
	}
}

func (f *Factory) BadRequest(err error, message string) horror.Error {
	return &errorHandler{
		message: message,
		wrapped: err,
		code:    http.StatusBadRequest,
		debug:   f.debug,
	}
}
