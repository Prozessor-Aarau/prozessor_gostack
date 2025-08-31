package handlers

import (
	"log"
	"net/http"

	"github.com/faebut/prozessor_gostack/internal/models"
	"github.com/faebut/prozessor_gostack/view/statistics"
)

// HandleStatisticsIndex is the statistics page handler
func (m *Repository) HandleStatisticsIndex(w http.ResponseWriter, r *http.Request) error {
	// send the data to the template
	visitor, ok := m.App.Session.Get(r.Context(), "visitor").(models.Visitor)
	if !ok {
		log.Println("could not get visitor from Session")
		m.App.Session.Put(r.Context(), "error", "Can't get visitor from session")

		return hxRedirect(w, r, "/admin/logout")
	}

	return rendering(w, r, statistics.Statistics(visitor))
}

// HandleVisitorChart is the visitor chart handler
func (m *Repository) HandleVisitorChart(w http.ResponseWriter, r *http.Request) error {

	visitCounts, err := m.DB.GetStatisticsVisitsByRange()

	if err != nil {
		log.Println(err)
		return err
	}

	log.Println(visitCounts)

	return rendering(w, r, statistics.Card())
}
