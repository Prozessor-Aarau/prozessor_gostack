package handlers

import (
	"fmt"
	"log/slog"
	"math"
	"net/http"
	"strconv"

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

	// send the data to the template
	logList, count, err := m.DB.AllLogs(page, limit)

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
