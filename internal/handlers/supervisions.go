package handlers

import (
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/faebut/prozessor_gostack/internal/events"
	"github.com/faebut/prozessor_gostack/internal/forms"
	"github.com/faebut/prozessor_gostack/internal/helpers"
	"github.com/faebut/prozessor_gostack/internal/logger"
	"github.com/faebut/prozessor_gostack/internal/models"
	"github.com/faebut/prozessor_gostack/view/supervisions"
	"github.com/go-chi/chi/v5"
)

// HandleSupervisionsIndex is the supervisions page handler
func (m *Repository) HandleSupervisionsIndex(w http.ResponseWriter, r *http.Request) error {
	// get visitor
	visitor, ok := m.App.Session.Get(r.Context(), "visitor").(models.Visitor)
	if !ok {
		slog.Error("could not get visitor from session", "path", r.URL.Path)
		m.App.Session.Put(r.Context(), "error", "Can't get visitor from session")

		return hxRedirect(w, r, "/admin/logout")
	}

	return rendering(w, r, supervisions.Supervisions(visitor))
}

// HandleSupervisionsOpeningHoursNew gets a new opening hours creation form
func (m *Repository) HandleSupervisionsOpeningHoursNew(w http.ResponseWriter, r *http.Request) error {
	return rendering(w, r, supervisions.CreateForm(supervisions.OpeningHoursForm{}, supervisions.OpeningHoursFormErrors{}))
}

// HandleSupervisionsOpeningHoursCreate handles the opening hours creation form
func (m *Repository) HandleSupervisionsOpeningHoursCreate(w http.ResponseWriter, r *http.Request) error {
	var formData supervisions.OpeningHoursForm
	var formErrors supervisions.OpeningHoursFormErrors

	// validation
	form := forms.New(r.PostForm)
	form.Required("start_date", "start_time", "end_time")
	form.MapForm(&formData)

	if formData.Repetition {
		form.Required("repetition_end")
	}

	form.MapErrors(&formErrors)

	// what to do if form not valid
	if !form.Valid() {
		// send the data to the template
		return rendering(w, r, supervisions.CreateForm(
			formData,
			formErrors,
		))
	}

	// if everything is good lets build our opening hours slice
	startTime, _ := time.Parse("2006-01-02 15:04", formData.StartDate+" "+formData.StartTime)
	endTime, _ := time.Parse("2006-01-02 15:04", formData.StartDate+" "+formData.EndTime)

	var openingHours []models.OpeningHour

	if formData.Repetition {
		log.Println("repetition")
		repetitionEnd, _ := time.Parse("2006-01-02", formData.RepetitionEnd)
		for _, d := range helpers.DateRepetition(startTime, repetitionEnd) {
			openingHours = append(openingHours, models.OpeningHour{
				StartTime: d,
				EndTime:   d.Add(endTime.Sub(startTime)),
			})
		}
	} else {
		log.Println("no repetition")
		openingHours = append(openingHours, models.OpeningHour{
			StartTime: startTime,
			EndTime:   endTime,
		})
	}

	// write to database
	err := m.DB.CreateOpeningHours(openingHours)
	if err != nil {
		return fmt.Errorf("failed to create opening hours: %w", err)
	}

	// log success
	logger.Success(r, m.App.Session,
		fmt.Sprintf("%d opening hour(s) starting from %s successfully added", len(openingHours), startTime.Format("2006-01-02")))

	// close dialog
	w.Header().Set("HX-Trigger", "closeModal")

	return nil
}

// HandleSupervisionsOpeningHours gets all opening hours and returns the supervisions calendar with all dates and supervisions
func (m *Repository) HandleSupervisionsOpeningHours(w http.ResponseWriter, r *http.Request) error {
	// get startDate and dayCount from url
	startDate, err := time.Parse("2006-01-02", chi.URLParam(r, "startdate"))
	if err != nil {
		return fmt.Errorf("failed to parse startdate: %s", err)
	}
	dayCount, err := strconv.Atoi(chi.URLParam(r, "daycount"))
	if err != nil {
		return fmt.Errorf("failed to parse daycount: %s", err)
	}

	// calculate the urls for next and prev requests
	baseUrl := "/admin/supervisions/opening_hours/"
	prevURL := baseUrl + startDate.AddDate(0, 0, -dayCount).Format("2006-01-02") + "/" + strconv.Itoa(dayCount)
	nextURL := baseUrl + startDate.AddDate(0, 0, dayCount).Format("2006-01-02") + "/" + strconv.Itoa(dayCount)

	openingHours, err := m.DB.AllOpeningHours(startDate, dayCount)
	if err != nil {
		return fmt.Errorf("failed to get opening hours: %w", err)
	}

	// calculate calendar entries
	entries := make([]models.CalendarEntry, 0, len(openingHours))
	for i := range openingHours {
		entries = append(entries, &openingHours[i])
	}
	calEntries := events.LayoutEvents(entries)

	// create calendar data
	calData := supervisions.SupervisionCalendarData{
		Entries:   calEntries,
		StartDate: startDate,
		Daycount:  dayCount,
		PrevUrl:   prevURL,
		NextUrl:   nextURL,
	}

	return rendering(w, r, supervisions.SupervisionsCalendar(calData))
}
