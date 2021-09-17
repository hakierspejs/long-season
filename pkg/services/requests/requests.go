// Package requests contains utilities for retrieving
// data stored in http.Request reference.
package requests

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/hakierspejs/long-season/pkg/services/ctxkey"
)

var ErrValueNotFound = errors.New("requests: value not found in ctx")

func id(key string, r *http.Request) (int, error) {
	id, ok := r.Context().Value(key).(string)
	if !ok {
		return 0, errors.New("ID stored in context has inapropriate type.")
	}

	res, err := strconv.Atoi(id)
	if err != nil {
		return 0, err
	}

	return res, nil
}

// UserID returns user id from url.
func UserID(r *http.Request) (int, error) {
	return id("user-id", r)
}

func DeviceID(r *http.Request) (int, error) {
	return id("device-id", r)
}

func Debug(r *http.Request) (bool, error) {
	mode, ok := r.Context().Value(ctxkey.DebugKey).(bool)
	if !ok {
		return false, ErrValueNotFound
	}
	return mode, nil
}
