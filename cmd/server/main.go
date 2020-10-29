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
	"github.com/hakierspejs/long-season/pkg/services/handlers/api/v1"
	lsmiddleware "github.com/hakierspejs/long-season/pkg/services/middleware"
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
		r.Get("/", api.UsersAll(factoryStorage.Users()))
		r.Post("/", api.UserCreate(factoryStorage.Users()))
		r.With(lsmiddleware.UserID).Route("/{user-id}", func(r chi.Router) {
			r.Get("/", api.UserRead(factoryStorage.Users()))
			r.Delete("/", api.UserRemove(factoryStorage.Users()))
		})
	})
	r.Post("/login", api.ApiAuth(*config, factoryStorage.Users(), rnd))
	r.With(lsmiddleware.JWT(*config, false)).Get("/secret", api.AuthResource())

	http.ListenAndServe(config.Address(), r)
}
