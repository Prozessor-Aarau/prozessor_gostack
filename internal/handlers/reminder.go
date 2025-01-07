package handlers

import (
	"log"
	"net/http"
	"time"

	"github.com/a-h/templ"
	"github.com/faebut/prozessor_gostack/internal/forms"
	"github.com/faebut/prozessor_gostack/internal/models"
	"github.com/faebut/prozessor_gostack/view/reminder"
)

// HandleReminderIndex is the reminder page handler
func (m *Repository) HandleReminderIndex(w http.ResponseWriter, r *http.Request) error {
	visitor, ok := m.App.Session.Get(r.Context(), "visitor").(models.Visitor)
	if !ok {
		log.Println("could not get visitor from Session")
		return hxRedirect(w, r, "/admin/logout")
	}

	return rendering(w, r, reminder.Reminder(visitor))
}

// HandleReminderList handles the list of visitors to remind
func (m *Repository) HandleReminderList(w http.ResponseWriter, r *http.Request) error {
	visitors, count, err := m.DB.AllToRemindVisitors()
	if err != nil {
		log.Println("could not get reminder visitors from DB: ", err)
		return err
	}

	return rendering(w, r, reminder.ReminderList(visitors, count))
}

// HandleReminderListPost handles the reminder list request before sending reminders and shows list for validation
func (m *Repository) HandleReminderListPost(w http.ResponseWriter, r *http.Request) error {
	visitors, _, err := m.DB.AllToRemindVisitors()
	if err != nil {
		log.Println("could not get reminder visitors from DB: ", err)
		return err
	}

	// check for all ids if the checkbox is checked and add visitor to new slice
	form := forms.New(r.Form)
	var selectedVisitors []models.ReminderVisitor
	for _, v := range visitors {
		if form.BoxValue(v.Visitor.ID) {
			selectedVisitors = append(selectedVisitors, v)
		}
	}

	// get all checkboxes and check if they are checked. If not remove string
	return rendering(w, r, reminder.ReminderCreateForm(selectedVisitors))
}

// HandleReminderProgressIndex handles the start of sending the reminders and renders the progress page
func (m *Repository) HandleReminderProgress(w http.ResponseWriter, r *http.Request) {
	visitors, _, err := m.DB.AllToRemindVisitors()
	if err != nil {
		log.Println("could not get reminder visitors from DB: ", err)
	}

	// check for all ids if the checkbox is checked and add visitor to new slice
	// var selectedVisitors []models.ReminderVisitor
	// for _, v := range visitors {
	// 	value := r.Form.Get(v.Visitor.ID)
	// 	if len(value) > 0 {
	// 		selectedVisitors = append(selectedVisitors, v)
	// 	}
	// }
	//
	// m.App.Reminders = &selectedVisitors

	// get length
	count := len(visitors)

	// Create a channel to send data to the template.
	data := make(chan models.Reminder)
	// Run a background process that will take 10 seconds to complete.
	go func() {
		// Always remember to close the channel.
		defer close(data)
		for i, v := range visitors {
			select {
			case <-r.Context().Done():
				// Quit early if the client is no longer connected.
				return
			case <-time.After(time.Second):
				// Send a new piece of data to the channel.
				data <- models.Reminder{
					Count:      count,
					Progress:   i + 1,
					FirstName:  v.Visitor.FirstName,
					FamilyName: v.Visitor.FamilyName,
					Email:      v.Visitor.Email,
				}
			}
		}
	}()

	// Pass the channel to the template.
	component := reminder.ReminderProgressPage(data)

	// Serve using the streaming mode of the handler.
	templ.Handler(component, templ.WithStreaming()).ServeHTTP(w, r)
}
