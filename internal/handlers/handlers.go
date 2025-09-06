package handlers

import (
	"context"
	"log/slog"
	"net/http"
	"strings"

	"github.com/a-h/templ"
	"github.com/faebut/prozessor_gostack/internal/config"
	"github.com/faebut/prozessor_gostack/internal/driver"
	"github.com/faebut/prozessor_gostack/internal/repository"
	"github.com/justinas/nosurf"
)

// Repo the repository used by the handlers
var Repo *Repository

// Repository is the repository type
type Repository struct {
	App *config.AppConfig
	DB  repository.DatabaseRepo
}

// NewRepo creates a new repository
func NewRepo(a *config.AppConfig, db *driver.DB) *Repository {
	return &Repository{
		App: a,
		DB:  repository.NewPostgresRepo(db.SQL, a),
	}
}

// NewHandlers sets the repository for the handlers
func NewHandlers(r *Repository) {
	Repo = r
}

// Make creates a http.HandlerFunction and deals with errors
func Make(h func(http.ResponseWriter, *http.Request) error) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := h(w, r); err != nil {
			// Get visitor ID from session if available
			visitorID := ""
			if id := Repo.App.Session.Get(r.Context(), "visitor_id"); id != nil {
				visitorID = id.(string)
			}

			// Clean error message by removing "ERROR:" prefix
			errMsg, _ := strings.CutPrefix(err.Error(), "ERROR: ")

			slog.Error(errMsg,
				"path", r.URL.Path,
				"visitor_id", visitorID,
			)
		}
	}
}

// TODO rename to render, as soon als render package is not needed anymore
// render helps rendering the templ and adds default data
func rendering(w http.ResponseWriter, r *http.Request, component templ.Component) error {
	// set nosurf Token in new context with value
	aCtx := context.WithValue(r.Context(), "permission", Repo.App.Session.Get(r.Context(), "permission"))
	ctx := context.WithValue(aCtx, "csrf.Token", nosurf.Token(r))
	return component.Render(ctx, w)
}

// hxRedirect makes a client side redirect through HTMX
// to override targets to full page
func hxRedirect(w http.ResponseWriter, r *http.Request, to string) error {
	if len(r.Header.Get("HX-Request")) > 0 {
		w.Header().Set("HX-Redirect", to)
		w.WriteHeader(http.StatusSeeOther)
		return nil
	}
	http.Redirect(w, r, to, http.StatusSeeOther)
	return nil
}

// Home is the home page handler simply redirect to /admin, will be /login if not logged in
func (m *Repository) Home(w http.ResponseWriter, r *http.Request) {
	// redirect
	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}
