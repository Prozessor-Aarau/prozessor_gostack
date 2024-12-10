package handlers

import (
	"fmt"
	"log"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/faebut/prozessor_gostack/internal/models"
	"github.com/faebut/prozessor_gostack/view/checkin"
	"github.com/go-chi/chi/v5"
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
	visits, vCount, err := m.DB.GetVisitsByDate(time.Now())
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
		visits,
		vCount,
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

// HandleCheckinFormIndex gets the checkin form for a visitor
func (m *Repository) HandleCheckinFormIndex(w http.ResponseWriter, r *http.Request) error {
	// get params, userID for visitor and vid for visit
	userID := chi.URLParam(r, "id")
	vid, err := strconv.Atoi(chi.URLParam(r, "vid"))
	if err != nil {
		vid = 0
	}
	// set edit mode
	edit := vid > 0

	var visit models.Visit
	checkedAteliers := []int{}
	if edit {
		visit, err = m.DB.GetVisitByID(vid)
		if err != nil {
			log.Println("Error getting visit from DB:", err)
			return err
		}
		// pushing all atelier IDs that are in visit to checkedAteliers slice
		for _, v := range visit.Ateliers {
			checkedAteliers = append(checkedAteliers, v.ID)
		}
	}

	visitor, err := m.DB.VisitorByID(userID)
	if err != nil {
		log.Println("could not get visitor from DB: ", err)
		return err
	}
	ateliers, err := m.DB.GetAllAteliers()
	if err != nil {
		log.Println("could not get ateliers from DB: ", err)
		return err
	}

	return rendering(w, r, checkin.CheckinForm(
		edit,
		visitor,
		ateliers,
		visit,
		checkedAteliers,
		"",
	))
}

// HandleCheckinForm handles the post request from the checkin form
func (m *Repository) HandleCheckinForm(w http.ResponseWriter, r *http.Request) error {
	userID := chi.URLParam(r, "id")
	vid, err := strconv.Atoi(chi.URLParam(r, "vid"))
	if err != nil {
		vid = 0
	}
	// set edit mode
	edit := vid > 0

	var visit models.Visit
	if edit {
		visit, err = m.DB.GetVisitByID(vid)
		if err != nil {
			log.Println("Error getting visit from DB:", err)
			return err
		}
	}

	checkedAteliers := make([]int, 0, len(r.PostForm))
	for a := range r.PostForm {
		kInt, err := strconv.Atoi(a)
		if err == nil {
			checkedAteliers = append(checkedAteliers, kInt)
		}
	}

	// check if ateliers are chosen
	var formError string
	if len(checkedAteliers) < 1 {
		formError = "Mindestens ein Atelier auswählen"
	}

	// what to do if form not valid
	if len(formError) > 0 {

		visitor, err := m.DB.VisitorByID(userID)
		if err != nil {
			log.Println("could not get visitor from DB: ", err)
			return err
		}
		ateliers, err := m.DB.GetAllAteliers()
		if err != nil {
			log.Println("could not get ateliers from DB: ", err)
			return err
		}

		// send the data to the template
		return rendering(w, r, checkin.CheckinForm(edit, visitor, ateliers, visit, checkedAteliers, formError))
	}

	// post visit to database
	visit.Visitor.ID = userID
	ressortID, err := strconv.Atoi(r.Form.Get("ressort"))
	if err != nil {
		log.Println(err)
	}
	if ressortID > 0 {
		visit.Ressort.ID = ressortID
	}
	if r.Form.Has("partner") {
		partnerID, _ := strconv.Atoi(r.Form.Get("partner"))
		if partnerID > 0 {
			visit.Partner.ID = partnerID
		} else if visit.Partner.ID > 0 {
			visit.Partner.ID = 0
		}
	}

	if edit {
		err = m.DB.UpdateVisit(visit, checkedAteliers, vid)
		if err != nil {
			log.Println(err)
		}
	} else {
		err = m.DB.PostVisit(visit, checkedAteliers)
		if err != nil {
			log.Println(err)
		}
	}

	// redirect
	return hxRedirect(w, r, "/admin/checkin")
}

// HandleVisitDelete deletes a visit by id
func (m *Repository) HandleCheckinDelete(w http.ResponseWriter, r *http.Request) error {
	visitId, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		log.Println(err)
	}

	err = m.DB.DeleteVisit(visitId)
	if err != nil {
		log.Println(err)
		return err
	}

	// get new visitorlist
	return hxRedirect(w, r, "/admin/checkin")
}
