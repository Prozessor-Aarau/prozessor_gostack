package handlers

import (
	"fmt"
	"log"
	"math"
	"net/http"
	"strconv"

	"github.com/faebut/prozessor_gostack/internal/labels"
	"github.com/faebut/prozessor_gostack/internal/models"
	"github.com/faebut/prozessor_gostack/view/bills"
	"github.com/faebut/prozessor_gostack/view/logistics"
	"github.com/go-chi/chi/v5"
)

// HandleLogisticsIndex is the logistics page handler
func (m *Repository) HandleBillingIndex(w http.ResponseWriter, r *http.Request) error {
	// get visitor
	visitor, ok := m.App.Session.Get(r.Context(), "visitor").(models.Visitor)
	if !ok {
		log.Println("could not get visitor from Session")
		return hxRedirect(w, r, "/admin/logout")
	}

	return rendering(w, r, bills.Bills(visitor))
}

// HandleLogisticsList gets and returns the unfiltered logistics list
func (m *Repository) HandleBillingList(w http.ResponseWriter, r *http.Request) error {
	// set initial params
	page := 1
	// filter := models.NewLogisticFilter("")

	// check if page param has been submitted and change to this values
	// if r.URL.Query().Has("page") {
	// 	page, _ = strconv.Atoi(r.URL.Query().Get("page"))
	// }
	// if r.URL.Query().Has("search") {
	// 	filter = models.NewLogisticFilter(r.URL.Query().Get("search"))
	// }
	// if r.URL.Query().Has("validity") {
	// 	filter.Validity = r.URL.Query().Get("validity")
	// }
	// if r.URL.Query().Has("inuse") {
	// 	filter.InUse = r.URL.Query().Get("inuse")
	// }
	// if r.URL.Query().Has("type") {
	// 	filter.Type, _ = strconv.Atoi(r.URL.Query().Get("type"))
	// }

	billsList, count, err := m.DB.GetAllBills()
	if err != nil {
		log.Println("could not get bills from DB: ", err)
		return err
	}

	// make url search string
	// urlSearchString := fmt.Sprintf(
	// 	"&search=%s&validity=%s&inuse=%s&type=%d",
	// 	filter.Filter,
	// 	filter.Validity,
	// 	filter.InUse,
	// 	filter.Type,
	// )

	return rendering(w, r, bills.BillsList(
		billsList,
		int(math.Ceil(float64(count)/10)),
		page,
		"",
	))
}

// HandleLogisticsListFilter gets and returns the filtered logistics list
func (m *Repository) HandleBillingListFilter(w http.ResponseWriter, r *http.Request) error {
	// set initial params
	// need empty filter to get logistics
	filterString := r.Form.Get("filter_string")
	filter := models.NewLogisticFilter(filterString)
	filter.InUse = r.Form.Get("inuse")
	filter.Validity = r.Form.Get("validity")
	filter.Type, _ = strconv.Atoi(r.Form.Get("type"))

	page := 1
	if r.URL.Query().Has("page") {
		page, _ = strconv.Atoi(r.URL.Query().Get("page"))
	}

	logisticsList, count, err := m.DB.AllLogistics(page, filter)
	if err != nil {
		log.Println("could not get logistics from DB: ", err)
		return err
	}

	// make url search string
	urlSearchString := fmt.Sprintf(
		"&search=%s&validity=%s&inuse=%s&type=%d",
		filter.Filter,
		filter.Validity,
		filter.InUse,
		filter.Type,
	)

	return rendering(w, r, logistics.LogisticsList(
		logisticsList,
		int(math.Ceil(float64(count)/10)),
		page,
		urlSearchString,
	))
}

// HandleLogisticsDelete is the handler for deleting a logistic place
func (m *Repository) HandleBillingDelete(w http.ResponseWriter, r *http.Request) error {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		log.Println("Error converting string to int:", err)
		return err
	}

	l, err := m.DB.LogisticsByID(id)
	if err != nil {
		log.Println("could not get logistic from DB: ", err)
		return err
	}

	// check if there is a label that has to be deleted
	if len(l.Label) > 0 {
		err := labels.DeleteLabel(l)
		if err != nil {
			return err
		}
	}

	// delete the logistic place
	err = m.DB.DeleteLogistic(id)
	if err != nil {
		log.Println("Error deleting logistics:", err)
		return err
	}
	// redirect on success
	return hxRedirect(w, r, "/admin/logistics")
}
