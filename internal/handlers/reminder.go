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
func (m *Repository) HandleReminderProgressPage(w http.ResponseWriter, r *http.Request) error {
	return rendering(w, r, reminder.ReminderProgressPage())
}

func (m *Repository) HandleReminderProgressBar(w http.ResponseWriter, r *http.Request) {
	// get visitors to remind TODO this needs to change to only take the selected ones
	visitors, _, err := m.DB.AllToRemindVisitors()
	if err != nil {
		log.Println("could not get reminder visitors from DB: ", err)
	}
	//
	// // check for all ids if the checkbox is checked and add visitor to new slice
	// // var selectedVisitors []models.ReminderVisitor
	// // for _, v := range visitors {
	// // 	value := r.Form.Get(v.Visitor.ID)
	// // 	if len(value) > 0 {
	// // 		selectedVisitors = append(selectedVisitors, v)
	// // 	}
	// // }
	// //
	// // m.App.Reminders = &selectedVisitors
	//
	// get length
	count := len(visitors)

	// Create a channel to send data to the template.
	dataCh := make(chan models.MailData)

	// go func() {
	// 	// Always remember to close the channel.
	// 	defer close(data)
	// 	for i, v := range visitors {
	// 		select {
	// 		case <-r.Context().Done():
	// 			// Quit early if the client is no longer connected.
	// 			return
	// 		case <-time.After(time.Second):
	// 			// Send a new piece of data to the channel.
	// 			data <- models.Reminder{
	// 				Count:      count,
	// 				Progress:   i + 1,
	// 				FirstName:  v.Visitor.FirstName,
	// 				FamilyName: v.Visitor.FamilyName,
	// 				Email:      v.Visitor.Email,
	// 			}
	// 		}
	// 	}
	// }()
	//
	// Set headers for SSE
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	// Create a context for handling client disconnection
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	// TODO quit loop when channel is closed
	// TODO do something and get channel from somewhere else
	// Send data to the client
	go func() {
		for data := range dataCh {
			template, err := templ.ToGoHTML(ctx, reminder.ReminderProgressBar(data))
			if err != nil {
				log.Println(err)
			}

			fmt.Fprintf(w, "data: %s\n\n", template)
			w.(http.Flusher).Flush()
		}
	}()

	// Simulate sending data periodically
	for i, v := range visitors {
		dataCh <- models.MailData{
			Count:      count,
			Progress:   i + 1,
			FirstName:  v.Visitor.FirstName,
			FamilyName: v.Visitor.FamilyName,
			To:         v.Visitor.Email,
		}
		time.Sleep(300 * time.Millisecond)
	}

	close(dataCh)
	log.Println("channel closed")
	// TODO how to tell client to not open connection again, because it will reset
	// TODO as channel will be filled elsewhere and the is empty, maybe this is not a Problem.
}
