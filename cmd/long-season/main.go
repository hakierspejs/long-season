package main

import (
	"context"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/cors"

	"github.com/hakierspejs/long-season/pkg/services/config"
	"github.com/hakierspejs/long-season/pkg/services/handlers"
	"github.com/hakierspejs/long-season/pkg/services/happier"
	"github.com/hakierspejs/long-season/pkg/services/jojo"
	"github.com/hakierspejs/long-season/pkg/services/router"
	"github.com/hakierspejs/long-season/pkg/services/session"
	"github.com/hakierspejs/long-season/pkg/services/status"
	"github.com/hakierspejs/long-season/pkg/storage"
	"github.com/hakierspejs/long-season/pkg/storage/abstract"
	"github.com/hakierspejs/long-season/pkg/storage/temp"
	"github.com/hakierspejs/long-season/web"
)

func main() {
	config := config.Env()

	factoryStorage, closer, err := abstract.Factory(config.DatabasePath, config.DatabaseType)
	if err != nil {
		log.Fatal(err.Error())
	}
	defer closer()

	onlineUsersStorage := temp.NewOnlineUsers()
	statusTx := temp.NewStatusTx()

	ctx := context.Background()
	macChannel, macDeamon := status.NewDaemon(ctx, status.DaemonArgs{
		OnlineUsers:   onlineUsersStorage,
		Devices:       factoryStorage.Devices(),
		Counters:      statusTx,
		RefreshTime:   config.RefreshTime,
		SingleAddrTTL: config.SingleAddrTTL,
	})

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

	var opener handlers.Opener
	if config.Debug {
		opener = func(path string) ([]byte, error) {
			return ioutil.ReadFile("web/" + path)
		}
	} else {
		opener = web.Open
	}

	jwtSession := &jojo.JWT{
		Secret:    []byte(config.JWTSecret),
		AppName:   config.AppName,
		CookieKey: "jwt-token",
	}

	userAdapter := storage.UserAdapter{
		OnlineUsersStorage: onlineUsersStorage,
	}

	r := router.NewRouter(*config, router.Args{
		Opener:      opener,
		Users:       factoryStorage.Users(),
		Devices:     factoryStorage.Devices(),
		StatusTx:    statusTx,
		TwoFactor:   factoryStorage.TwoFactor(),
		OnlineUsers: onlineUsersStorage,
		UserAdapter: userAdapter,
		MacsChan:    macChannel,
		PublicCors:  publicCors,
		Adapter:     happier.NewAdapter(),
		SessionRenewer: session.RenewerComposite(
			jwtSession.RenewFromHeaderToken("Authorization", "Bearer"),
			jwtSession.RenewFromCookies(),
		),
		SessionSaver:  jwtSession,
		SessionKiller: jwtSession,
	})

	// start daemon for updating mac addresses
	go macDeamon()

	http.ListenAndServe(config.Address(), r)
}
