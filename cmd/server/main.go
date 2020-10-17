package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

type Config struct {
	Host string
	Port string
}

func (c Config) Address() string {
	return fmt.Sprintf("%s:%s", c.Host, c.Port)
}

const (
	hostEnv = "LS_HOST"
	portEnv = "LS_PORT"
)

func EnvConfig() *Config {
	return &Config{
		Host: DefaultEnv(hostEnv, "127.0.0.1"),
		Port: DefaultEnv(portEnv, "3000"),
	}
}

func DefaultEnv(key, fallback string) string {
	res := os.Getenv(key)
	if res == "" {
		return fallback
	}
	return res
}

func main() {
	config := EnvConfig()

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello HS!"))
	})

	http.ListenAndServe(config.Address(), r)
}
