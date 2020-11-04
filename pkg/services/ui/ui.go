package ui

import (
	"net/http"
	"text/template"
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
