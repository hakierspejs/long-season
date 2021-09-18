// Package requests contains utilities for retrieving
// data stored in http.Request reference.
package requests

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/hakierspejs/long-season/pkg/services/ctxkey"
)

var ErrValueNotFound = errors.New("requests: value not found in ctx")
var ErrEmptyParam = errors.New("requested URL parameter is empty")

// UserID returns user id from url.
func UserID(r *http.Request) (string, error) {
	res := chi.URLParam(r, "user-id")
	if res == "" {
		return "", ErrEmptyParam
	}
	return res, nil
}

func DeviceID(r *http.Request) (string, error) {
	res := chi.URLParam(r, "device-id")
	if res == "" {
		return "", ErrEmptyParam
	}
	return res, nil
}

func Debug(r *http.Request) (bool, error) {
	mode, ok := r.Context().Value(ctxkey.DebugKey).(bool)
	if !ok {
		return false, ErrValueNotFound
	}
	return mode, nil
}
