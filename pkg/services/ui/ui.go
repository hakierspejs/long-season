package ui

import (
	"net/http"
	"time"

	"github.com/hakierspejs/long-season/pkg/models"
	"github.com/hakierspejs/long-season/pkg/services/single"
)

func must(err error) {
	if err != nil {
		panic(err)
	}
}

func withVendor(a single.Assets) single.Assets {
	vendorScripts := []string{
		"web/static/vendor/handlebars.min.js",
		"web/static/vendor/umbrella.min.js",
	}
	vendorStyles := []string{
		"web/static/vendor/normalize.css",
	}
	return single.Assets{
		Content: a.Content,
		Scripts: append(vendorScripts, a.Scripts...),
		Styles:  append(vendorStyles, a.Styles...),
	}
}

func page(c *models.Config, a single.Assets) http.HandlerFunc {
	p := single.New(c, withVendor(a))
	exec, err := p.Executor()
	must(err)
	return func(w http.ResponseWriter, r *http.Request) {
		exec(w)
	}
}

func Home(c *models.Config) http.HandlerFunc {
	return page(c, single.Assets{
		Content: "web/tmpl/home.html",
		Scripts: []string{
			"web/static/js/utils.js",
			"web/static/js/navbar.js",
			"web/static/js/home.js",
		},
		Styles: []string{
			"web/static/style.css",
		},
	})
}

func LoginPage(c *models.Config) http.HandlerFunc {
	return page(c, single.Assets{
		Content: "web/tmpl/login.html",
		Scripts: []string{
			"web/static/js/utils.js",
			"web/static/js/navbar.js",
			"web/static/js/login.js",
		},
		Styles: []string{
			"web/static/style.css",
		},
	})
}

func Register(c *models.Config) http.HandlerFunc {
	return page(c, single.Assets{
		Content: "web/tmpl/register.html",
		Scripts: []string{
			"web/static/js/utils.js",
			"web/static/js/navbar.js",
			"web/static/js/register.js",
		},
		Styles: []string{
			"web/static/style.css",
		},
	})
}

func Devices(c *models.Config) http.HandlerFunc {
	return page(c, single.Assets{
		Content: "web/tmpl/devices.html",
		Scripts: []string{
			"web/static/js/utils.js",
			"web/static/js/navbar.js",
			"web/static/js/devices.js",
		},
		Styles: []string{
			"web/static/style.css",
		},
	})
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
