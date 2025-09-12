package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/a-h/templ"
	"github.com/faebut/prozessor_gostack/internal/events"
	"github.com/faebut/prozessor_gostack/internal/forms"
	"github.com/faebut/prozessor_gostack/internal/logger"
	"github.com/faebut/prozessor_gostack/internal/mailtemplates"
	"github.com/faebut/prozessor_gostack/internal/models"
	"github.com/faebut/prozessor_gostack/view/newsletter"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/nfnt/resize"
	"golang.org/x/net/html"
)

// TODO: there is a lot of code repetition in here, needs refactoring

// HandleNewsletterIndex is the newsletter page handler
func (m *Repository) HandleNewsletterIndex(w http.ResponseWriter, r *http.Request) error {
	// get visitor
	visitor, ok := m.App.Session.Get(r.Context(), "visitor").(models.Visitor)
	if !ok {
		slog.Error("could not get visitor from session", "path", r.URL.Path)
		m.App.Session.Put(r.Context(), "error", "Can't get visitor from session")

		return hxRedirect(w, r, "/admin/logout")
	}
	return rendering(w, r, newsletter.Newsletters(visitor))
}

// HandleNewsletterList gets and returns the unfiltered newsletter list
func (m *Repository) HandleNewsletterList(w http.ResponseWriter, r *http.Request) error {

	newsletterList, _, err := m.DB.GetAllNewsletters()
	if err != nil {
		return fmt.Errorf("failed to get newsletters: %w", err)
	}

	return rendering(w, r, newsletter.NewsletterList(newsletterList))
}

// HandleNewsletterCreateIndex returns the form for a new newsletter
func (m *Repository) HandleNewsletterCreateIndex(w http.ResponseWriter, r *http.Request) error {
	// create new empty structs
	var nl models.Newsletter
	var e newsletter.NewsletterErrors

	templates, err := m.DB.GetNewsletterTemplates()
	if err != nil {
		return fmt.Errorf("failed to get templates: %w", err)
	}

	var elements []models.NewsletterElement

	// render template
	return rendering(w, r, newsletter.NewsletterCreate(nl, templates, elements, e))
}

// HandleNewsletterElementCreate handles the creation of a new newsletter element
func (m *Repository) HandleNewsletterElementCreate(w http.ResponseWriter, r *http.Request) error {
	var element models.NewsletterElement

	elType := r.Form.Get("eltype")

	if len(elType) == 0 {
		return errors.New("No type selected")
	}

	// create a new uuid for the element
	uuid, err := uuid.NewUUID()
	if err != nil {
		return fmt.Errorf("failed to create uuid: %w", err)
	}
	element.ID = uuid.String()
	element.Type = elType

	return rendering(w, r, newsletter.Element(element, []newsletter.ElementError{}))
}

func (m *Repository) HandleNewsletterEventsLoad(w http.ResponseWriter, r *http.Request) error {
	// get events
	events, err := events.GetEvents()
	if err != nil {
		return fmt.Errorf("failed to get events: %w", err)
	}

	// put events into string
	eventString, err := json.Marshal(events)
	if err != nil {
		return fmt.Errorf("failed marshal events to string: %w", err)
	}

	return rendering(w, r, newsletter.EventList(events, string(eventString)))
}

