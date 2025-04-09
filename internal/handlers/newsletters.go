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
	"log"
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
	"github.com/faebut/prozessor_gostack/internal/mailtemplates"
	"github.com/faebut/prozessor_gostack/internal/models"
	"github.com/faebut/prozessor_gostack/view/newsletter"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/nfnt/resize"
	"golang.org/x/net/html"
)

// HandleNewsletterIndex is the newsletter page handler
func (m *Repository) HandleNewsletterIndex(w http.ResponseWriter, r *http.Request) error {
	// get visitor
	visitor, ok := m.App.Session.Get(r.Context(), "visitor").(models.Visitor)
	if !ok {
		log.Println("could not get visitor from Session")
		return hxRedirect(w, r, "/admin/logout")
	}
	return rendering(w, r, newsletter.Newsletters(visitor))
}

// HandleNewsletterList gets and returns the unfiltered newsletter list
func (m *Repository) HandleNewsletterList(w http.ResponseWriter, r *http.Request) error {

	newsletterList, _, err := m.DB.GetAllNewsletters()
	if err != nil {
		log.Println("could not get newsletters from DB: ", err)

		m.App.Session.Put(r.Context(), "error", "Can't get badges from database")
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
		log.Println("could not get templates from DB: ", err)
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
		return err
	}
	element.ID = uuid.String()
	element.Type = elType

	return rendering(w, r, newsletter.Element(element, []newsletter.ElementError{}))
}

func (m *Repository) HandleNewsletterEventsLoad(w http.ResponseWriter, r *http.Request) error {
	// get events
	events, err := events.GetEvents()
	if err != nil {
		return err
	}

	// put events into string
	eventString, err := json.Marshal(events)
	if err != nil {
		return err
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
			log.Println("Error converting string to int:", err)
			return err
		}
		nl.ID = id
		edit = true
	}

	templates, err := m.DB.GetNewsletterTemplates()
	if err != nil {
		log.Println("could not get templates from DB: ", err)
	}

	// parse form
	err = r.ParseMultipartForm(5 << 20)
	if err != nil {
		log.Println("Error parsing multipart form: ", err)
		return nil
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
					return err
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

			log.Println("there is a picture element")

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
					log.Println(err)
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
						log.Println("filetype", filetype)
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
	nl.Members = form.BoxValue("members")
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
			log.Println("got a picture")
			file, fileHeader, err := r.FormFile(e.ID + "_fileupload")
			if err != nil {
				log.Println("error handling file")
				return err
				// TODO: how to else handle error
			} else {
				defer file.Close()

				log.Println("handling picture")

				// Generate a unique filename to prevent overwriting and conflicts
				uuid, err := uuid.NewRandom()
				if err != nil {
					log.Println("Error generating uuid")
					return err
					// TODO: how to else handle error
				}
				// Append the filetype extension to uuid to create filename
				filename := uuid.String() + filepath.Ext(fileHeader.Filename)

				log.Println("made filename:", filename)

				// TODO: maybe don't save on db server but on webserver

				// check if uploads directory exists, if not create it
				if _, err := os.Stat("static/newsletter"); err != nil {
					err = os.Mkdir("static/newsletter/", os.ModePerm)
					if err != nil {
						log.Println("Error generating folder")
						return err
						// TODO: how to else handle error
					}
				}

				log.Println("filepath made")

				// Create the full path for saving the file
				filePath := filepath.Join("static/newsletter", filename)

				log.Println("filepath joined")

				// Save the file to the server
				dst, err := os.Create(filePath)
				if err != nil {
					log.Println("error saving file", err)
					return err

					// TODO: how to else handle error

				}

				log.Println("created file")

				defer dst.Close()

				log.Println("handling types")
				// convert to image
				var imgFile image.Image
				if filetype == "jpeg" {
					imgFile, err = jpeg.Decode(file)
					if err != nil {
						log.Println(err)
						return err
					}
				} else if filetype == "png" {
					imgFile, err = png.Decode(file)
					if err != nil {
						log.Println(err)
						return err
					}
				}

				log.Println("prepare to resize")

				// resize image
				resFile := resize.Resize(600, 0, imgFile, resize.Lanczos3)

				log.Println("resized, encoding back")

				// write image
				if filetype == "jpeg" {
					err = jpeg.Encode(dst, resFile, nil)
					if err != nil {
						log.Println(err)
						return err
					}
				} else if filetype == "png" {
					err = png.Encode(dst, resFile)
					if err != nil {
						log.Println(err)
						return err
					}
				}

				log.Println("encoded")

				// Update the element with the update. Need to use index
				// here to modify the struct directly
				elements[i].Upload = filename
			}
		}
	}

	// what to do if form valid
	if formErrors.Edit {
		err = m.DB.UpdateNewsletter(nl)
		if err != nil {
			log.Println("ERROR:", err)
			return err
		}
	} else {
		err = m.DB.CreateNewsletter(nl)
		if err != nil {
			log.Println("ERROR:", err)
			return err
		}
	}

	// TODO: do a better workflow
	// if all is well, send to the newsletter overview page
	redirectURL := "/admin/newsletter"

	return hxRedirect(w, r, redirectURL)
}

