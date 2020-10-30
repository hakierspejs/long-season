// Package result implements utilities for writing different
// data formats to http.ResponseWriter.
package result

import (
	"net/http"

	"github.com/alioygur/gores"
)

const (
	defaultPrefix string = ""
	defaultIndent        = "    "
)

// JSONErrorBody implements arguments for
// JSONError function.
type JSONErrorBody struct {
	Code    int    `json:"code"`
	Type    string `json:"type"`
	Message string `json:"message"`
}

// JSONError writes json error message wth specified http return
// code represented by JSONErrorBody to given http.ResponseWriter.
func JSONError(w http.ResponseWriter, b *JSONErrorBody) {
	gores.JSONIndent(w, b.Code, b, defaultPrefix, defaultIndent)
}