// HandleNewsletterCreateIndex returns the form for a new newsletter
func (m *Repository) HandleNewsletterCreateEdit(w http.ResponseWriter, r *http.Request) error {
	// set initial value
	edit := false

	// create newsletter struct
	var nl models.Newsletter

	// get route from request url
	exploded := strings.Split(r.RequestURI, "/")
	if exploded[3] == "edit" {
		id, err := strconv.Atoi(chi.URLParam(r, "id"))
		if err != nil {
			return fmt.Errorf("failed to convert string to int: %w", err)
		}
		nl.ID = id
		edit = true
	}

	templates, err := m.DB.GetNewsletterTemplates()
	if err != nil {
		return fmt.Errorf("failed to get templates: %w", err)
	}

	// parse form
	err = r.ParseMultipartForm(5 << 20)
	if err != nil {
		return fmt.Errorf("failed to parse multipart form: %w", err)
	}

	// handle elements
	var element_id []string
	var element_position []int
	var element_type []string
	var element_content []string
	var element_alttext []string
	var element_url []string

	for key, values := range r.Form {

		if key == "element_id" {
			for _, val := range values {
				element_id = append(element_id, val)
			}
		}

		if key == "element_position" {
			for _, val := range values {
				position, err := strconv.Atoi(val)
				if err != nil {
					return fmt.Errorf("failed to convert string to int: %w", err)
				}
				element_position = append(element_position, position)
			}
		}

		if key == "element_type" {
			for _, val := range values {
				element_type = append(element_type, val)
			}
		}

		if key == "element_content" {
			for _, val := range values {
				element_content = append(element_content, val)
			}
		}

		if key == "element_alttext" {
			for _, val := range values {
				element_alttext = append(element_alttext, val)
			}
		}

		if key == "element_url" {
			for _, val := range values {
				element_url = append(element_url, val)
			}
		}
	}

	// build elements and push to slice
	form := forms.New(r.Form)
	var elements []models.NewsletterElement

	for i := range element_id {
		element := models.NewsletterElement{
			ID:       element_id[i],
			Position: element_position[i],
			Type:     element_type[i],
			Content:  element_content[i],
			AltText:  element_alttext[i],
			Url:      element_url[i],
		}

		elements = append(elements, element)
	}

	var elementErrors []newsletter.ElementError
	// save filetype
	var filetype string

	for i, e := range elements {
		isError := false
		var eError newsletter.ElementError

		switch e.Type {
		case "text":
			if len(e.Content) == 0 {
				eError.Content = "Feld ist obligatorisch"
				isError = true
			}
		case "button":
			if len(e.Content) == 0 {
				eError.Content = "Feld ist obligatorisch"
				isError = true
			}
			if len(e.Url) == 0 {
				eError.URL = "Feld ist obligatorisch"
				isError = true
			}
			if len(e.Url) > 0 {
				if !regexp.MustCompile(`[-a-zA-Z0-9@:%._\+~#=]{1,256}\.[a-zA-Z0-9()]{1,6}\b([-a-zA-Z0-9()@:%_\+.~#?&//=]*)`).MatchString(e.Url) {
					eError.URL = "keine gültige Url"
					isError = true
				}
			}
		case "picture":
			if len(e.Content) == 0 {
				eError.Content = "Feld ist obligatorisch"
				isError = true
			}
			if len(e.AltText) == 0 {
				eError.AltText = "Feld ist obligatorisch"
				isError = true
			}

			// check if there is already a file
			oldFile := r.Form.Get(e.ID + "_upload")

			// if there is, set the upload in this element to old filename
			if len(oldFile) > 0 {
				elements[i].Upload = oldFile
			} else {

				// else check the fileupload element
				file, _, err := r.FormFile(e.ID + "_fileupload")
				if err != nil {
					eError.Upload = "gültiges File angeben"
					isError = true
				} else {
					defer file.Close()
					buff := make([]byte, 512) // docs tell that it take only first 512 bytes into consideration
					if _, err = file.Read(buff); err != nil {
						fmt.Println(err) // do something with that error
					}

					// get ft
					ft := http.DetectContentType(buff)

					// what to do if it's not an image
					if !strings.Contains(ft, "image") {
						eError.Upload = "nur Bilder erlaubt (png/jpeg)"
						isError = true
					} else {
						// save img type to filetype
						filetype = strings.Split(ft, "/")[1]
						// check if its png or jpeg
						if filetype != "png" && filetype != "jpeg" {
							eError.Upload = "nur Bilder erlaubt (png/jpeg)"
							isError = true
						}
					}
				}
			}

		}

		if isError {
			eError.ID = e.ID
			elementErrors = append(elementErrors, eError)
			form.Errors.Add("item", "Feld ist obligatorisch")
		}
	}

	nl.Elements = elements

	// handle other fields
	formErrors := newsletter.NewsletterErrors{}
	formErrors.Elements = elementErrors
	// set edit if it's edit
	formErrors.Edit = edit

	// check if there are items and set error if not
	if len(nl.Elements) == 0 {
		form.Errors.Add("items", "Mindestens ein Element hinzufügen")
		formErrors.Item = "Mindestens ein Element hinzufügen"
	}

	// check validity of fields
	form.Required("title", "description", "template")
	formErrors.Title = form.Errors.Get("title")
	formErrors.Description = form.Errors.Get("description")
	formErrors.Template = form.Errors.Get("template")

	nl.Title = r.Form.Get("title")
	nl.Description = r.Form.Get("description")
	nl.Template.ID, _ = strconv.Atoi(r.Form.Get("template"))
	nl.RecipientGroup = r.Form.Get("recipient_group")
	nl.ColorBase = r.Form.Get("color_base")
	nl.ColorAccent = r.Form.Get("color_accent")

	// what to do if form not valid
	if !form.Valid() {
		// send the data to the template
		return rendering(w, r, newsletter.NewsletterCreate(nl, templates, elements, formErrors))
	}

	// handling file
	// Retrieve the file from form data if there is not an old one
	for i, e := range elements {
		if len(e.AltText) > 0 && len(e.Upload) == 0 {
			file, fileHeader, err := r.FormFile(e.ID + "_fileupload")
			if err != nil {
				return fmt.Errorf("failed handle file: %w", err)
			} else {
				defer file.Close()

				// Generate a unique filename to prevent overwriting and conflicts
				uuid, err := uuid.NewRandom()
				if err != nil {
					return fmt.Errorf("failed to generate uuid: %w", err)
				}
				// Append the filetype extension to uuid to create filename
				filename := uuid.String() + filepath.Ext(fileHeader.Filename)

				// TODO: don't save on db server but on webserver. Use ftp implementation but needs login with more permission

				// check if uploads directory exists, if not create it
				if _, err := os.Stat("static/newsletter"); err != nil {
					os.Mkdir("static/newsletter/", os.ModePerm)
				}

				// Create the full path for saving the file
				filePath := filepath.Join("static/newsletter", filename)

				// Save the file to the server
				dst, err := os.Create(filePath)
				if err != nil {
					return fmt.Errorf("failed to save file: %w", err)
				}
				defer dst.Close()

				// convert to image
				var imgFile image.Image
				switch filetype {
				case "jpeg":
					imgFile, err = jpeg.Decode(file)
				case "png":
					imgFile, err = png.Decode(file)
				}
				if err != nil {
					return fmt.Errorf("failed decode file: %w", err)
				}

				// resize image
				resFile := resize.Resize(600, 0, imgFile, resize.Lanczos3)

				// write image
				switch filetype {
				case "jpeg":
					err = jpeg.Encode(dst, resFile, nil)
				case "png":
					err = png.Encode(dst, resFile)
				}
				if err != nil {
					return fmt.Errorf("failed encode file: %w", err)
				}

				// Update the element with the update. Need to use index
				// here to modify the struct directly
				elements[i].Upload = filename
			}
		}
	}

	// what to do if form valid
	var action string
	if formErrors.Edit {
		err = m.DB.UpdateNewsletter(nl)
		action = "updated"
		if err != nil {
			return fmt.Errorf("failed to update newsletter: %w", err)
		}
	} else {
		err = m.DB.CreateNewsletter(nl)
		action = "created"
		if err != nil {
			return fmt.Errorf("failed to create newsletter: %w", err)
		}
	}

	// TODO: do a better workflow

	// log success
	logger.Success(r, m.App.Session,
		fmt.Sprintf("Newsletter %s successfully %s", nl.Title, action))

	redirectURL := "/admin/newsletter"

	return hxRedirect(w, r, redirectURL)
}

