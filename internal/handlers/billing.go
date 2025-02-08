package handlers

import (
	"fmt"
	"log"
	"math"
	"net/http"
	"strconv"

	"github.com/faebut/prozessor_gostack/internal/billing"
	"github.com/faebut/prozessor_gostack/internal/models"
	"github.com/faebut/prozessor_gostack/view/bills"
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

	categories, err := m.DB.GetBillCategories()
	if err != nil {
		return err
	}

	return rendering(w, r, bills.Bills(visitor, categories))
}

// HandleLogisticsList gets and returns the unfiltered logistics list
func (m *Repository) HandleBillingList(w http.ResponseWriter, r *http.Request) error {
	// set initial params
	page := 1
	filter := models.NewBillFilter("")

	// check if page param has been submitted and change to this values
	if r.URL.Query().Has("page") {
		page, _ = strconv.Atoi(r.URL.Query().Get("page"))
	}
	if r.URL.Query().Has("search") {
		filter = models.NewBillFilter(r.URL.Query().Get("search"))
	}
	if r.URL.Query().Has("payment") {
		filter.Payment = r.URL.Query().Get("payment")
	}
	if r.URL.Query().Has("type") {
		filter.Type = r.URL.Query().Get("type")
	}

	billsList, count, err := m.DB.GetAllBills(page, filter)
	if err != nil {
		log.Println("could not get bills from DB: ", err)
		return err
	}

	// make url search string
	urlSearchString := fmt.Sprintf(
		"&search=%s&payment=%s&type=%s",
		filter.Filter,
		filter.Payment,
		filter.Type,
	)

	return rendering(w, r, bills.BillsList(
		billsList,
		int(math.Ceil(float64(count)/10)),
		page,
		urlSearchString,
	))
}

// HandleLogisticsListFilter gets and returns the filtered logistics list
func (m *Repository) HandleBillingListFilter(w http.ResponseWriter, r *http.Request) error {
	// set initial params
	// need empty filter to get logistics
	filterString := r.Form.Get("filter_string")
	filter := models.NewBillFilter(filterString)
	filter.Payment = r.Form.Get("payment")
	filter.Type = r.Form.Get("type")

	page := 1
	if r.URL.Query().Has("page") {
		page, _ = strconv.Atoi(r.URL.Query().Get("page"))
	}

	billsList, count, err := m.DB.GetAllBills(page, filter)
	if err != nil {
		log.Println("could not get bills from DB: ", err)
		return err
	}

	// make url search string
	urlSearchString := fmt.Sprintf(
		"&search=%s&payment=%s&type=%s",
		filter.Filter,
		filter.Payment,
		filter.Type,
	)

	return rendering(w, r, bills.BillsList(
		billsList,
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

	bill, err := m.DB.GetBillByID(id)
	if err != nil {
		log.Println("could not get bill from DB: ", err)
		return err
	}

	// check if there is an invoice that has to be deleted
	if len(bill.FileName) > 0 {
		err := billing.DeleteBill(bill)
		if err != nil {
			return err
		}
	}

	// delete the delete
	err = m.DB.DeleteBill(id)
	if err != nil {
		log.Println("Error deleting bill:", err)
		return err
	}

	// return empty string
	fmt.Fprint(w, "")
	return nil
}

// HandleBillingPayment is the handler for changing the payment status
func (m *Repository) HandleBillingPayment(w http.ResponseWriter, r *http.Request) error {
	billStringID := r.Form.Get("id")

	billID, err := strconv.Atoi(billStringID)
	if err != nil {
		log.Println("could not convert id to int:", err)
		return err
	}

	bill, err := m.DB.GetBillByID(billID)
	if err != nil {
		log.Println("could not get bill from db:", err)
		return err
	}

	bill.Paid = !bill.Paid

	err = m.DB.SetBillPayment(bill)
	if err != nil {
		log.Println("could not set payment:", err)
		return err
	}

	return rendering(w, r, bills.PayButton(bill))
}

// HandleBillCreateIndex is the handler for the Bill creation form
func (m *Repository) HandleBillCreateIndex(w http.ResponseWriter, r *http.Request) error {
	// return an empty form
	return rendering(w, r, bills.CustomBillForm())
}

// HandleBillCreateNewItem returns an empty billing item
func (m *Repository) HandleBillCreateNewItem(w http.ResponseWriter, r *http.Request) error {
	// return an empty form
	return rendering(w, r, bills.BillItem(models.BillItem{}))
}
