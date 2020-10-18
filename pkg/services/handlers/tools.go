package handlers

import (
	"net/http"

	"github.com/alioygur/gores"
)

const (
	defaultPrefix string = ""
	defaultIndent        = "    "
)

type jsonErrorBody struct {
	Code    int    `json:"code"`
	Type    string `json:"type"`
	Message string `json:"message"`
}

func jsonError(w http.ResponseWriter, b *jsonErrorBody) {
	gores.JSONIndent(w, b.Code, b, defaultPrefix, defaultIndent)
}