// HandleNewsletterEditIndex returns the form for an existing newsletter
func (m *Repository) HandleNewsletterEditIndex(w http.ResponseWriter, r *http.Request) error {
	nlID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		log.Println("Error converting newsletter id to int", err)
		return err
	}

	// get newsletter from database
	nl, err := m.DB.GetNewsletterByID(nlID)
	if err != nil {
		log.Println("could not get newsletter from DB: ", err)
	}

	// get elements for newsletter
	elements, err := m.DB.GetAllNewsletterElementsByID(nl.ID)
	if err != nil {
		log.Println("could not get newsletter elements from DB: ", err)
	}

	nl.Elements = elements

	templates, err := m.DB.GetNewsletterTemplates()
	if err != nil {
		log.Println("could not get templates from DB: ", err)
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
		return err
	}

	// delete image in database and get element
	el, err := m.DB.DeleteNewsletterImage(eID)
	if err != nil {
		return err
	}

	el.Type = "picture"

	return rendering(w, r, newsletter.Element(el, []newsletter.ElementError{}))
}

func (m *Repository) HandleNewsletterPreview(w http.ResponseWriter, r *http.Request) error {
	// get newsletter id from request url
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		return err
	}

	// get newsletter
	nl, err := m.DB.GetNewsletterByID(id)
	if err != nil {
		return err
	}

	// get elements for newsletter
	elements, err := m.DB.GetAllNewsletterElementsByID(id)
	if err != nil {
		return err
	}

	nl.Elements = elements

	var b bytes.Buffer
	var node *html.Node

	node, err = html.Parse(bytes.NewBufferString("<html>"))
	if err != nil {
		log.Println("making bytes of template:", err)
		return err
	}

	// create template
	template, err := mailtemplates.Newsletter(nl)

	b = *bytes.NewBuffer([]byte(template))
	if err != nil {
		log.Println("error creating buffer from template:", err)
		return err
	}

	html.Render(&b, node)

	return rendering(w, r, newsletter.NewsletterPreview(b.String()))
}

func (m *Repository) HandleNewsletterSendList(w http.ResponseWriter, r *http.Request) error {
	// get newsletter id from request url
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		return err
	}

	// get newsletter so we know if its internal or not
	nl, err := m.DB.GetNewsletterByID(id)
	if err != nil {
		return err
	}

	// get recipients
	recipients, err := m.DB.NewsletterRecipients(nl.Members)
	if err != nil {
		return err
	}

	return rendering(w, r, newsletter.NewsletterPreSend(len(recipients), nl))
}

// HandleNewsletterSend handles the start of sending the newsletters and renders the progress page
func (m *Repository) HandleNewsletterSend(w http.ResponseWriter, r *http.Request) error {
	// get newsletter id from request url
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		return err
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
		// TODO: error handling
		log.Println(err)
	}

	// get elements for newsletter
	elements, err := m.DB.GetAllNewsletterElementsByID(id)
	if err != nil {
		// TODO: error handling
		log.Println(err)
	}

	nl.Elements = elements

	// get recipients
	recipients, err := m.DB.NewsletterRecipients(nl.Members)
	if err != nil {
		// TODO: error handling
		log.Println(err)
	}

	// Open new Channel to send data to the template.
	log.Println("open new channel")
	m.App.Newsletter.WaitList = make(chan models.MailData, len(recipients))
	m.App.Newsletter.Active = true

	// create template
	template, err := mailtemplates.Newsletter(nl)
	if err != nil {
		log.Println("error creating template:", err)
		return
	}

	for i, r := range recipients {

		// replace per user strings
		unsubscribeString := r.ID + "?registered=" + strconv.FormatBool(len(r.City) > 0)
		vTemplate := strings.Replace(template, "[%unsubscribe%]", unsubscribeString, 1)

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

	log.Println("closing channel")
	m.App.Newsletter.Active = false
	close(m.App.Newsletter.WaitList)
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

// HandleNewsletterUnsubscribe the unsubscribe from newsletters
func (m *Repository) HandleNewsletterUnsubscribe(w http.ResponseWriter, r *http.Request) error {
	// get id
	id := chi.URLParam(r, "id")

	// get if registered or not
	memberParam := r.URL.Query().Get("registered")
	mem, err := strconv.ParseBool(memberParam)
	if err != nil {
		return err
	}

	// remove recipient from db or set false for newsletter
	recipient, err := m.DB.NewsletterUnsubscribe(id, mem)
	if err != nil {
		return err
	}

	return rendering(w, r, newsletter.Unsubscribe(recipient))
}

// HandleNewsletterSubscribeIndex sends a form for newsletter subscriptions
func (m *Repository) HandleNewsletterSubscribeIndex(w http.ResponseWriter, r *http.Request) error {

	// send back an empty form
	return rendering(w, r, newsletter.SubscriptionForm(newsletter.RecipientForm{}))
}

// HandleNewsletterSubscribe handles the newsletter subscription
func (m *Repository) HandleNewsletterSubscribe(w http.ResponseWriter, r *http.Request) error {

	// parse form
	r.ParseForm()

	var recipient newsletter.RecipientForm
	recipient.FirstName = r.Form.Get("first_name")
	recipient.FamilyName = r.Form.Get("family_name")
	recipient.Email = r.Form.Get("email")

	log.Println("subscribing", recipient.FirstName)

	// send back success message
	return rendering(w, r, newsletter.SubscriptionSuccess(recipient))
}
