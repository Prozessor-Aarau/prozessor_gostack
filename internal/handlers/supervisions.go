package handlers

import (
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"time"

	"github.com/faebut/prozessor_gostack/internal/forms"
	"github.com/faebut/prozessor_gostack/internal/helpers"
	"github.com/faebut/prozessor_gostack/internal/logger"
	"github.com/faebut/prozessor_gostack/internal/models"
	"github.com/faebut/prozessor_gostack/view/supervisions"
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

func (m *Repository) HandleSupervisionsOpeningHoursNew(w http.ResponseWriter, r *http.Request) error {
	return rendering(w, r, supervisions.CreateForm(supervisions.OpeningHoursForm{}, supervisions.OpeningHoursFormErrors{}))
}

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
