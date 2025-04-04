package handlers

import (
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

	"github.com/faebut/prozessor_gostack/internal/events"
	"github.com/faebut/prozessor_gostack/internal/forms"
	"github.com/faebut/prozessor_gostack/internal/models"
	"github.com/faebut/prozessor_gostack/view/newsletter"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/nfnt/resize"
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
	log.Println("get events")
	events, err := events.GetEvents()
	if err != nil {
		return err
	}
	log.Printf("Events: %v", events)

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

	for _, e := range elements {
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
	// Retrieve the file from form data
	for i, e := range elements {
		if len(e.AltText) > 0 {
			file, fileHeader, err := r.FormFile(e.ID + "_fileupload")
			if err != nil {
				log.Println("error handling file")
				// TODO: how to else handle error
			} else {
				defer file.Close()

				// Generate a unique filename to prevent overwriting and conflicts
				uuid, err := uuid.NewRandom()
				if err != nil {
					log.Println("Error generating uuid")
					// TODO: how to else handle error
				}
				// Append the filetype extension to uuid to create filename
				filename := uuid.String() + filepath.Ext(fileHeader.Filename)

				// TODO: maybe don't save on db server but on webserver

				// check if uploads directory exists, if not create it
				if _, err := os.Stat("static/newsletter"); err != nil {
					os.Mkdir("static/newsletter/", os.ModePerm)
				}

				// Create the full path for saving the file
				filePath := filepath.Join("static/newsletter", filename)

				// Save the file to the server
				dst, err := os.Create(filePath)
				if err != nil {
					log.Println("error saving file", err)

					// TODO: how to else handle error

				}
				defer dst.Close()

				// convert to image
				var imgFile image.Image
				if filetype == "jpeg" {
					imgFile, err = jpeg.Decode(file)
					if err != nil {
						log.Println(err)
					}
				} else if filetype == "png" {
					imgFile, err = png.Decode(file)
					if err != nil {
						log.Println(err)
					}
				}

				// resize image
				resFile := resize.Resize(600, 0, imgFile, resize.Lanczos3)

				// write image
				if filetype == "jpeg" {
					jpeg.Encode(dst, resFile, nil)
				} else if filetype == "png" {
					png.Encode(dst, resFile)
				}

				// TODO: handle old files, if there has been a change
				// for this we need to check, if there was a saved image
				// before and delete them
				// INFO: handle this, when edit is to be implemented

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
		}
	} else {
		err = m.DB.CreateNewsletter(nl)
		if err != nil {
			log.Println("ERROR:", err)
		}
	}

	// TODO: go on
	// if all is well, send to the preview page next

	return hxRedirect(w, r, "/admin/newsletter")
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
