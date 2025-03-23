package handlers

import (
	"log"
	"net/http"

	"github.com/faebut/prozessor_gostack/internal/models"
	"github.com/faebut/prozessor_gostack/view/newsletter"
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

func (m *Repository) HandleNewsletterCreateIndex(w http.ResponseWriter, r *http.Request) error {
	// create new empty structs
	var nl models.Newsletter
	var e newsletter.NewsletterErrors

	templates, err := m.DB.GetNewsletterTemplates()
	if err != nil {
		log.Println("could not get templates from DB: ", err)
	}

	// render template
	return rendering(w, r, newsletter.NewsletterCreate(nl, templates, e))
}
