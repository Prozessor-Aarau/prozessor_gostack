package handlers

import (
	"context"
	"log"
	"log/slog"
	"net/http"

	"github.com/a-h/templ"
	"github.com/faebut/prozessor_gostack/internal/config"
	"github.com/faebut/prozessor_gostack/internal/driver"
	"github.com/faebut/prozessor_gostack/internal/models"
	"github.com/faebut/prozessor_gostack/internal/render"
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
			slog.Error("internal server error", "err", err, "path", r.URL.Path)
		}
	}
}

// TODO rename to render, as soon als render package is not needed anymore
// render helps rendering the templ and adds default data
func rendering(w http.ResponseWriter, r *http.Request, component templ.Component) error {
	// set nosurf Token in new context with value
	ctx := context.WithValue(r.Context(), "csrf.Token", nosurf.Token(r))
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

// Home is the home page handler
func (m *Repository) Home(w http.ResponseWriter, r *http.Request) {
	// redirect
	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}

// Admin is the admin page handler
func (m *Repository) Admin(w http.ResponseWriter, r *http.Request) {
	// send the data to the template
	user, ok := m.App.Session.Get(r.Context(), "visitor").(models.Visitor)
	if !ok {
		log.Println("could not get visitor from Session")
		m.App.Session.Put(r.Context(), "error", "Can't get visitor from session")

		http.Redirect(w, r, "/admin/logout", http.StatusSeeOther)
		return
	}

	data := make(map[string]interface{})
	data["user"] = user

	render.Template(w, r, "admin.html", &models.TemplateData{
		Data: data,
	})
}
