package ui

import (
	"html/template"
	"net/http"
	"strings"
	"time"

	"github.com/hakierspejs/long-season/pkg/models"
	"github.com/hakierspejs/long-season/pkg/services/handlers"
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

func Logout() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		now := time.Now()
		http.SetCookie(w, &http.Cookie{
			Name:     "jwt-token",
			Expires:  now.Add(time.Hour * 4),
			Value:    "",
			HttpOnly: true,
			Path:     "/",
		})
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
	}
}

func Devices(config models.Config, opener handlers.Opener) http.HandlerFunc {
	tmpl := template.Must(renderTemplate(opener, "tmpl/devices.html"))
	return func(w http.ResponseWriter, r *http.Request) {
		tmpl.ExecuteTemplate(w, "layout", newData(r, config))
	}
}
