package handlers

import (
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

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
	var member bool
	var logistics bool
	if r.URL.Query().Has("pricing") {
		pricing = true
	}
	if r.URL.Query().Has("member") {
		member = true
	}
	if r.URL.Query().Has("logistics") {
		logistics = true
	}
	// get filtered visitors from DB
	visitors, _, err := m.DB.AllVisitors(1, 0, filter)
	if err != nil {
		return fmt.Errorf("failed to get visitors: %w", err)
	}

	return rendering(w, r, ui.MemberSelectBox(visID, visitors, pricing, member, logistics))
}

// HandleSelectMembership returns the membership from the select box
func (m *Repository) HandleSelectMembership(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	visitor, err := m.DB.VisitorByID(id)
	if err != nil {
		slog.Error(fmt.Errorf("failed to get visitor: %w", err).Error())
	}

	returnString := fmt.Sprintf("<input type=\"hidden\" id=\"membership\" name=\"membership\" value=\"%s\" />", strconv.FormatBool(visitor.Ressort.ID > 0))

	fmt.Fprint(w, returnString)
}
