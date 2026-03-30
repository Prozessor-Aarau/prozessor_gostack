package handlers

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/faebut/prozessor_gostack/internal/models"
	"github.com/faebut/prozessor_gostack/view/admin"
)

// HandleAdminIndex is the admin page handler
func (m *Repository) HandleAdminIndex(w http.ResponseWriter, r *http.Request) error {
	// send the data to the template
	visitor, ok := m.App.Session.Get(r.Context(), "visitor").(models.Visitor)
	if !ok {
		slog.Error("could not get visitor from session", "path", r.URL.Path)
		m.App.Session.Put(r.Context(), "error", "Can't get visitor from session")

		return hxRedirect(w, r, "/admin/logout")
	}

	ateliers, err := m.DB.GetAllAteliers()
	if err != nil {
		return fmt.Errorf("retrieving ateliers: %w", err)
	}

	return rendering(w, r, admin.Admin(visitor, ateliers))
}
