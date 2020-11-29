package ui

import (
	"html/template"
	"net/http"
	"strings"
	"time"

	"github.com/hakierspejs/long-season/pkg/models"
	"github.com/hakierspejs/long-season/pkg/services/config"
)

func renderWithOpener(path string, readFunc config.Opener) (*template.Template, error) {
	str := new(strings.Builder)

	b, err := readFunc("web/tmpl/layout.html")
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

func renderTemplate(c *models.Config, path string) (*template.Template, error) {
	return renderWithOpener(path, config.MakeOpener(c))
}

func Home(c *models.Config) http.HandlerFunc {
	tmpl := template.Must(renderTemplate(c, "web/tmpl/home.html"))
	return func(w http.ResponseWriter, r *http.Request) {
		tmpl.ExecuteTemplate(w, "layout", nil)
	}
}

func LoginPage(c *models.Config) http.HandlerFunc {
	tmpl := template.Must(renderTemplate(c, "web/tmpl/login.html"))
	return func(w http.ResponseWriter, r *http.Request) {
		tmpl.ExecuteTemplate(w, "layout", nil)
	}
}

func Register(c *models.Config) http.HandlerFunc {
	tmpl := template.Must(renderTemplate(c, "web/tmpl/register.html"))
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

func Devices(c *models.Config) http.HandlerFunc {
	tmpl := template.Must(renderTemplate(c, "web/tmpl/devices.html"))
	return func(w http.ResponseWriter, r *http.Request) {
		tmpl.ExecuteTemplate(w, "layout", nil)
	}
}
