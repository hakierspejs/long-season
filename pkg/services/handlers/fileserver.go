package handlers

import (
	"net/http"
	"strings"

	"github.com/go-chi/chi"

	"github.com/hakierspejs/long-season/pkg/static"
)

// FileServer conveniently sets up a http.FileServer handler to serve
// static files from a http.FileSystem.
func FileServer(r chi.Router, path string, root http.FileSystem) {
	if strings.ContainsAny(path, "{}*") {
		panic("FileServer does not permit any URL parameters.")
	}

	if path != "/" && path[len(path)-1] != '/' {
		r.Get(path, http.RedirectHandler(path+"/", 301).ServeHTTP)
		path += "/"
	}
	path += "*"

	r.Get(path, func(w http.ResponseWriter, r *http.Request) {
		rctx := chi.RouteContext(r.Context())
		pathPrefix := strings.TrimSuffix(rctx.RoutePattern(), "/*")
		fs := http.StripPrefix(pathPrefix, http.FileServer(root))
		fs.ServeHTTP(w, r)
	})
}

// EmbeddedFileServer sets up handler to server static files
// embedded with embedfiles utility.
func EmbeddedFileServer(r chi.Router, path string) {
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
			w.Header().Add("Content-Type", "text/js")
		}

		file, err := static.Open(filepath)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		w.Write(file)
	})
}
