package handlers

import (
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
)

// Opener is generic function for reading files.
// For example: Opener can open files from filesystem or
// files embedded withing binary. Given string could be a
// path or other indicator.
type Opener func(string) ([]byte, error)

// FileServer sets up handler to serve static files within given URL path.
func FileServer(r chi.Router, path string, opener Opener) {
	if strings.ContainsAny(path, "{}*") {
		panic("FileServer does not permit any URL parameters.")
	}

	if path != "/" && path[len(path)-1] != '/' {
		r.Get(path, http.RedirectHandler(path+"/", 301).ServeHTTP)
		path += "/"
	}
	path += "*"

	r.Get(path, func(w http.ResponseWriter, r *http.Request) {
		filepath := strings.TrimPrefix(r.URL.Path, "/")

		if strings.HasSuffix(filepath, ".css") {
			w.Header().Add("Content-Type", "text/css")
		}

		if strings.HasSuffix(filepath, ".js") {
			w.Header().Add("Content-Type", "application/javascript")
		}

		file, err := opener(filepath)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		w.Write(file)
	})
}
