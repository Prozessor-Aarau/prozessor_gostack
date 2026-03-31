package handlers

import (
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/faebut/prozessor_gostack/internal/models"
	"github.com/faebut/prozessor_gostack/view/admin"
)

// HandleExpertiseUpdate handles the update of the expertises for a user
func (m *Repository) HandleExpertiseUpdate(w http.ResponseWriter, r *http.Request) error {
	visitor, ok := m.App.Session.Get(r.Context(), "visitor").(models.Visitor)
	if !ok {
		slog.Error("could not get visitor from session", "path", r.URL.Path)
		m.App.Session.Put(r.Context(), "error", "Can't get visitor from session")

		return hxRedirect(w, r, "/admin/logout")
	}

	// get checked ateliers for expertise
	checkedAteliers := make([]int, 0, len(r.PostForm))
	for a := range r.PostForm {
		kInt, err := strconv.Atoi(a)
		if err == nil {
			checkedAteliers = append(checkedAteliers, kInt)
		}
	}

	// update expertise in db
	err := m.DB.UpdateExpertise(visitor.ID, checkedAteliers)
	if err != nil {
		return fmt.Errorf("updating expertise: %w", err)
	}

	// update the visitor the session cookie
	visitor.Expertise = checkedAteliers
	m.App.Session.Put(r.Context(), "visitor", visitor)

	return rendering(w, r, admin.ExpertUpdateButton(true))
}