// HandleNewsletterEditIndex returns the form for an existing newsletter
func (m *Repository) HandleNewsletterEditIndex(w http.ResponseWriter, r *http.Request) error {
	nlID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		return fmt.Errorf("failed convert string to int: %w", err)
	}

	// get newsletter from database
	nl, err := m.DB.GetNewsletterByID(nlID)
	if err != nil {
		return fmt.Errorf("failed get newsletter: %w", err)
	}

	// get elements for newsletter
	elements, err := m.DB.GetAllNewsletterElementsByID(nl.ID)
	if err != nil {
		return fmt.Errorf("failed get newsletter elements: %w", err)
	}

	nl.Elements = elements

	templates, err := m.DB.GetNewsletterTemplates()
	if err != nil {
		return fmt.Errorf("failed get templates: %w", err)
	}

	// render template
	return rendering(w, r, newsletter.NewsletterCreate(
		nl,
		templates,
		elements, newsletter.NewsletterErrors{Edit: true},
	))
}

func (m *Repository) HandleNewsletterImageDelete(w http.ResponseWriter, r *http.Request) error {
	eID := chi.URLParam(r, "id")

	// prepare form data and get picture name
	r.ParseForm()
	picture := r.Form.Get(eID + "_upload")

	// delete picture from disk
	picPath := filepath.Join("static/newsletter/", picture)
	if err := os.Remove(picPath); err != nil {
		return fmt.Errorf("failed get delete image from filesystem: %w", err)
	}

	// delete image in database and get element
	el, err := m.DB.DeleteNewsletterImage(eID)
	if err != nil {
		return fmt.Errorf("failed get delete image from DB: %w", err)
	}

	el.Type = "picture"

	// log success
	logger.Success(r, m.App.Session,
		fmt.Sprintf("newsletter picture in element %s successfully deleted", eID))

	return rendering(w, r, newsletter.Element(el, []newsletter.ElementError{}))
}

