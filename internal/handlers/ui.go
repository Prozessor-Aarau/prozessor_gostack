package handlers

import (
	"log"
	"net/http"

	"github.com/faebut/prozessor_gostack/internal/models"
	"github.com/faebut/prozessor_gostack/view/ui"
	"github.com/go-chi/chi/v5"
)

// VisitorsSelectBox returns a select box with the filtered and selected visitors
func (m *Repository) HandleVisitorsSelectBox(w http.ResponseWriter, r *http.Request) error {
	visID := chi.URLParam(r, "id")

	searchTerm := r.Form.Get("search_term")

	filter := models.NewVisitorSearch(searchTerm)

	// Get the queryparams to send back to the right page
	var pricing bool
	if r.URL.Query().Has("pricing") {
		pricing = true
	}
	// get filtered visitors from DB
	visitors, _, err := m.DB.AllVisitors(1, 0, filter)
	if err != nil {
		log.Println("could not get visitors from DB: ", err)
		return err
	}

	return rendering(w, r, ui.MemberSelectBox(visID, visitors, pricing))
}
