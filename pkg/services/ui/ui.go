package ui

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"strings"
	"time"

	"github.com/hakierspejs/long-season/pkg/models"
	"github.com/hakierspejs/long-season/pkg/services/handlers"
	"github.com/hakierspejs/long-season/pkg/services/happier"
	"github.com/hakierspejs/long-season/pkg/services/session"
	"github.com/hakierspejs/long-season/pkg/services/users"
	"github.com/hakierspejs/long-season/pkg/storage"
	"github.com/thinkofher/horror"
)

func renderWithOpener(path string, readFunc handlers.Opener) (*template.Template, error) {
	str := new(strings.Builder)

	b, err := readFunc("tmpl/layout.html")
	if err != nil {
		return nil, err
	}
	_, err = str.Write(b)
	if err != nil {
		return nil, err
	}

	b, err = readFunc(path)
	if err != nil {
		return nil, err
	}
	_, err = str.Write(b)
	if err != nil {
		return nil, err
	}

	return template.New("ui").Parse(str.String())
}

type layout struct {
	PrideMonth bool
	City       string
	Space      string
}

type data struct {
	Layout layout
}

func newData(r *http.Request, c models.Config) *data {
	now := time.Now()
	return &data{
		Layout: layout{
			PrideMonth: now.Month() == time.June,
			City:       c.City,
			Space:      c.Space,
		},
	}
}

func renderTemplate(opener handlers.Opener, path string) (*template.Template, error) {
	return renderWithOpener(path, opener)
}

func Home(config models.Config, opener handlers.Opener) http.HandlerFunc {
	tmpl := template.Must(renderTemplate(opener, "tmpl/home.html"))
	return func(w http.ResponseWriter, r *http.Request) {
		tmpl.ExecuteTemplate(w, "layout", newData(r, config))
	}
}

func LoginPage(config models.Config, opener handlers.Opener) http.HandlerFunc {
	tmpl := template.Must(renderTemplate(opener, "tmpl/login.html"))
	return func(w http.ResponseWriter, r *http.Request) {
		tmpl.ExecuteTemplate(w, "layout", newData(r, config))
	}
}

func Register(config models.Config, opener handlers.Opener) http.HandlerFunc {
	tmpl := template.Must(renderTemplate(opener, "tmpl/register.html"))
	return func(w http.ResponseWriter, r *http.Request) {
		tmpl.ExecuteTemplate(w, "layout", newData(r, config))
	}
}

func Logout(killer session.Killer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_ = killer.Kill(r.Context(), w)
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
	}
}

func Devices(config models.Config, opener handlers.Opener) http.HandlerFunc {
	tmpl := template.Must(renderTemplate(opener, "tmpl/devices.html"))
	return func(w http.ResponseWriter, r *http.Request) {
		tmpl.ExecuteTemplate(w, "layout", newData(r, config))
	}
}

func Account(config models.Config, opener handlers.Opener) http.HandlerFunc {
	tmpl := template.Must(renderTemplate(opener, "tmpl/account.html"))
	return func(w http.ResponseWriter, r *http.Request) {
		tmpl.ExecuteTemplate(w, "layout", newData(r, config))
	}
}

func Auth(saver session.Saver, db storage.Users) horror.HandlerFunc {
	type payload struct {
		Nickname string `json:"nickname"`
		Password string `json:"password"`
	}
	return func(w http.ResponseWriter, r *http.Request) error {
		ctx := r.Context()
		errFactory := happier.FromRequest(r)

		input := new(payload)
		err := json.NewDecoder(r.Body).Decode(input)
		if err != nil {
			return errFactory.BadRequest(
				fmt.Errorf("json.NewDecoder().Decode: %w", err),
				fmt.Sprintf("Invalid input: %s.", err.Error()),
			)
		}

		match, err := users.AuthenticateWithPassword(ctx, users.AuthenticationDependencies{
			Request: users.AuthenticationRequest{
				Nickname: input.Nickname,
				Password: []byte(input.Password),
			},
			Storage:      db,
			ErrorFactory: errFactory,
		})
		if err != nil {
			return fmt.Errorf("users.AuthenticateWithPassword: %w", err)
		}

		newSession := session.New(ctx, session.Builder{
			UserID:   match.ID,
			Nickname: match.Nickname,
		})

		if err := saver.Save(ctx, w, *newSession); err != nil {
			return fmt.Errorf("saver.Save: %w", err)
		}

		w.WriteHeader(http.StatusOK)
		return nil
	}
}