func (m *Repository) HandleNewsletterPreview(w http.ResponseWriter, r *http.Request) error {
	// get newsletter id from request url
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		return fmt.Errorf("failed convert string to int: %w", err)
	}

	// get newsletter
	nl, err := m.DB.GetNewsletterByID(id)
	if err != nil {
		return fmt.Errorf("failed get newsletter: %w", err)
	}

	// get elements for newsletter
	elements, err := m.DB.GetAllNewsletterElementsByID(id)
	if err != nil {
		return fmt.Errorf("failed get elements: %w", err)
	}

	nl.Elements = elements

	var b bytes.Buffer
	var node *html.Node

	node, err = html.Parse(bytes.NewBufferString("<html>"))
	if err != nil {
		return fmt.Errorf("failed making bytes of template: %w", err)
	}

	// create template
	template, err := mailtemplates.Newsletter(nl)

	b = *bytes.NewBuffer([]byte(template))
	if err != nil {
		return fmt.Errorf("failed creating buffer from template: %w", err)
	}

	html.Render(&b, node)

	return rendering(w, r, newsletter.NewsletterPreview(b.String()))
}

func (m *Repository) HandleNewsletterSendList(w http.ResponseWriter, r *http.Request) error {
	// get newsletter id from request url
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		return fmt.Errorf("failed to convert string to int: %w", err)
	}

	// get newsletter so we know if its internal or not
	nl, err := m.DB.GetNewsletterByID(id)
	if err != nil {
		return fmt.Errorf("failed to get newsletter: %w", err)
	}

	// get recipients
	recipients, err := m.DB.NewsletterRecipients(nl.RecipientGroup)
	if err != nil {
		return fmt.Errorf("failed to get recipients: %w", err)
	}

	return rendering(w, r, newsletter.NewsletterPreSend(len(recipients), nl))
}

// HandleNewsletterSend handles the start of sending the newsletters and renders the progress page
func (m *Repository) HandleNewsletterSend(w http.ResponseWriter, r *http.Request) error {
	// get newsletter id from request url
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		return fmt.Errorf("failed to convert string to int: %w", err)
	}

	// send newsletters in seperate goroutine
	go m.sendNewsletters(id)

	return rendering(w, r, newsletter.NewsletterProgressPage())
}

// sendNewsletters sends the newsletter to the mail channel and the channel for the frontend visualization
func (m *Repository) sendNewsletters(id int) {
	// get newsletter so we know if its internal or not
	nl, err := m.DB.GetNewsletterByID(id)
	if err != nil {
		slog.Error(fmt.Errorf("failed to get newsletter: %w", err).Error())
		return
	}

	// get elements for newsletter
	elements, err := m.DB.GetAllNewsletterElementsByID(id)
	if err != nil {
		slog.Error(fmt.Errorf("failed to get newsletter elements: %w", err).Error())
		return
	}

	nl.Elements = elements

	// get recipients
	recipients, err := m.DB.NewsletterRecipients(nl.RecipientGroup)
	if err != nil {
		slog.Error(fmt.Errorf("failed to get recipients: %w", err).Error())
		return
	}

	// Open new Channel to send data to the template.
	slog.Info(fmt.Sprintf("open new channel for newsletter: %s", nl.Title))
	m.App.Newsletter.WaitList = make(chan models.MailData, len(recipients))
	m.App.Newsletter.Active = true

	// clean up after being done
	defer func() {
		slog.Info(fmt.Sprintf("close channel for newsletter: %s", nl.Title))
		m.App.Newsletter.Active = false
		close(m.App.Newsletter.WaitList)
	}()

	// create template
	template, err := mailtemplates.Newsletter(nl)
	if err != nil {
		slog.Error(fmt.Errorf("failed to create template: %w", err).Error())
		return
	}

	for i, r := range recipients {

		// replace per user strings for recipient groups
		// unsubscribe if its a newsletter
		vTemplate := template
		if nl.RecipientGroup == "newsletter" {
			// string to replace with
			unsubscribeString := r.ID + "?registered=" + strconv.FormatBool(len(r.City) > 0)

			// replace string
			vTemplate = strings.Replace(template, "[%unsubscribe%]", unsubscribeString, 1)
		}

		// create model for mail
		msg := models.MailData{
			To:      r.Email,
			Subject: nl.Title,
			Content: vTemplate,
		}

		// send to channel
		m.App.MailChan <- msg

		// put info in the channel to be read in frontend
		m.App.Newsletter.WaitList <- models.MailData{
			Reason:     "Newsletter",
			Count:      len(recipients),
			Progress:   i + 1,
			FirstName:  r.FirstName,
			FamilyName: r.FamilyName,
			To:         r.Email,
		}
		time.Sleep(time.Millisecond * 500)
	}
}

