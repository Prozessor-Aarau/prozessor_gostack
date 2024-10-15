package handlers

import (
	"fmt"
	"log"
	"math"
	"net/http"
	"strconv"

	"github.com/faebut/prozessor_gostack/internal/forms"
	"github.com/faebut/prozessor_gostack/internal/models"
	"github.com/faebut/prozessor_gostack/internal/render"
	"github.com/go-chi/chi/v5"
)

// Abonnements is the abonnements page handler
func (m *Repository) Abonnements(w http.ResponseWriter, r *http.Request) {
	user, ok := m.App.Session.Get(r.Context(), "visitor").(models.Visitor)
	if !ok {
		log.Println("could not get visitor from Session")
		m.App.Session.Put(r.Context(), "error", "Can't get visitor from session")
		http.Redirect(w, r, "/admin/logout", http.StatusSeeOther)
		return
	}

	data := make(map[string]interface{})
	data["user"] = user

	render.Template(w, r, "admin_abonnements.html", &models.TemplateData{
		Data: data,
	})
}

// AbonnementCreate is the abonnements creation page handler
func (m *Repository) AbonnementCreate(w http.ResponseWriter, r *http.Request) {
	var abonnement models.Abonnement

	data := make(map[string]interface{})
	data["user"] = abonnement

	render.Template(w, r, "part_abonnement_create.html", &models.TemplateData{
		Data: data,
		Form: forms.New(nil),
	})
}

// AbonnementCreate is the abonnements creation page handler
func (m *Repository) AbonnementCreatePrice(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	visitor, err := m.DB.VisitorByID(id)
	if err != nil {
		log.Println("could not get visitor from DB: ", err)
		m.App.Session.Put(r.Context(), "error", "Can't get visitors from database")
	}
	member := "Abonnement"
	price := 400

	if visitor.Ressort.ID > 0 {
		member = "Mitglied"
		price = 320
	}

	fmt.Fprintf(
		w,
		`<input type="hidden" name="price" value="%d" /><span>%d.- %s</span>`,
		price,
		price,
		member,
	)
}

// Badge is the handler for a single Badge
func (m *Repository) PostAbonnementCreate(w http.ResponseWriter, r *http.Request) {
	// create empty badge
	var abonnement models.Abonnement

	form := forms.New(r.Form)
	form.Required("price", "member", "buydate")

	abonnement.BuyDate = r.Form.Get("buydate")
	stringPrice, err := strconv.Atoi(r.Form.Get("price"))
	if err != nil {
		log.Println("error converting string to int:", err)
	}
	abonnement.Price = stringPrice
	abonnement.VisitorID = r.Form.Get("member")
	abonnement.PartialPayment = form.BoxValue("partial_payment")

	// set member with price
	if abonnement.Price == 320 {
		abonnement.Member = true
	} else {
		abonnement.Member = false
	}

	if !form.Valid() {
		data := make(map[string]interface{})
		data["abonnement"] = abonnement

		render.Template(w, r, "part_abonnement_create.html", &models.TemplateData{
			Data: data,
			Form: form,
		})
		return
	}

	id, err := m.DB.CreateAbonnement(abonnement)
	if err != nil {
		m.DB.CreateLog(
			models.Log{
				Typ:     "error",
				Module:  "abonnements",
				Message: fmt.Sprintf("Error creating abonnement: %s", err),
			},
		)
		log.Println("Error deleting badge:", err)
		return
	} else {
		m.DB.CreateLog(
			models.Log{
				Typ:     "info",
				Module:  "abonnements",
				Message: fmt.Sprintf("New Abonnement with id %d has been created", id),
			},
		)
	}
	// redirect on success
	w.Header().Set("HX-Redirect", "/admin/l3/abonnements")
	w.WriteHeader(http.StatusSeeOther)
}

// AbonnementsList is the handler to get the abonnements list
func (m *Repository) AbonnementsList(w http.ResponseWriter, r *http.Request) {

	// set initial params
	page := 1
	filter := models.NewAbonnementFilter("")
	if r.URL.Query().Has("search") {
		filter = models.NewAbonnementFilter(r.URL.Query().Get("search"))
	}

	// check if a page param has been submitted and change to this value
	if r.URL.Query().Has("page") {
		page, _ = strconv.Atoi(r.URL.Query().Get("page"))
	}

	abonnements, count, err := m.DB.AllAbonnements(page, filter)
	if err != nil {
		log.Println("could not get abonnements from DB: ", err)
		m.App.Session.Put(r.Context(), "error", "Can't get visitors from database")
	}

	data := make(map[string]interface{})
	data["abonnements"] = abonnements
	data["count"] = count
	data["pages"] = int(math.Ceil(float64(count) / 10))
	data["onPage"] = page
	data["filter"] = filter

	render.Template(w, r, "part_abonnements_list.html", &models.TemplateData{
		Data: data,
	})
}
