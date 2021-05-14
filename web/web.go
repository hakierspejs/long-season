package web

import (
	"embed"
)

//go:embed static/*
//go:embed tmpl/*
var content embed.FS

// Open opens file with given path, that is stored in
// long-season binary.
func Open(path string) ([]byte, error) {
	return content.ReadFile(path)
}
