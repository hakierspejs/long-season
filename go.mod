module github.com/hakierspejs/long-season

// +heroku goVersion go1.16
// +heroku install ./cmd/...

go 1.16

require (
	github.com/alioygur/gores v1.2.1
	github.com/cristalhq/jwt/v3 v3.0.4
	github.com/go-chi/chi v4.1.2+incompatible
	github.com/go-chi/cors v1.1.1
	github.com/google/uuid v1.1.2
	github.com/matryer/is v1.4.0
	github.com/pquerna/otp v1.3.0
	github.com/thinkofher/horror v0.1.2
	github.com/urfave/cli/v2 v2.3.0
	go.etcd.io/bbolt v1.3.5
	golang.org/x/crypto v0.0.0-20201016220609-9e8e0b390897
)
