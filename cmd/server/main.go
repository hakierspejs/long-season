package main

import (
	"log"
	"net/http"

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

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello HS!"))
	})

	r.Route("/users", func(r chi.Router) {
		r.Get("/", api.UsersAll(factoryStorage.Users()))
		r.Post("/", api.UserCreate(factoryStorage.Users()))
		r.With(lsmiddleware.UserID).Route("/{user-id}", func(r chi.Router) {
			r.Get("/", api.UserRead(factoryStorage.Users()))
			r.Delete("/", api.UserRemove(factoryStorage.Users()))
			r.With(
				lsmiddleware.ApiAuth(*config, false),
				lsmiddleware.Private,
			).Route("/devices", func(r chi.Router) {
				r.Get("/", api.UserDevices(factoryStorage.Devices()))
				r.Post("/", api.DeviceAdd(factoryStorage.Devices()))
				r.With(lsmiddleware.DeviceID).Route("/{device-id}", func(r chi.Router) {
					r.Get("/", api.DeviceRead(factoryStorage.Devices()))
					r.Delete("/", api.DeviceRemove(factoryStorage.Devices()))
					r.Patch("/", api.DeviceUpdate(factoryStorage.Devices()))
				})
			})
		})
	})

	r.Post("/login", api.ApiAuth(*config, factoryStorage.Users()))
	r.With(lsmiddleware.ApiAuth(*config, false)).Get("/secret", api.AuthResource())

	http.ListenAndServe(config.Address(), r)
}
