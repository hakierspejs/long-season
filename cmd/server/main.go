package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/cors"
	bolt "go.etcd.io/bbolt"

	"github.com/hakierspejs/long-season/pkg/services/config"
	"github.com/hakierspejs/long-season/pkg/services/handlers"
	"github.com/hakierspejs/long-season/pkg/services/handlers/api/v1"
	lsmiddleware "github.com/hakierspejs/long-season/pkg/services/middleware"
	"github.com/hakierspejs/long-season/pkg/services/status"
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

	ctx := context.Background()
	macChannel, macDeamon := status.NewDaemon(ctx, factoryStorage.Devices(), factoryStorage.Users())

	// CORS (Cross-Origin Resource Sharing) middleware that enables public
	// access to GET/OPTIONS requests. Used to expose APIs to XHR consumers in
	// other domains.
	publicCors := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "OPTIONS"},
	})
	// Logging CORS is useful to understand why a preflight request was denied.
	// This is not the healthiest way to log in chi, but even with this CORS
	// middleware being chi-specific, it doesn't seem to be able to do it in
	// any nicer (ie. request-oriented) way.
	if config.Debug {
		publicCors.Log = log.New(os.Stdout, "CORS ", 0)
	}

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(middleware.NoCache)

	r.Get("/", ui.Home(config))
	r.With(lsmiddleware.ApiAuth(*config, true), lsmiddleware.RedirectLoggedIn).Get("/login", ui.LoginPage(config))

	r.Route("/api/v1", func(r chi.Router) {
		r.Route("/users", func(r chi.Router) {
			// Use CORS to allow GET/OPTIONS to GET /api/v1/users from
			// anywhere.
			// This technically wraps a nil handler around publicCors, but the
			// nil handler never gets called. This is weirdness stemming from
			// how go-chi/cors is supposed to be applied globally to the entire
			// application, and not to particular endpoints.
			r.With(publicCors.Handler).Options("/", nil)
			r.With(publicCors.Handler).Get("/", api.UsersAll(factoryStorage.Users()))
			r.Post("/", api.UserCreate(factoryStorage.Users()))

			r.With(lsmiddleware.UserID).Route("/{user-id}", func(r chi.Router) {
				r.With(
					lsmiddleware.ApiAuth(*config, true),
				).Get("/", api.UserRead(factoryStorage.Users()))

				// Users can only delete themselves.
				r.With(
					lsmiddleware.ApiAuth(*config, false),
					lsmiddleware.Private,
				).Delete("/", api.UserRemove(factoryStorage.Users()))

				r.With(
					lsmiddleware.ApiAuth(*config, false),
					lsmiddleware.Private,
				).Patch("/", api.UserUpdate(factoryStorage.Users()))

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
		r.With(lsmiddleware.UpdateAuth(config)).Put(
			"/update",
			api.UpdateStatus(macChannel),
		)
	})

	r.With(lsmiddleware.ApiAuth(*config, false)).Get("/who", handlers.Who())
	r.With(lsmiddleware.ApiAuth(*config, false)).Get("/devices", ui.Devices(config))
	r.Get("/logout", ui.Logout())
	r.With(lsmiddleware.ApiAuth(*config, true), lsmiddleware.RedirectLoggedIn).Get("/register", ui.Register(config))

	handlers.FileServer(r, "/static", config)

	// start daemon for updating mac addresses
	go macDeamon()

	http.ListenAndServe(config.Address(), r)
}