// HandleNewsletterProgressBar handles the SSE channel to send an updated progress bar for the sending of newsletters
func (m *Repository) HandleNewsletterProgressBar(w http.ResponseWriter, r *http.Request) {
	// Set headers for SSE
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	// Create a context for handling client disconnection
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	// Send data to the client
	for data := range m.App.Newsletter.WaitList {
		select {
		case <-r.Context().Done():
			// Quit early if the client is no longer connected.
			return
		default:
			template, err := templ.ToGoHTML(ctx, newsletter.NewsletterProgressBar(data))
			if err != nil {
				slog.Error(fmt.Errorf("failed to render template: %w", err).Error())
			}

			fmt.Fprintf(w, "data: %s\n\n", template)
			w.(http.Flusher).Flush()
		}
	}

	// send event to stop listening
	fmt.Fprint(w, "event: stop\ndata: stop\n\n")
	w.(http.Flusher).Flush()
}

// HandleNewsletterUnsubscribe the unsubscribe from newsletters
func (m *Repository) HandleNewsletterUnsubscribe(w http.ResponseWriter, r *http.Request) error {
	// get id
	id := chi.URLParam(r, "id")

	// get if registered or not
	memberParam := r.URL.Query().Get("registered")
	mem, err := strconv.ParseBool(memberParam)
	if err != nil {
		return fmt.Errorf("failed to parse bool: %w", err)
	}

	// remove recipient from db or set false for newsletter
	recipient, err := m.DB.NewsletterUnsubscribe(id, mem)
	if err != nil {
		return fmt.Errorf("to unsubscribe recipient with id %s from newsletter: %w", id, err)
	}

	// log success
	logger.Success(r, m.App.Session,
		fmt.Sprintf("Recipient with id %s successfully unsubscribed from newsletter", id))

	return rendering(w, r, newsletter.Unsubscribe(recipient))
}

// HandleNewsletterSubscribeIndex sends a form for newsletter subscriptions
func (m *Repository) HandleNewsletterSubscribeIndex(w http.ResponseWriter, r *http.Request) error {

	// send back an empty form
	return rendering(w, r, newsletter.SubscriptionForm(newsletter.RecipientForm{}, newsletter.RecipientFormError{}))
}

// HandleNewsletterSubscribe handles the newsletter subscription
func (m *Repository) HandleNewsletterSubscribe(w http.ResponseWriter, r *http.Request) error {

	// parse form
	r.ParseForm()

	var recipient newsletter.RecipientForm
	recipient.FirstName = r.Form.Get("first_name")
	recipient.FamilyName = r.Form.Get("family_name")
	recipient.Email = r.Form.Get("email")

	// validation
	form := forms.New(r.PostForm)
	form.Required("first_name", "family_name", "email")
	form.IsEmail("email")

	// UniqueMail checks if the mail address is unique
	if !form.IsEmpty("email") {
		validmail, _ := m.DB.CheckSubscriptionUnique(recipient.Email)
		if !validmail {
			form.Errors.Add("email", "Email Adresse wurde schon verwendet")
		}
	}

	// what to do if form not valid
	if !form.Valid() {
		return rendering(w, r, newsletter.SubscriptionForm(recipient, newsletter.RecipientFormError{
			FirstName:  form.Errors.Get("first_name"),
			FamilyName: form.Errors.Get("family_name"),
			Email:      form.Errors.Get("email"),
		}))
	}

	err := m.DB.NewsletterSubscribe(recipient)
	if err != nil {
		return err
	}

	// log success
	logger.Success(r, m.App.Session,
		fmt.Sprintf("%s %s successfully subscribed to newsletter", recipient.FirstName, recipient.FamilyName))

	// send back success message
	return rendering(w, r, newsletter.SubscriptionSuccess(recipient))
}
