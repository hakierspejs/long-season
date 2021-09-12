/*
Package router implements http long-season server with routing logic.
*/
package router

import (
	"net"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"

	"github.com/hakierspejs/long-season/pkg/models"
	"github.com/hakierspejs/long-season/pkg/services/handlers"
	"github.com/hakierspejs/long-season/pkg/services/handlers/api/v1"
	"github.com/hakierspejs/long-season/pkg/services/happier"
	lsmiddleware "github.com/hakierspejs/long-season/pkg/services/middleware"
	"github.com/hakierspejs/long-season/pkg/services/ui"
	"github.com/hakierspejs/long-season/pkg/storage"
)

// Cors interface contains Handler for setting up CORS.
type Cors interface {
	// Handler is a middleware that applies CORS settings
	// to given handler.
	Handler(http.Handler) http.Handler
}

// Args contains dependencies for router.
type Args struct {
	Opener     handlers.Opener
	Users      storage.Users
	Devices    storage.Devices
	StatusTx   storage.StatusTx
	MacsChan   chan<- []net.HardwareAddr
	PublicCors Cors
	Adapter    *happier.Adapter
}

// NewRouter returns Handler, which contains all the handlers and
// routing logic for long-season http server.
func NewRouter(config models.Config, args Args) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(middleware.NoCache)
	r.Use(lsmiddleware.Debug(config))

	r.Get("/", ui.Home(config, args.Opener))
	r.With(
		lsmiddleware.ApiAuth(config, true),
		lsmiddleware.RedirectLoggedIn,
	).Get("/login", ui.LoginPage(config, args.Opener))

	r.Route("/api/v1", func(r chi.Router) {
		r.Route("/users", func(r chi.Router) {
			// Use CORS to allow GET/OPTIONS to GET /api/v1/users from
			// anywhere.
			// This technically wraps a nil handler around publicCors, but the
			// nil handler never gets called. This is weirdness stemming from
			// how go-chi/cors is supposed to be applied globally to the entire
			// application, and not to particular endpoints.
			r.With(args.PublicCors.Handler).Options("/", nil)
			r.With(args.PublicCors.Handler).Get("/", args.Adapter.WithError(api.UsersAll(args.Users)))
			r.Post("/", args.Adapter.WithError(api.UserCreate(args.Users)))

			r.With(lsmiddleware.UserID).Route("/{user-id}", func(r chi.Router) {
				r.With(
					lsmiddleware.ApiAuth(config, true),
				).Get("/", args.Adapter.WithError(api.UserRead(args.Users)))

				// Users can only delete themselves.
				r.With(
					lsmiddleware.ApiAuth(config, false),
					lsmiddleware.Private,
				).Delete("/", args.Adapter.WithError(api.UserRemove(args.Users)))

				r.With(
					lsmiddleware.ApiAuth(config, false),
					lsmiddleware.Private,
				).Patch("/", args.Adapter.WithError(api.UserUpdate(args.Users)))

				r.With(
					lsmiddleware.ApiAuth(config, false),
					lsmiddleware.Private,
				).Put("/password", args.Adapter.WithError(api.UpdateUserPassword(args.Users)))

				r.With(
					lsmiddleware.ApiAuth(config, false),
					lsmiddleware.Private,
				).Route("/devices", func(r chi.Router) {
					r.Get("/", args.Adapter.WithError(api.UserDevices(args.Devices)))
					r.Post("/", args.Adapter.WithError(api.DeviceAdd(args.Devices)))

					r.With(lsmiddleware.DeviceID).Route("/{device-id}", func(r chi.Router) {
						r.Get("/", args.Adapter.WithError(api.DeviceRead(args.Devices)))
						r.Delete("/", args.Adapter.WithError(api.DeviceRemove(args.Devices)))
						r.Patch("/", args.Adapter.WithError(api.DeviceUpdate(args.Devices)))
					})
				})
			})
		})
		r.Post("/login", args.Adapter.WithError(api.ApiAuth(config, args.Users)))
		r.With(lsmiddleware.UpdateAuth(&config)).Put(
			"/update",
			args.Adapter.WithError(api.UpdateStatus(args.MacsChan)),
		)
		r.Get("/status", args.Adapter.WithError(api.Status(args.StatusTx)))
	})

	r.With(lsmiddleware.ApiAuth(config, false)).Get("/who", handlers.Who())
	r.With(lsmiddleware.ApiAuth(config, false)).Get("/devices", ui.Devices(config, args.Opener))
	r.With(lsmiddleware.ApiAuth(config, false)).Get("/account", ui.Account(config, args.Opener))
	r.Get("/logout", ui.Logout())
	r.With(
		lsmiddleware.ApiAuth(config, true),
		lsmiddleware.RedirectLoggedIn,
	).Get("/register", ui.Register(config, args.Opener))

	handlers.FileServer(r, "/static", args.Opener)

	return r
}
