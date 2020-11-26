package ui

import (
	"html/template"
	"net/http"
	"strings"
	"time"

	"github.com/hakierspejs/long-season/pkg/static"
)

func embeddedTemplate(path string) (*template.Template, error) {
	str := new(strings.Builder)

	b, err := static.Open("web/tmpl/layout.html")
	if err != nil {
		return nil, err
	}
	_, err = str.Write(b)
	if err != nil {
		return nil, err
	}

	b, err = static.Open(path)
	if err != nil {
		return nil, err
	}
	_, err = str.Write(b)
	if err != nil {
		return nil, err
	}

	return template.New("ui").Parse(str.String())
}

func Home() http.HandlerFunc {
	tmpl := template.Must(embeddedTemplate("web/tmpl/home.html"))
	return func(w http.ResponseWriter, r *http.Request) {
		tmpl.ExecuteTemplate(w, "layout", nil)
	}
}

func LoginPage() http.HandlerFunc {
	tmpl := template.Must(embeddedTemplate("web/tmpl/login.html"))
	return func(w http.ResponseWriter, r *http.Request) {
		tmpl.ExecuteTemplate(w, "layout", nil)
	}
}

func Register() http.HandlerFunc {
	tmpl := template.Must(embeddedTemplate("web/tmpl/register.html"))
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

func Devices() http.HandlerFunc {
	tmpl := template.Must(embeddedTemplate("web/tmpl/devices.html"))
	return func(w http.ResponseWriter, r *http.Request) {
		tmpl.ExecuteTemplate(w, "layout", nil)
	}
}
