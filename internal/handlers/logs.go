package handlers

import (
	"fmt"
	"log/slog"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/faebut/prozessor_gostack/internal/models"
	"github.com/faebut/prozessor_gostack/view/logs"
)

// HandleLogIndex is the logs page handler
func (m *Repository) HandleLogIndex(w http.ResponseWriter, r *http.Request) error {
	// send the data to the template
	visitor, ok := m.App.Session.Get(r.Context(), "visitor").(models.Visitor)
	if !ok {
		slog.Error("could not get visitor from session", "path", r.URL.Path)
		m.App.Session.Put(r.Context(), "error", "Can't get visitor from session")

		return hxRedirect(w, r, "/admin/logout")
	}

	return rendering(w, r, logs.Index(visitor))
}

// HandleLogList is the logs list handler
func (m *Repository) HandleLogList(w http.ResponseWriter, r *http.Request) error {
	// get page
	page, err := strconv.Atoi(r.URL.Query().Get("page"))
	if err != nil {
		return fmt.Errorf("failed to convert string to int: %w", err)
	}

	// get limit
	limit, err := strconv.Atoi(r.URL.Query().Get("limit"))
	if err != nil {
		return fmt.Errorf("failed to convert string to int: %w", err)
	}

	// get min date
	minDate, err := time.Parse("2006-01-02", r.URL.Query().Get("min"))
	if err != nil {
		return fmt.Errorf("failed to convert string to date: %w", err)
	}

	// get max date
	maxDate, err := time.Parse("2006-01-02", r.URL.Query().Get("max"))
	if err != nil {
		return fmt.Errorf("failed to convert string to date: %w", err)
	}

	// send the data to the template
	logList, count, err := m.DB.AllLogs(minDate, maxDate, page, limit)

	if err != nil {
		return fmt.Errorf("failed to get logs: %w", err)
	}

	// render the template
	return rendering(w, r, logs.LogsList(
		logList,
		page,
		int(math.Ceil(float64(count)/float64(limit))),
	))
}

// HandleLogsFilter handles the logs filtering
func (m *Repository) HandleLogsFilter(w http.ResponseWriter, r *http.Request) error {
	// get date constraints from the form
	minString := r.Form.Get("min_date")
	maxString := r.Form.Get("max_date")
	limitString := r.Form.Get("limit")

	minDate, err := time.Parse("2006-01-02", minString)
	maxDate, err := time.Parse("2006-01-02", maxString)
	if err != nil {
		return fmt.Errorf("failed to parse date: %w", err)
	}

	limit, err := strconv.Atoi(limitString)
	if err != nil {
		return fmt.Errorf("failed to convert string to int: %w", err)
	}

	logList, count, err := m.DB.AllLogs(minDate, maxDate, 1, limit)

	// render the template
	return rendering(w, r, logs.LogsList(
		logList,
		1,
		int(math.Ceil(float64(count)/float64(limit))),
	))
}
