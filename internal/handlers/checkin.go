package handlers

import (
	"fmt"
	"log"
	"math"
	"net/http"
	"strconv"

	"github.com/faebut/prozessor_gostack/internal/models"
	"github.com/faebut/prozessor_gostack/view/checkin"
)

// HandleCheckinIndex is the check-in page handler
func (m *Repository) HandleCheckinIndex(w http.ResponseWriter, r *http.Request) error {
	visitor, ok := m.App.Session.Get(r.Context(), "visitor").(models.Visitor)
	if !ok {
		log.Println("could not get visitor from Session")
		m.App.Session.Put(r.Context(), "error", "Can't get visitor from session")

		return hxRedirect(w, r, "/admin/logout")
	}

	// reset filter and search and go to page 1
	page := 1
	filter := models.NewVisitorSearch("")

	visitors, count, err := m.DB.CheckinAllVisitors(page, filter)
	if err != nil {
		log.Println(err)
		return err
	}

	return rendering(w, r, checkin.Checkin(
		visitor,
		visitors,
		count,
		int(math.Ceil(float64(count)/10)),
		page,
		"",
	))
}

// HandleCheckinList is the handler to get the visitors list
func (m *Repository) HandleCheckinList(w http.ResponseWriter, r *http.Request) error {

	// set initial params
	page := 1
	filter := models.NewVisitorSearch("")
	if r.URL.Query().Has("search") {
		filter = models.NewVisitorSearch(r.URL.Query().Get("search"))
	}

	// check if a page param has been submitted and change to this value
	if r.URL.Query().Has("page") {
		page, _ = strconv.Atoi(r.URL.Query().Get("page"))
	}

	visitors, count, err := m.DB.CheckinAllVisitors(page, filter)
	if err != nil {
		log.Println("could not get visitors from DB: ", err)
		return err
	}

	return rendering(w, r, checkin.VisitorList(
		visitors,
		count,
		int(math.Ceil(float64(count)/10)),
		page,
		fmt.Sprintf("&search=%s", filter.SearchString),
	))
}

// HandleCheckinListFilter is the handler to get the visitors list with a new filter
func (m *Repository) HandleCheckinListFilter(w http.ResponseWriter, r *http.Request) error {

	// set initial params
	page := 1

	// check if a page param has been submitted and change to this value
	// if r.URL.Query().Has("page") {
	// 	page, _ = strconv.Atoi(r.URL.Query().Get("page"))
	// }

	// create new filter
	searchTerm := r.Form.Get("search_term")
	filter := models.NewVisitorSearch(searchTerm)

	visitors, count, err := m.DB.CheckinAllVisitors(page, filter)
	if err != nil {
		log.Println("could not get visitors from DB: ", err)
		return err
	}

	return rendering(w, r, checkin.VisitorList(
		visitors,
		count,
		int(math.Ceil(float64(count)/10)),
		page,
		fmt.Sprintf("&search=%s", filter.SearchString),
	))
}
