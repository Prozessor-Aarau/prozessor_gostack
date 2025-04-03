package handlers

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/faebut/prozessor_gostack/internal/events"
	"github.com/faebut/prozessor_gostack/internal/models"
	"github.com/faebut/prozessor_gostack/view/newsletter"
	"github.com/google/uuid"
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

	return rendering(w, r, newsletter.Element(element))
}

// TODO: get real data from frontmatter parser
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

// TODO: get real data from newsletter element
func (m *Repository) HandleNewsletterEventsShow(w http.ResponseWriter, r *http.Request) error {

	var events []models.Event

	jsonString := `[{"title":"Ein Event","date":"2025-05-12T13:00:00+02:00","url":"https://www.eventurl.ch/urlibus"},{"title":"Zwei Event","date":"2025-05-22T08:00:00+02:00","url":"https://www.eventurl.ch/urlibus"},{"title":"Noch ein Event","date":"2025-06-02T17:00:00+02:00","url":"https://www.eventurl.ch/urlibus"},{"title":"Letzter Event","date":"2025-04-30T13:00:00+02:00","url":"https://www.eventurl.ch/urlibus"}]`

	err := json.Unmarshal([]byte(jsonString), &events)
	if err != nil {
		panic(err)
	}

	return rendering(w, r, newsletter.EventList(events, jsonString))
}

// HandleNewsletterCreateIndex returns the form for a new newsletter
func (m *Repository) HandleNewsletterCreate(w http.ResponseWriter, r *http.Request) error {
	// create new empty structs
	var nl models.Newsletter

	// var e newsletter.NewsletterErrors
	//
	// templates, err := m.DB.GetNewsletterTemplates()
	// if err != nil {
	// 	log.Println("could not get templates from DB: ", err)
	// }
	//

	// parse form
	r.ParseForm()

	// handle elements
	// TODO: handle file upload
	var element_id []string
	// var element_position []int
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

		// if key == "element_position" {
		// 	for _, val := range values {
		// 		position, err := strconv.Atoi(val)
		// 		if err != nil {
		// 			return err
		// 		}
		// 		element_position = append(element_position, position)
		// 	}
		// }

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
	var elements []models.NewsletterElement

	for i := range element_id {
		element := models.NewsletterElement{
			ID: element_id[i],
			// Position: element_position[i],
			Type:    element_type[i],
			Content: element_content[i],
			AltText: element_alttext[i],
			Url:     element_url[i],
		}

		elements = append(elements, element)
	}

	nl.Elements = elements

	// handle other fields
	nl.Title = r.Form.Get("title")
	nl.Description = r.Form.Get("description")
	nl.Template.ID, _ = strconv.Atoi(r.Form.Get("template"))
	nl.ColorBase = r.Form.Get("color_base")
	nl.ColorAccent = r.Form.Get("color_accent")

	log.Println("Newsletter:", nl)

	// render template
	// return rendering(w, r, newsletter.NewsletterCreate(nl, templates, elements, e))
	return nil
}
