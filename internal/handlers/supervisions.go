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
	"github.com/faebut/prozessor_gostack/view/layout"
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
	loc, _ := time.LoadLocation("Europe/Zurich")
	startTime, _ := time.ParseInLocation("2006-01-02 15:04", formData.StartDate+" "+formData.StartTime, loc)
	endTime, _ := time.ParseInLocation("2006-01-02 15:04", formData.StartDate+" "+formData.EndTime, loc)

	var openingHours []models.OpeningHour

	if formData.Repetition {
		log.Println("repetition")
		repetitionEnd, _ := time.ParseInLocation("2006-01-02", formData.RepetitionEnd, loc)
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
	w.Header().Set("HX-Trigger", `{"closeModal": null, "supervisionUpdated": null}`)

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

	if dayCount == 7 {
		weekday := startDate.Weekday()
		if weekday == time.Sunday {
			startDate = startDate.AddDate(0, 0, -6)
		} else if weekday != time.Monday {
			startDate = startDate.AddDate(0, 0, -int(weekday-time.Monday))
		}
	}

	// calculate the urls for next and prev requests
	baseUrl := "/admin/supervisions/opening_hours/"
	prevURL := baseUrl + startDate.AddDate(0, 0, -dayCount).Format("2006-01-02") + "/" + strconv.Itoa(dayCount)
	nextURL := baseUrl + startDate.AddDate(0, 0, dayCount).Format("2006-01-02") + "/" + strconv.Itoa(dayCount)

	// get opening hours from database
	openingHours, err := m.DB.AllOpeningHours(startDate, dayCount)
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
	calData := supervisions.SupervisionCalendarData{
		Entries:   calEntries,
		StartDate: startDate,
		Daycount:  dayCount,
		PrevUrl:   prevURL,
		NextUrl:   nextURL,
	}

	return rendering(w, r, supervisions.SupervisionsCalendar(calData))
}

// HandleSuperVisionsPost handles the supervision post request
func (m *Repository) HandleSupervisionsSupervisionModal(w http.ResponseWriter, r *http.Request) error {
	dateID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		return fmt.Errorf("parsing opening hour id: %w", err)
	}

	currentVisitor, ok := m.App.Session.Get(r.Context(), "visitor").(models.Visitor)
	if !ok {
		return hxRedirect(w, r, "/admin/logout")
	}

	supervisors, err := m.DB.GetSupervisionsByDateId(dateID)
	if err != nil {
		return fmt.Errorf("getting supervisors: %w", err)
	}

	data := supervisions.SupervisionModalData{
		Supervisors:      supervisors,
		DateID:           dateID,
		CurrentVisitorID: currentVisitor.ID,
	}

	return rendering(w, r, supervisions.SupervisionModalContent(data))
}

func (m *Repository) HandleSuperVisionsPost(w http.ResponseWriter, r *http.Request) error {
	dateID, err := strconv.Atoi(chi.URLParam(r, "date_id"))
	if err != nil {
		return fmt.Errorf("parsing date id: %w", err)
	}

	visitorId := r.FormValue("visitor_id")
	if visitorId == "" {
		visitorId = r.FormValue("member")
	}

	err = m.DB.CreateSupervision(dateID, visitorId)
	if err != nil {
		return fmt.Errorf("adding supervision: %w", err)
	}

	logger.Success(r, m.App.Session, fmt.Sprintf("supervision added for opening hour %d for visitor %s", dateID, visitorId))

	supervisors, err := m.DB.GetSupervisionsByDateId(dateID)
	if err != nil {
		return fmt.Errorf("getting supervisors: %w", err)
	}

	currentVisitor, ok := m.App.Session.Get(r.Context(), "visitor").(models.Visitor)
	if !ok {
		return hxRedirect(w, r, "/admin/logout")
	}

	w.Header().Set("HX-Trigger", "supervisionUpdated")

	data := supervisions.SupervisionModalData{
		Supervisors:      supervisors,
		DateID:           dateID,
		CurrentVisitorID: currentVisitor.ID,
	}

	return rendering(w, r, supervisions.SupervisionModalContent(data))
}

func (m *Repository) HandleSuperVisionsDelete(w http.ResponseWriter, r *http.Request) error {
	visitorID := chi.URLParam(r, "visitor_id")
	dateID, err := strconv.Atoi(chi.URLParam(r, "date_id"))
	if err != nil {
		return fmt.Errorf("parsing date id: %w", err)
	}

	currentVisitor, ok := m.App.Session.Get(r.Context(), "visitor").(models.Visitor)
	if !ok {
		return hxRedirect(w, r, "/admin/logout")
	}

	role, _ := m.App.Session.Get(r.Context(), "permission").(models.Role)

	if !(role.EditAllSupervisions || (role.EditOwnSupervisions && currentVisitor.ID == visitorID)) {
		return rendering(w, r, layout.Error("Keine Berechtigung für diese Aktion"))
	}

	err = m.DB.DeleteSupervision(dateID, visitorID)
	if err != nil {
		return fmt.Errorf("deleting supervision: %w", err)
	}

	logger.Success(r, m.App.Session, fmt.Sprintf("supervision removed for opening hour %d for visitor %s", dateID, visitorID))

	supervisors, err := m.DB.GetSupervisionsByDateId(dateID)
	if err != nil {
		return fmt.Errorf("getting supervisors: %w", err)
	}

	w.Header().Set("HX-Trigger", "supervisionUpdated")

	data := supervisions.SupervisionModalData{
		Supervisors:      supervisors,
		DateID:           dateID,
		CurrentVisitorID: currentVisitor.ID,
	}

	return rendering(w, r, supervisions.SupervisionModalContent(data))
}

// HandleSuperVisionsOpeningHoursDelete handles the deletion of an opening hour
func (m *Repository) HandleSuperVisionsOpeningHoursDelete(w http.ResponseWriter, r *http.Request) error {
	ohID, err := strconv.Atoi(chi.URLParam(r, "oh_id"))
	if err != nil {
		return fmt.Errorf("parsing date id: %w", err)
	}

	err = m.DB.DeleteOpeningHour(ohID)
	if err != nil {
		return fmt.Errorf("deleting opening hour: %w", err)
	}

	logger.Success(r, m.App.Session, fmt.Sprintf("removed opening hour %d", ohID))

	w.Header().Set("HX-Trigger", "supervisionUpdated")

	return nil
}
