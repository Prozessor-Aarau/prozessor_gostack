package handlers

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/a-h/templ"
	"github.com/faebut/prozessor_gostack/internal/billing"
	"github.com/faebut/prozessor_gostack/internal/forms"
	"github.com/faebut/prozessor_gostack/internal/mailtemplates"
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

	return rendering(w, r, reminder.ReminderList(visitors, count, m.App.Reminder.Active))
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
	m.App.Reminder.WaitList = make(chan models.MailData, count)
	m.App.Reminder.Active = true

	for i, v := range visitors {

		// create template
		template, err := mailtemplates.Reminder(v)
		if err != nil {
			log.Println("error creating template:", err)
			return
		}

		// set amount
		var amount float64
		if v.RessortID > 0 {
			amount = 320
		} else {
			amount = 400
		}

		// create new start date
		var newStartDate time.Time
		if !v.EndDate.IsZero() {
			newStartDate = v.EndDate.AddDate(0, 0, 1)
		} else {
			// if there was no abo before, set it to the June 2024 date
			newStartDate, _ = time.Parse("2006-01-02", "2024-06-01")
		}

		// create bill
		file, err := billing.CreateBill(
			models.BillData{
				Purpose: "abonnement",
				Visitor: v.Visitor,
				Amount:  amount,
				Infos: fmt.Sprintf(
					"Jahresabonnement für %s %s, gültig ab: %s",
					v.Visitor.FirstName,
					v.Visitor.FamilyName,
					newStartDate.Format("02.01.2006"),
				),
				Message: fmt.Sprintf(
					"Abo %s %s",
					v.Visitor.FirstName,
					v.Visitor.FamilyName,
				),
			},
			m.DB,
		)

		if err != nil {
			log.Println("error creating bill:", err)
			return
		}

		// create model for mail
		msg := models.MailData{
			To:      v.Visitor.Email,
			Subject: "Erneuerung Abonnement",
			Content: template,
			File:    file,
		}

		// Set date of reminder in DB, for those who already had an abonnement
		if v.AboID > 0 {
			err := m.DB.SetReminderDate(v)
			if err != nil {
				log.Println("error setting reminder date:", err)
				return
			}
		}

		// send to channel
		m.App.MailChan <- msg

		// put info in the channel to be read in frontend
		m.App.Reminder.WaitList <- models.MailData{
			Reason:     "Reminder",
			Count:      count,
			Progress:   i + 1,
			FirstName:  v.Visitor.FirstName,
			FamilyName: v.Visitor.FamilyName,
			To:         v.Visitor.Email,
		}
		time.Sleep(time.Millisecond * 500)
	}

	log.Println("closing channel")
	m.App.Reminder.Active = false
	close(m.App.Reminder.WaitList)
}

// HandleReminderProgressBar handles the SSE channel to send an updated progress bar for the sending of reminders
func (m *Repository) HandleReminderProgressBar(w http.ResponseWriter, r *http.Request) {
	// Set headers for SSE
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	// Create a context for handling client disconnection
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	// Send data to the client
	for data := range m.App.Reminder.WaitList {
		select {
		case <-r.Context().Done():
			// Quit early if the client is no longer connected.
			return
		default:
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
