package main

import (
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/hakierspejs/long-season/pkg/services/config"
	"github.com/hakierspejs/long-season/pkg/services/handlers"
	"github.com/hakierspejs/long-season/pkg/storage/mock"
)

func main() {
	config := config.Env()
	factoryStorage := mock.New()

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello HS!"))
	})
	r.Route("/users", func(r chi.Router) {
		r.Get("/", handlers.UsersAll(factoryStorage.Users()))
		r.Post("/", handlers.UserCreate(factoryStorage.Users()))
		r.With(handlers.URLParamInjection("id")).Route("/{id}", func(r chi.Router) {
			r.Get("/", handlers.UserRead(factoryStorage.Users()))
			r.Delete("/", handlers.UserRemove(factoryStorage.Users()))
		})
	})

	http.ListenAndServe(config.Address(), r)
}
