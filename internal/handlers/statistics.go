package handlers

import (
	"log"
	"net/http"
	"time"

	"github.com/faebut/prozessor_gostack/internal/chart"
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

	startDate, _ := time.Parse("2006-01-02", "2024-01-31")
	endDate, _ := time.Parse("2006-01-02", "2024-06-30")

	visitCounts, err := m.DB.GetStatisticsVisitsByType(startDate, endDate)

	if err != nil {
		log.Println(err)
		return err
	}

	bar := chart.Bar(visitCounts, true)

	return rendering(w, r, statistics.Chart(bar, "Besuche pro Tag"))
}

// HandleVisitorChart is the visitor chart handler
func (m *Repository) HandleWeekdayChart(w http.ResponseWriter, r *http.Request) error {

	startDate, _ := time.Parse("2006-01-02", "2024-01-31")
	endDate, _ := time.Parse("2006-01-02", "2024-06-30")

	visitCounts, err := m.DB.GetStatisticsAverageWeekday(startDate, endDate)

	if err != nil {
		log.Println(err)
		return err
	}

	bar := chart.Bar(visitCounts, false)

	return rendering(w, r, statistics.Chart(bar, "Durchschnitt pro Wochentag"))
}
