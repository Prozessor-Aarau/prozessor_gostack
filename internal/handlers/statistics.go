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

// HandleStatisticsFilter handles the statistics filtering
func (m *Repository) HandleStatisticsFilter(w http.ResponseWriter, r *http.Request) error {

	minString := r.Form.Get("min_date")
	maxString := r.Form.Get("max_date")

	minDate, err := time.Parse("2006-01-02", minString)
	maxDate, err := time.Parse("2006-01-02", maxString)

	if err != nil {
		return err
	}

	filter := statistics.Filter{
		MinDate: minDate,
		MaxDate: maxDate,
	}

	return rendering(w, r, statistics.CardStack(filter))
}

// HandleVisitorChart is the visitor chart handler
func (m *Repository) HandleVisitorChart(w http.ResponseWriter, r *http.Request) error {

	var darkMode bool

	startString := r.URL.Query().Get("min")
	endString := r.URL.Query().Get("max")
	mode := r.URL.Query().Get("mode")

	startDate, err := time.Parse("2006-01-02", startString)
	endDate, err := time.Parse("2006-01-02", endString)
	if err != nil {
		return err
	}

	if mode == "dark" {
		darkMode = true
	} else {
		darkMode = false
	}

	visitCounts, err := m.DB.GetStatisticsVisitsByType(startDate, endDate)

	if err != nil {
		return err
	}

	bar := chart.Bar(visitCounts, true, darkMode)

	return rendering(w, r, statistics.Chart(bar, "Besuche pro Tag"))
}

// HandleWeekdayChart is the weekday chart handler
func (m *Repository) HandleWeekdayChart(w http.ResponseWriter, r *http.Request) error {

	var darkMode bool

	startString := r.URL.Query().Get("min")
	endString := r.URL.Query().Get("max")
	mode := r.URL.Query().Get("mode")

	startDate, err := time.Parse("2006-01-02", startString)
	endDate, err := time.Parse("2006-01-02", endString)
	if err != nil {
		return err
	}

	if mode == "dark" {
		darkMode = true
	} else {
		darkMode = false
	}

	visitCounts, err := m.DB.GetStatisticsAverageWeekday(startDate, endDate)

	if err != nil {
		return err
	}

	bar := chart.Bar(visitCounts, true, darkMode)

	return rendering(w, r, statistics.Chart(bar, "Durchschnitt pro Wochentag"))
}

// HandleAtelierChart is the atelier chart handler
func (m *Repository) HandleAtelierChart(w http.ResponseWriter, r *http.Request) error {

	var darkMode bool

	startString := r.URL.Query().Get("min")
	endString := r.URL.Query().Get("max")
	mode := r.URL.Query().Get("mode")

	startDate, err := time.Parse("2006-01-02", startString)
	endDate, err := time.Parse("2006-01-02", endString)
	if err != nil {
		return err
	}

	if mode == "dark" {
		darkMode = true
	} else {
		darkMode = false
	}

	atelierCounts, err := m.DB.GetStatisticsAteliers(startDate, endDate)

	if err != nil {
		return err
	}

	pie := chart.Pie(atelierCounts, darkMode)

	return rendering(w, r, statistics.PieChart(pie, "Nutzung Ateliers"))
}

// HandleAgesChart is the ages chart handler
func (m *Repository) HandleAgesChart(w http.ResponseWriter, r *http.Request) error {

	var darkMode bool

	startString := r.URL.Query().Get("min")
	endString := r.URL.Query().Get("max")
	mode := r.URL.Query().Get("mode")

	startDate, err := time.Parse("2006-01-02", startString)
	endDate, err := time.Parse("2006-01-02", endString)
	if err != nil {
		return err
	}

	if mode == "dark" {
		darkMode = true
	} else {
		darkMode = false
	}

	agesCounts, err := m.DB.GetStatisticsAges(startDate, endDate)

	if err != nil {
		return err
	}

	pie := chart.Pie(agesCounts, darkMode)

	return rendering(w, r, statistics.PieChart(pie, "Nutzung nach Altersgruppe"))
}

// HandleKeyFigures is the key figures handler
func (m *Repository) HandleKeyFigures(w http.ResponseWriter, r *http.Request) error {

	startString := r.URL.Query().Get("min")
	endString := r.URL.Query().Get("max")

	startDate, err := time.Parse("2006-01-02", startString)
	endDate, err := time.Parse("2006-01-02", endString)
	if err != nil {
		return err
	}

	figures, err := m.DB.GetStatisticsKeyFigures(startDate, endDate)

	if err != nil {
		return err
	}

	return rendering(w, r, statistics.FigureCard(figures, "Kennzahlen"))
}
