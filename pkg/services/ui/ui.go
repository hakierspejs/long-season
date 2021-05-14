package ui

import (
	"html/template"
	"net/http"
	"strings"
	"time"

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

func renderTemplate(opener handlers.Opener, path string) (*template.Template, error) {
	return renderWithOpener(path, opener)
}

func Home(opener handlers.Opener) http.HandlerFunc {
	tmpl := template.Must(renderTemplate(opener, "tmpl/home.html"))
	return func(w http.ResponseWriter, r *http.Request) {
		tmpl.ExecuteTemplate(w, "layout", nil)
	}
}

func LoginPage(opener handlers.Opener) http.HandlerFunc {
	tmpl := template.Must(renderTemplate(opener, "tmpl/login.html"))
	return func(w http.ResponseWriter, r *http.Request) {
		tmpl.ExecuteTemplate(w, "layout", nil)
	}
}

func Register(opener handlers.Opener) http.HandlerFunc {
	tmpl := template.Must(renderTemplate(opener, "tmpl/register.html"))
	return func(w http.ResponseWriter, r *http.Request) {
		tmpl.ExecuteTemplate(w, "layout", nil)
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

func Devices(opener handlers.Opener) http.HandlerFunc {
	tmpl := template.Must(renderTemplate(opener, "tmpl/devices.html"))
	return func(w http.ResponseWriter, r *http.Request) {
		tmpl.ExecuteTemplate(w, "layout", nil)
	}
}
