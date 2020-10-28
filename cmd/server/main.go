package main

import (
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	bolt "go.etcd.io/bbolt"

	"github.com/hakierspejs/long-season/pkg/services/config"
	"github.com/hakierspejs/long-season/pkg/services/handlers"
	"github.com/hakierspejs/long-season/pkg/storage/memory"
)

func main() {
	config := config.Env()

	boltDB, err := bolt.Open(config.DatabasePath, 0666, nil)
	if err != nil {
		log.Fatal(err.Error())
	}
	defer boltDB.Close()

	factoryStorage, err := memory.New(boltDB)
	if err != nil {
		log.Fatal(err.Error())
	}

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)

	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
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
	r.Post("/login", handlers.ApiAuth(*config, factoryStorage.Users(), rnd))
	r.With(handlers.AuthMiddleware(*config, false)).Get("/secret", handlers.AuthResource())

	http.ListenAndServe(config.Address(), r)
}
