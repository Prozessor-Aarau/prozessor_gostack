package handlers

import (
	"log"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/faebut/prozessor_gostack/internal/forms"
	"github.com/faebut/prozessor_gostack/internal/models"
	"github.com/faebut/prozessor_gostack/internal/render"
	"github.com/go-chi/chi/v5"
)

// UpdateProfile updates the visitors profile
func (m *Repository) GetVisitsByDate(w http.ResponseWriter, r *http.Request) {
	today := time.Now()

	visitList, count, err := m.DB.GetVisitsByDate(today)
	if err != nil {
		log.Println("could not get visits from DB: ", err)
		m.App.Session.Put(r.Context(), "error", "Can't get visitors from database")
	}

	data := make(map[string]interface{})
	data["visits"] = visitList
	data["count"] = count

	render.Template(w, r, "part_visitslist.html", &models.TemplateData{
		Data: data,
	})
}

// CheckinVisitorForm gets the Checkin Form for a visitor
func (m *Repository) CheckinVisitorForm(w http.ResponseWriter, r *http.Request) {
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
		}
		// pushing all atelier IDs that are in visit to checkedAteliers slice
		for _, v := range visit.Ateliers {
			checkedAteliers = append(checkedAteliers, v.ID)
		}
	}

	visitor, err := m.DB.VisitorByID(userID)
	if err != nil {
		log.Println("could not get visitor from DB: ", err)
		m.App.Session.Put(r.Context(), "error", "Can't get visitors from database")
	}
	ateliers, err := m.DB.GetAllAteliers()
	if err != nil {
		log.Println("could not get ateliers from DB: ", err)
		m.App.Session.Put(r.Context(), "error", "Can't get ateliers from database")
	}

	data := make(map[string]interface{})
	data["visitor"] = visitor
	data["ateliers"] = ateliers
	data["ateliers_checked"] = checkedAteliers
	data["edit"] = edit
	data["visit"] = visit

	render.Template(w, r, "part_checkinform.html", &models.TemplateData{
		Data: data,
		Form: forms.New(nil),
	})
}

// PostCheckin posts the Checkin Form for a visitor
func (m *Repository) PostCheckin(w http.ResponseWriter, r *http.Request) {
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
		}
	}

	form := forms.New(r.PostForm)
	checkedAteliers := make([]int, 0, len(r.PostForm))
	for a := range r.PostForm {
		kInt, err := strconv.Atoi(a)
		if err == nil {
			checkedAteliers = append(checkedAteliers, kInt)
		}
	}

	if len(checkedAteliers) < 1 {
		form.Errors.Add("ateliers", "Mindestens ein Atelier auswählen")
	}

	// what to do if form not valid
	if !form.Valid() {

		visitor, err := m.DB.VisitorByID(userID)
		if err != nil {
			log.Println("could not get visitor from DB: ", err)
			m.App.Session.Put(r.Context(), "error", "Can't get visitors from database")
		}
		ateliers, err := m.DB.GetAllAteliers()
		if err != nil {
			log.Println("could not get ateliers from DB: ", err)
			m.App.Session.Put(r.Context(), "error", "Can't get ateliers from database")
		}

		data := make(map[string]interface{})
		data["visitor"] = visitor
		data["ateliers"] = ateliers
		data["ateliers_checked"] = checkedAteliers
		data["edit"] = edit
		data["visit"] = visit

		// send the data to the template
		w.Header().Set("HX-Target", "#addVisit")
		w.WriteHeader(http.StatusSeeOther)
		render.Template(w, r, "part_checkinform.html", &models.TemplateData{
			Data: data,
			Form: form,
		})
		return
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
	w.Header().Set("HX-Redirect", "/admin/checkin")
	w.WriteHeader(http.StatusSeeOther)
}

func (m *Repository) DeleteCheckin(w http.ResponseWriter, r *http.Request) {
	visitId, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		log.Println(err)
	}

	err = m.DB.DeleteVisit(visitId)
	if err != nil {
		log.Println(err)
		return
	}

	// TODO: find a more elegant way without reloading the whole page
	// get new visitorlist
	w.Header().Set("HX-Redirect", "/admin/checkin")
	w.WriteHeader(http.StatusSeeOther)
}
