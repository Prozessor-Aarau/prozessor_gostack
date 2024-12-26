package handlers

import (
	"fmt"
	"log"
	"math"
	"net/http"
	"strconv"
	"strings"

	"github.com/faebut/prozessor_gostack/internal/forms"
	"github.com/faebut/prozessor_gostack/internal/models"
	"github.com/faebut/prozessor_gostack/view/visitors"
)

// HandleVisitorsIndex is the visitors page handler
func (m *Repository) HandleVisitorsIndex(w http.ResponseWriter, r *http.Request) error {
	visitor, ok := m.App.Session.Get(r.Context(), "visitor").(models.Visitor)
	if !ok {
		log.Println("could not get visitor from Session")
		m.App.Session.Put(r.Context(), "error", "Can't get visitor from session")

		return hxRedirect(w, r, "/admin/logout")
	}

	return rendering(w, r, visitors.Visitors(visitor))
}

// HandleVisitorList is the handler to get the visitors list
func (m *Repository) HandleVisitorList(w http.ResponseWriter, r *http.Request) error {

	// set initial params
	page := 1
	limit := 10
	filter := models.NewVisitorSearch("")
	if r.URL.Query().Has("search") {
		filter = models.NewVisitorSearch(r.URL.Query().Get("search"))
	}
	filter.Sort = "family_name"
	filter.SortDir = "ASC"

	// check if a page param has been submitted and change to this value
	if r.URL.Query().Has("page") {
		page, _ = strconv.Atoi(r.URL.Query().Get("page"))
	}
	if r.URL.Query().Has("sort") {
		filter.Sort = r.URL.Query().Get("sort")
	}
	if r.URL.Query().Has("sortdir") {
		filter.SortDir = strings.ToUpper(r.URL.Query().Get("sortdir"))
	}
	if r.URL.Query().Has("abo") {
		filter.Abo, _ = strconv.ParseBool(r.URL.Query().Get("abo"))
	}
	if r.URL.Query().Has("valid") {
		filter.AboValidity = r.URL.Query().Get("valid")
	}
	if r.URL.Query().Has("member") {
		filter.Member, _ = strconv.ParseBool(r.URL.Query().Get("member"))
	}
	if r.URL.Query().Has("partner") {
		filter.Partner, _ = strconv.ParseBool(r.URL.Query().Get("partner"))
	}
	if r.URL.Query().Has("badge") {
		filter.Badge, _ = strconv.ParseBool(r.URL.Query().Get("badge"))
	}

	visitorList, count, err := m.DB.AllVisitors(page, limit, filter)
	if err != nil {
		log.Println("could not get visitors from DB: ", err)
		return err
	}

	urlSearchString := fmt.Sprintf(
		"&search=%s&sort=%s&sortdir=%s",
		filter.SearchString,
		filter.Sort,
		filter.SortDir,
	)
	if filter.Abo {
		urlSearchString = urlSearchString + fmt.Sprintf("&abo=true&valid=%s", filter.AboValidity)
	}
	if filter.Member {
		urlSearchString = urlSearchString + "&member=true"
	}
	if filter.Partner {
		urlSearchString = urlSearchString + "&partner=true"
	}
	if filter.Badge {
		urlSearchString = urlSearchString + "&badge=true"
	}

	return rendering(w, r, visitors.VisitorList(
		visitorList,
		count,
		int(math.Ceil(float64(count)/10)),
		page,
		urlSearchString,
	))
}

// HandleVisitorListFilter gets the filter values and returns the filtered visitor list
func (m *Repository) HandleVisitorListFilter(w http.ResponseWriter, r *http.Request) error {

	// set initial page param
	page := 1
	limit := 10
	// check if a page param has been submitted and change to this value
	if r.URL.Query().Has("page") {
		page, _ = strconv.Atoi(r.URL.Query().Get("page"))
	}
	searchTerm := r.Form.Get("search_term")

	form := forms.New(r.PostForm)
	filter := models.NewVisitorSearch(searchTerm)
	filter.Sort = r.Form.Get("sort")
	filter.SortDir = r.Form.Get("sort_dir")
	filter.Abo = form.BoxValue("hasabo")
	filter.AboValidity = r.Form.Get("abovalidity")
	filter.Badge = form.BoxValue("badge")
	filter.Member = form.BoxValue("member")
	filter.Partner = form.BoxValue("partner")

	visitorList, count, err := m.DB.AllVisitors(page, limit, filter)
	if err != nil {
		log.Println("could not get visitors from DB: ", err)
		m.App.Session.Put(r.Context(), "error", "Can't get visitors from database")
		return err
	}

	urlSearchString := fmt.Sprintf(
		"&search=%s&sort=%s&sortdir=%s",
		filter.SearchString,
		filter.Sort,
		filter.SortDir,
	)
	if filter.Abo {
		urlSearchString = urlSearchString + fmt.Sprintf("&abo=true&valid=%s", filter.AboValidity)
	}
	if filter.Member {
		urlSearchString = urlSearchString + "&member=true"
	}
	if filter.Partner {
		urlSearchString = urlSearchString + "&partner=true"
	}
	if filter.Badge {
		urlSearchString = urlSearchString + "&badge=true"
	}

	return rendering(w, r, visitors.VisitorList(
		visitorList,
		count,
		int(math.Ceil(float64(count)/10)),
		page,
		urlSearchString,
	))
}
