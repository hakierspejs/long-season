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
	"github.com/hakierspejs/long-season/pkg/services/session"
	"github.com/hakierspejs/long-season/pkg/services/toussaint"
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
	Opener         handlers.Opener
	Users          storage.Users
	Devices        storage.Devices
	StatusTx       storage.StatusTx
	TwoFactor      storage.TwoFactor
	MacsChan       chan<- []net.HardwareAddr
	PublicCors     Cors
	Adapter        *happier.Adapter
	SessionRenewer session.Renewer
	SessionSaver   session.Saver
	SessionKiller  session.Killer
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

	sessionGuard := session.Guard(args.SessionRenewer)
	twoFactorGuard := toussaint.Guard(args.TwoFactor, args.SessionRenewer)

	guard := func(next http.Handler) http.Handler {
		res := next
		res = sessionGuard(res)
		res = twoFactorGuard(res)
		return res
	}

	twoFactorCleaner := toussaint.Cleaner(args.SessionRenewer, args.SessionKiller, "/")
	twoFactorOnly := toussaint.TwoFactorOnly(args.SessionRenewer, "/")

	// UI routes.
	r.With(twoFactorCleaner).Get("/", ui.Home(config, args.Opener))

	r.With(
		twoFactorCleaner, lsmiddleware.RedirectLoggedIn(args.SessionRenewer),
	).Get("/login", ui.LoginPage(config, args.Opener))
	r.Post("/login", args.Adapter.WithError(ui.Auth(ui.AuthArguments{
		Saver:                args.SessionSaver,
		Users:                args.Users,
		TwoFactor:            args.TwoFactor,
		TwoFactorRedirectURI: "/twofactor",
	})))

	r.With(guard, twoFactorCleaner).Get("/who", handlers.Who(args.SessionRenewer))

	r.With(guard, twoFactorCleaner).Get("/devices", ui.Devices(config, args.Opener))

	r.With(guard, twoFactorCleaner).Get("/account", ui.Account(config, args.Opener))

	r.Get("/logout", ui.Logout(args.SessionKiller))

	r.With(
		twoFactorCleaner, lsmiddleware.RedirectLoggedIn(args.SessionRenewer),
	).Get("/register", ui.Register(config, args.Opener))

	r.With(sessionGuard, twoFactorOnly).Get("/twofactor", ui.TwoFactor(config, args.Opener))
	r.With(sessionGuard, twoFactorOnly).Post(
		"/twofactor/codes",
		args.Adapter.WithError(ui.AuthWithCodes(ui.AuthWithCodesArguments{
			Renewer:   args.SessionRenewer,
			Saver:     args.SessionSaver,
			TwoFactor: args.TwoFactor,
		})),
	)

	// API routes.
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
				r.Get("/", args.Adapter.WithError(api.UserRead(args.SessionRenewer, args.Users)))

				// Users can only delete themselves.
				r.With(
					lsmiddleware.Private(args.SessionRenewer),
				).Delete("/", args.Adapter.WithError(api.UserRemove(args.Users)))

				r.With(
					guard, lsmiddleware.Private(args.SessionRenewer),
				).Patch("/", args.Adapter.WithError(api.UserUpdate(args.Users)))

				r.With(
					guard, lsmiddleware.Private(args.SessionRenewer),
				).Put("/password", args.Adapter.WithError(api.UpdateUserPassword(args.Users)))

				r.With(
					guard, lsmiddleware.Private(args.SessionRenewer),
				).Route("/twofactor", func(r chi.Router) {
					r.Get("/", args.Adapter.WithError(api.TwoFactorMethods(args.TwoFactor)))
					r.Get("/{twofactor-id}", args.Adapter.WithError(api.TwoFactorMethod(args.TwoFactor)))
					r.Delete("/{twofactor-id}", args.Adapter.WithError(api.TwoFactorMethodRemove(args.TwoFactor)))
					r.Post("/otp", args.Adapter.WithError(api.AddOTP(args.SessionRenewer, args.TwoFactor)))
					r.Post("/recovery", args.Adapter.WithError(api.AddRecovery(args.SessionRenewer, args.TwoFactor)))
				})

				r.With(
					guard, lsmiddleware.Private(args.SessionRenewer),
				).Route("/devices", func(r chi.Router) {
					r.Get("/", args.Adapter.WithError(api.UserDevices(args.Devices)))
					r.Post("/", args.Adapter.WithError(api.DeviceAdd(args.SessionRenewer, args.Devices)))

					r.With(lsmiddleware.DeviceID).Route("/{device-id}", func(r chi.Router) {
						r.Get("/", args.Adapter.WithError(api.DeviceRead(args.SessionRenewer, args.Devices)))
						r.Delete("/", args.Adapter.WithError(api.DeviceRemove(args.SessionRenewer, args.Devices)))
					})
				})
			})
		})
		r.With(lsmiddleware.UpdateAuth(&config)).Put(
			"/update",
			args.Adapter.WithError(api.UpdateStatus(args.MacsChan)),
		)
		r.Get("/status", args.Adapter.WithError(api.Status(args.StatusTx)))

		r.With(guard).Route("/twofactor", func(r chi.Router) {
			r.Get("/otp/options", args.Adapter.WithError(api.OptionsOTP(config, args.SessionRenewer)))
		})
	})

	handlers.FileServer(r, "/static", args.Opener)

	return r
}
