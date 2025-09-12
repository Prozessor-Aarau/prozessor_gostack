package handlers

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/faebut/prozessor_gostack/internal/chart"
	"github.com/faebut/prozessor_gostack/internal/models"
	"github.com/faebut/prozessor_gostack/view/statistics"
)

// parseDateUrl parses the request url for the min and max date
func parseDateUrl(r *http.Request) (minDate time.Time, maxDate time.Time, err error) {
	minString := r.Form.Get("min_date")
	maxString := r.Form.Get("max_date")

	minDate, err = time.Parse("2006-01-02", minString)
	maxDate, err = time.Parse("2006-01-02", maxString)

	if err != nil {
		return minDate, maxDate, fmt.Errorf("failed to parse date: %w", err)
	}

	return minDate, maxDate, nil
}

// parseDarkModeUrl parses the url to get dark or light mode
func parseDarkModeUrl(r *http.Request) (darkMode bool) {
	mode := r.URL.Query().Get("mode")

	if mode == "dark" {
		darkMode = true
	} else {
		darkMode = false
	}

	return darkMode
}

// HandleStatisticsIndex is the statistics page handler
func (m *Repository) HandleStatisticsIndex(w http.ResponseWriter, r *http.Request) error {
	// send the data to the template
	visitor, ok := m.App.Session.Get(r.Context(), "visitor").(models.Visitor)
	if !ok {
		slog.Error("could not get visitor from session", "path", r.URL.Path)
		m.App.Session.Put(r.Context(), "error", "Can't get visitor from session")

		return hxRedirect(w, r, "/admin/logout")
	}

	return rendering(w, r, statistics.Statistics(visitor))
}

// HandleStatisticsFilter handles the statistics filtering
func (m *Repository) HandleStatisticsFilter(w http.ResponseWriter, r *http.Request) error {
	// get date constraints
	minDate, maxDate, err := parseDateUrl(r)
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
	// get date constraints
	minDate, maxDate, err := parseDateUrl(r)
	if err != nil {
		return err
	}

	// get dark mode
	darkMode := parseDarkModeUrl(r)

	visitCounts, err := m.DB.GetStatisticsVisitsByType(minDate, maxDate)

	if err != nil {
		return fmt.Errorf("failed to get statistics for visits by type: %w", err)
	}

	bar := chart.Bar(visitCounts, true, darkMode)

	return rendering(w, r, statistics.Chart(bar, "Besuche pro Tag"))
}

// HandleWeekdayChart is the weekday chart handler
func (m *Repository) HandleWeekdayChart(w http.ResponseWriter, r *http.Request) error {
	// get date constraints
	minDate, maxDate, err := parseDateUrl(r)
	if err != nil {
		return err
	}

	// get dark mode
	darkMode := parseDarkModeUrl(r)

	visitCounts, err := m.DB.GetStatisticsAverageWeekday(minDate, maxDate)

	if err != nil {
		return fmt.Errorf("failed to get statistics for visits by weekday: %w", err)
	}

	bar := chart.Bar(visitCounts, true, darkMode)

	return rendering(w, r, statistics.Chart(bar, "Durchschnitt pro Wochentag"))
}

// HandleAtelierChart is the atelier chart handler
func (m *Repository) HandleAtelierChart(w http.ResponseWriter, r *http.Request) error {
	// get date constraints
	minDate, maxDate, err := parseDateUrl(r)
	if err != nil {
		return err
	}

	// get dark mode
	darkMode := parseDarkModeUrl(r)

	atelierCounts, err := m.DB.GetStatisticsAteliers(minDate, maxDate)

	if err != nil {
		return fmt.Errorf("failed to get statistics for visits by ateliers: %w", err)
	}

	pie := chart.Pie(atelierCounts, darkMode)

	return rendering(w, r, statistics.PieChart(pie, "Nutzung Ateliers"))
}

// HandleAgesChart is the ages chart handler
func (m *Repository) HandleAgesChart(w http.ResponseWriter, r *http.Request) error {
	// get date constraints
	minDate, maxDate, err := parseDateUrl(r)
	if err != nil {
		return err
	}

	// get dark mode
	darkMode := parseDarkModeUrl(r)

	agesCounts, err := m.DB.GetStatisticsAges(minDate, maxDate)

	if err != nil {
		return fmt.Errorf("failed to get statistics for visits by age group: %w", err)
	}

	pie := chart.Pie(agesCounts, darkMode)

	return rendering(w, r, statistics.PieChart(pie, "Nutzung nach Altersgruppe"))
}

// HandleKeyFigures is the key figures handler
func (m *Repository) HandleKeyFigures(w http.ResponseWriter, r *http.Request) error {
	// get date constraints
	minDate, maxDate, err := parseDateUrl(r)
	if err != nil {
		return err
	}

	figures, err := m.DB.GetStatisticsKeyFigures(minDate, maxDate)

	if err != nil {
		return fmt.Errorf("failed to get statistics key figures for visits: %w", err)
	}

	return rendering(w, r, statistics.FigureCard(figures, "Kennzahlen"))
}
