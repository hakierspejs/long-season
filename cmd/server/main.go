package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	bolt "go.etcd.io/bbolt"

	"github.com/hakierspejs/long-season/pkg/services/config"
	"github.com/hakierspejs/long-season/pkg/services/handlers"
	"github.com/hakierspejs/long-season/pkg/services/handlers/api/v1"
	lsmiddleware "github.com/hakierspejs/long-season/pkg/services/middleware"
	"github.com/hakierspejs/long-season/pkg/services/ui"
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
	r.Use(middleware.NoCache)

	r.Get("/", ui.Home())
	r.Get("/login", ui.LoginPage())

	r.Route("/api/v1", func(r chi.Router) {
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
		r.Put("/update", api.UpdateStatus(factoryStorage.Users(), factoryStorage.Devices()))
	})

	r.With(lsmiddleware.ApiAuth(*config, false)).Get("/who", handlers.Who())
	r.With(lsmiddleware.ApiAuth(*config, false)).Get("/devices", ui.Devices())
	r.Get("/logout", ui.Logout())
	r.Get("/register", ui.Register())

	workDir, _ := os.Getwd()
	filesDir := http.Dir(filepath.Join(workDir, "static"))
	handlers.FileServer(r, "/static", filesDir)

	http.ListenAndServe(config.Address(), r)
}
