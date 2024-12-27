package main

import (
	"log"
	"net/http"

	"github.com/faebut/prozessor_gostack/internal/helpers"
	"github.com/faebut/prozessor_gostack/internal/models"
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
			to := "/login"
			// check if htmx request to ensure full page reload
			if len(r.Header.Get("HX-Request")) > 0 {
				w.Header().Set("HX-Redirect", to)
				w.WriteHeader(http.StatusSeeOther)
				return
			}
			http.Redirect(w, r, to, http.StatusSeeOther)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// CheckPermission checks if someone has the right permission for a route
func CheckPermission(permission string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// function to run if the permission is not valid
			redirect := func() {
				log.Println("No permission for this route")
				http.Redirect(w, r, "/admin", http.StatusSeeOther)
			}

			// get permissions
			permissions := session.Get(r.Context(), "permission").(models.Role)

			switch permission {
			case "ViewVisitors":
				if !permissions.ViewVisitors {
					redirect()
					return
				}
			case "EditVisitor":
				if !permissions.EditVisitor {
					redirect()
					return
				}
			case "EditVisitorMembership":
				if !permissions.EditVisitorMembership {
					redirect()
					return
				}
			case "EditVisitorRole":
				if !permissions.EditVisitorRole {
					redirect()
					return
				}
			case "ViewMachines":
				if !permissions.ViewMachines {
					redirect()
					return
				}
			case "ViewBadges":
				if !permissions.ViewBadges {
					redirect()
					return
				}
			case "AssignBadge":
				if !permissions.AssignBadge {
					redirect()
					return
				}
			case "EditBadge":
				if !permissions.EditBadge {
					redirect()
					return
				}
			case "ViewAbonnements":
				if !permissions.EditBadge {
					redirect()
					return
				}
			case "EditAbonnement":
				if !permissions.EditBadge {
					redirect()
					return
				}
			case "ViewReminders":
				if !permissions.ViewReminders {
					redirect()
					return
				}
			case "AddReminder":
				if !permissions.AddReminder {
					redirect()
					return
				}
			case "ViewNewsletters":
				if !permissions.ViewNewsletters {
					redirect()
					return
				}
			case "AddNewsletter":
				if !permissions.AddNewsletter {
					redirect()
					return
				}
			case "ViewLogistics":
				if !permissions.ViewLogistics {
					redirect()
					return
				}
			case "ViewStatistics":
				if !permissions.ViewStatistics {
					redirect()
					return
				}
			case "ViewLogs":
				if !permissions.ViewLogs {
					redirect()
					return
				}
			case "ViewFinances":
				if !permissions.ViewFinances {
					redirect()
					return
				}
			case "AddFinances":
				if !permissions.AddFinances {
					redirect()
					return
				}
			case "EditFinances":
				if !permissions.EditFinances {
					redirect()
					return
				}
			}

			next.ServeHTTP(w, r)
		})
	}
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
