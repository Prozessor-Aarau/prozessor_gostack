package main

import (
	"log"
	"net/http"

	"github.com/faebut/prozessor_gostack/internal/helpers"
	"github.com/justinas/nosurf"
)

// NoSurf adds CSRF protection to all POST requests
func NoSurf(next http.Handler) http.Handler {
	csrfHandler := nosurf.New(next)

	csrfHandler.SetBaseCookie(http.Cookie{
		HttpOnly: true,
		Path:     "/",
		Secure:   app.InProduction,
		SameSite: http.SameSiteLaxMode,
	})

	return csrfHandler
}

// SessionLoad loads and saves the session on every request
func SessionLoad(next http.Handler) http.Handler {
	return session.LoadAndSave(next)
}

// Auth checks if someone is logged in
func Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !helpers.IsAuthenticated(r) {
			session.Put(r.Context(), "error", "Log in first")
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// Level functions check if someone has the appropriate level
func Level2(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		accessLevel := session.Get(r.Context(), "access").(int)
		required := 2
		if accessLevel < required {
			session.Put(r.Context(), "error", "no access to this page")
			log.Printf(
				"Your level is not enough. Should be %d, but is %d",
				required,
				accessLevel,
			)
			http.Redirect(w, r, "/admin", http.StatusSeeOther)
			return
		}
		next.ServeHTTP(w, r)
	})
}
func Level3(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		accessLevel := session.Get(r.Context(), "access").(int)
		required := 3
		if accessLevel < required {
			session.Put(r.Context(), "error", "no access to this page")
			log.Printf(
				"Your level is not enough. Should be %d, but is %d",
				required,
				accessLevel,
			)
			http.Redirect(w, r, "/admin", http.StatusSeeOther)
			return
		}
		next.ServeHTTP(w, r)
	})
}
func Level4(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		accessLevel := session.Get(r.Context(), "access").(int)
		required := 4

		if accessLevel < required {
			session.Put(r.Context(), "error", "no access to this page")
			log.Printf(
				"Your level is not enough. Should be %d, but is %d",
				required,
				accessLevel,
			)
			http.Redirect(w, r, "/admin", http.StatusSeeOther)
			return
		}
		next.ServeHTTP(w, r)
	})
}
