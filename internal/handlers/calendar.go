package handlers

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/faebut/prozessor_gostack/internal/events"
	"github.com/faebut/prozessor_gostack/internal/logger"
	"github.com/faebut/prozessor_gostack/internal/models"
	"github.com/faebut/prozessor_gostack/view/calendar"
	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"
)

// HandleCalendarIndex is the calendar page handler
func (m *Repository) HandleCalendarIndex(w http.ResponseWriter, r *http.Request) error {
	// load env file to read base URL
	err := godotenv.Load()
	if err != nil {
		logger.Fatal("loading .env file", slog.String("error", err.Error()))
	}

	baseUrl := os.Getenv("BASE_URL")
	baseUrl += "/calendar"

	return rendering(w, r, calendar.Calendar(baseUrl))
}

// HandleCalendar handles the external calendar for opening hours and events
func (m *Repository) HandleCalendar(w http.ResponseWriter, r *http.Request) error {
	// get startDate and dayCount from url
	startDate, err := time.Parse("2006-01-02", chi.URLParam(r, "startdate"))
	if err != nil {
		return fmt.Errorf("failed to parse startdate: %s", err)
	}
	dayCount, err := strconv.Atoi(chi.URLParam(r, "daycount"))
	if err != nil {
		return fmt.Errorf("failed to parse daycount: %s", err)
	}

	if dayCount == 7 {
		weekday := startDate.Weekday()
		if weekday == time.Sunday {
			startDate = startDate.AddDate(0, 0, -6)
		} else if weekday != time.Monday {
			startDate = startDate.AddDate(0, 0, -int(weekday-time.Monday))
		}
	}

	// calculate the urls for next and prev requests
	// load env file to read base URL
	err = godotenv.Load()
	if err != nil {
		logger.Fatal("loading .env file", slog.String("error", err.Error()))
	}

	baseUrl := os.Getenv("BASE_URL")
	baseUrl += "/calendar/"

	prevURL := baseUrl + startDate.AddDate(0, 0, -dayCount).Format("2006-01-02") + "/" + strconv.Itoa(dayCount)
	nextURL := baseUrl + startDate.AddDate(0, 0, dayCount).Format("2006-01-02") + "/" + strconv.Itoa(dayCount)

	// get opening hours from database
	openingHours, err := m.DB.CalendarOpeningHours(startDate, dayCount)
	if err != nil {
		return fmt.Errorf("failed to get opening hours: %w", err)
	}

	// get events from events cache
	calEvents, err := events.GetCalendarEvents()
	if err != nil {
		return fmt.Errorf("failed to get events: %w", err)
	}

	// calculate calendar entries
	entries := make([]models.CalendarEntry, 0, len(openingHours)+len(calEvents))
	for i := range openingHours {
		entries = append(entries, &openingHours[i])
	}
	for i := range calEvents {
		entries = append(entries, &calEvents[i])
	}

	calEntries := events.LayoutEvents(entries)

	// create calendar data
	calData := calendar.CalendarData{
		Entries:   calEntries,
		StartDate: startDate,
		Daycount:  dayCount,
		PrevUrl:   prevURL,
		NextUrl:   nextURL,
	}

	return rendering(w, r, calendar.CalendarView(calData))
}
