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
	"github.com/faebut/prozessor_gostack/view/abonnements"
	"github.com/go-chi/chi/v5"
)

// Abonnements is the abonnements page handler
func (m *Repository) HandleAbonnementsIndex(w http.ResponseWriter, r *http.Request) error {
	visitor, ok := m.App.Session.Get(r.Context(), "visitor").(models.Visitor)
	if !ok {
		log.Println("could not get visitor from Session")
		return hxRedirect(w, r, "/admin/logout")
	}

	return rendering(w, r, abonnements.Abonnements(visitor))
}

// HandleAbonnementCreateIndex is the abonnements creation page handler
func (m *Repository) HandleAbonnementCreateIndex(w http.ResponseWriter, r *http.Request) error {
	var abo models.Abonnement

	return rendering(w, r, abonnements.AbonnementForm(abo, abonnements.FormErrors{}, false))

}

// AbonnementCreate is the abonnements creation page handler
func (m *Repository) AbonnementCreatePrice(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	visitor, err := m.DB.VisitorByID(id)
	if err != nil {
		log.Println("could not get visitor from DB: ", err)
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
func (m *Repository) HandleAbonnementCreate(w http.ResponseWriter, r *http.Request) error {
	// create empty badge
	var abo models.Abonnement
	var free bool

	form := forms.New(r.Form)
	form.Required("price", "member", "buydate")

	abo.BuyDate = r.Form.Get("buydate")
	stringPrice, err := strconv.Atoi(r.Form.Get("price"))
	if err != nil {
		log.Println("error converting string to int:", err)
	}
	abo.Price = stringPrice
	abo.VisitorID = r.Form.Get("member")
	abo.PartialPayment = form.BoxValue("partial_payment")
	free = form.BoxValue("free")

	// set member with price
	if abo.Price == 320 {
		abo.Member = true
	} else {
		abo.Member = false
	}

	var formErrors abonnements.FormErrors
	if free && !abo.Member {
		form.Errors.Add("free", "Nur Mitglieder können beitragsbefreit sein")
		formErrors.Free = form.Errors.Get("free")
	}
	formErrors.Member = form.Errors.Get("member")
	formErrors.Price = form.Errors.Get("price")
	formErrors.BuyDate = form.Errors.Get("buydate")

	if !form.Valid() {
		return rendering(w, r, abonnements.AbonnementForm(abo, formErrors, free))
	}

	// if free membership, delete price
	if free {
		abo.Price = 0
	}

	_, err = m.DB.CreateAbonnement(abo)
	if err != nil {
		log.Println("Error creating abonnement:", err)
		return err
	}

	// redirect on success
	return hxRedirect(w, r, "/admin/abonnements")
}

// AbonnementsList is the handler to get the abonnements list
func (m *Repository) HandleAbonnementsList(w http.ResponseWriter, r *http.Request) error {
	// set initial params
	page := 1
	filter := models.NewAbonnementFilter("")

	// check if page param has been submitted and change to this values
	if r.URL.Query().Has("page") {
		page, _ = strconv.Atoi(r.URL.Query().Get("page"))
	}
	if r.URL.Query().Has("search") {
		filter = models.NewAbonnementFilter(r.URL.Query().Get("search"))
	}
	if r.URL.Query().Has("sort") {
		filter.Sort = r.URL.Query().Get("sort")
	}
	if r.URL.Query().Has("sortdir") {
		filter.SortDir = strings.ToUpper(r.URL.Query().Get("sortdir"))
	}
	if r.URL.Query().Has("validity") {
		filter.AboValidity = r.URL.Query().Get("validity")
	}
	if r.URL.Query().Has("member") {
		filter.Member, _ = strconv.ParseBool(r.URL.Query().Get("member"))
	}
	// TODO here is something off with this filter
	if r.URL.Query().Has("unique") {
		filter.Unique, _ = strconv.ParseBool(r.URL.Query().Get("partner"))
	}

	aboList, count, err := m.DB.AllAbonnements(page, filter)
	if err != nil {
		log.Println("could not get abonnements from DB: ", err)
		return err
	}

	// make url search string
	urlSearchString := fmt.Sprintf(
		"&search=%s&sort=%s&sortdir=%s&validity=%s",
		filter.SearchString,
		filter.Sort,
		filter.SortDir,
		filter.AboValidity,
	)
	if filter.Member {
		urlSearchString = urlSearchString + "&member=true"
	}
	if filter.Unique {
		urlSearchString = urlSearchString + "&unique=true"
	}

	return rendering(w, r, abonnements.AbonnementsList(
		aboList,
		count,
		int(math.Ceil(float64(count)/10)),
		page,
		urlSearchString,
	))
}

// PostAbonnementsList is the handler to get the abonnements list with filter
func (m *Repository) HandleAbonnementsListFilter(w http.ResponseWriter, r *http.Request) error {

	// set initial params
	page := 1
	// check if a page param has been submitted and change to this value
	if r.URL.Query().Has("page") {
		page, _ = strconv.Atoi(r.URL.Query().Get("page"))
	}
	searchTerm := r.Form.Get("search_term")

	form := forms.New(r.PostForm)
	filter := models.NewAbonnementFilter(searchTerm)
	filter.Sort = r.Form.Get("sort")
	filter.SortDir = r.Form.Get("sort_dir")
	filter.AboValidity = r.Form.Get("abovalidity")
	filter.Member = form.BoxValue("member")
	filter.Unique = form.BoxValue("unique")
	if r.URL.Query().Has("search") {
		filter = models.NewAbonnementFilter(r.URL.Query().Get("search"))
	}

	aboList, count, err := m.DB.AllAbonnements(page, filter)
	if err != nil {
		log.Println("could not get abonnements from DB: ", err)
		return err
	}

	// make url search string
	urlSearchString := fmt.Sprintf(
		"&search=%s&sort=%s&sortdir=%s&validity=%s",
		filter.SearchString,
		filter.Sort,
		filter.SortDir,
		filter.AboValidity,
	)
	if filter.Member {
		urlSearchString = urlSearchString + "&member=true"
	}
	if filter.Unique {
		urlSearchString = urlSearchString + "&unique=true"
	}

	return rendering(w, r, abonnements.AbonnementsList(
		aboList,
		count,
		int(math.Ceil(float64(count)/10)),
		page,
		urlSearchString,
	))
}
