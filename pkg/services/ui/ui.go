package ui

import (
	"net/http"
	"text/template"
	"time"
)

func Home() http.HandlerFunc {
	tmpl := template.Must(template.ParseFiles("tmpl/layout.html", "tmpl/home.html"))
	return func(w http.ResponseWriter, r *http.Request) {
		tmpl.ExecuteTemplate(w, "layout", nil)
	}
}

func LoginPage() http.HandlerFunc {
	tmpl := template.Must(template.ParseFiles("tmpl/layout.html", "tmpl/login.html"))
	return func(w http.ResponseWriter, r *http.Request) {
		tmpl.ExecuteTemplate(w, "layout", nil)
	}
}

func Register() http.HandlerFunc {
	tmpl := template.Must(template.ParseFiles("tmpl/layout.html", "tmpl/register.html"))
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
	tmpl := template.Must(template.ParseFiles("tmpl/layout.html", "tmpl/devices.html"))
	return func(w http.ResponseWriter, r *http.Request) {
		tmpl.ExecuteTemplate(w, "layout", nil)
	}
}
