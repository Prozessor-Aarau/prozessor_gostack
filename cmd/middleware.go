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

	// do exempt newsletter subscribtion
	csrfHandler.ExemptPath("/subscribe")

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

// CheckPermission checks if someone has at least one of the provided permissions
func CheckPermission(permissions ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			redirect := func() {
				to := "/admin"
				log.Println("No permission for this route")
				if len(r.Header.Get("HX-Request")) > 0 {
					w.Header().Set("HX-Redirect", to)
					w.WriteHeader(http.StatusSeeOther)
					return
				}
				http.Redirect(w, r, to, http.StatusSeeOther)
			}

			role := session.Get(r.Context(), "permission").(models.Role)

			for _, permission := range permissions {
				if hasPermission(role, permission) {
					next.ServeHTTP(w, r)
					return
				}
			}

			redirect()
		})
	}
}

func hasPermission(p models.Role, permission string) bool {
	switch permission {
	case "ViewVisitors":
		return p.ViewVisitors
	case "EditVisitor":
		return p.EditVisitor
	case "EditVisitorMembership":
		return p.EditVisitorMembership
	case "EditVisitorRole":
		return p.EditVisitorRole
	case "ViewMachines":
		return p.ViewMachines
	case "ViewBadges":
		return p.ViewBadges
	case "AssignBadge":
		return p.AssignBadge
	case "EditBadge":
		return p.EditBadge
	case "ViewAbonnements":
		return p.ViewAbonnements
	case "AddAbonnement":
		return p.AddAbonnement
	case "EditAbonnement":
		return p.EditAbonnement
	case "ViewReminders":
		return p.ViewReminders
	case "AddReminder":
		return p.AddReminder
	case "ViewNewsletters":
		return p.ViewNewsletters
	case "AddNewsletter":
		return p.AddNewsletter
	case "ViewLogistics":
		return p.ViewLogistics
	case "ViewStatistics":
		return p.ViewStatistics
	case "ViewLogs":
		return p.ViewLogs
	case "ViewFinances":
		return p.ViewFinances
	case "AddFinances":
		return p.AddFinances
	case "EditFinances":
		return p.EditFinances
	case "EditOpeningHours":
		return p.EditOpeningHours
	case "DeleteOpeningHours":
		return p.DeleteOpeningHours
	case "ViewSupervisions":
		return p.ViewSupervisions
	case "EditOwnSupervisions":
		return p.EditOwnSupervisions
	case "EditAllSupervisions":
		return p.EditAllSupervisions
	}
	return false
}
