package handlers

import (
	"context"
	"fmt"
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
func (m *Repository) HandleReminderSend(w http.ResponseWriter, r *http.Request) error {
	visitors, _, err := m.DB.AllToRemindVisitors()
	if err != nil {
		log.Println("could not get reminder visitors from DB: ", err)
	}

	// check for all ids if the checkbox is checked and add visitor to new slice
	var selectedVisitors []models.ReminderVisitor
	for _, v := range visitors {
		value := r.Form.Get(v.Visitor.ID)
		if len(value) > 0 {
			selectedVisitors = append(selectedVisitors, v)
		}
	}

	// send reminders in seperate goroutine
	go m.sendReminder(selectedVisitors)

	return rendering(w, r, reminder.ReminderProgressPage())
}

// sendReminder sends the reminders to the mail channel and the channel for the frontend visualization
func (m *Repository) sendReminder(visitors []models.ReminderVisitor) {

	// set up channel
	count := len(visitors)

	// Open new Channel to send data to the template.
	log.Println("open new channel")
	m.App.WaitMailChan = make(chan models.MailData, count)

	for i, v := range visitors {
		log.Println(i, ":", v.Visitor.FirstName, v.Visitor.FamilyName)

		// put info in the channel to be read in frontend
		m.App.WaitMailChan <- models.MailData{
			Count:      count,
			Progress:   i + 1,
			FirstName:  v.Visitor.FirstName,
			FamilyName: v.Visitor.FamilyName,
			To:         v.Visitor.Email,
		}
		time.Sleep(time.Millisecond * 500)
	}

	log.Println("closing channel")
	close(m.App.WaitMailChan)
}

func (m *Repository) HandleReminderProgressBar(w http.ResponseWriter, r *http.Request) {
	// Set headers for SSE
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	// Create a context for handling client disconnection
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	// Send data to the client
	for data := range m.App.WaitMailChan {
		select {
		case <-r.Context().Done():
			// Quit early if the client is no longer connected.
			return
		default:
			log.Println("Now with:", data.FirstName)
			template, err := templ.ToGoHTML(ctx, reminder.ReminderProgressBar(data))
			if err != nil {
				log.Println(err)
			}

			fmt.Fprintf(w, "data: %s\n\n", template)
			w.(http.Flusher).Flush()
		}
	}

	// send event to stop listening
	fmt.Fprint(w, "event: stop\ndata: stop\n\n")
	w.(http.Flusher).Flush()
}
