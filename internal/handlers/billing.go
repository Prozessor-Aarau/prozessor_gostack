package handlers

import (
	"fmt"
	"log"
	"math"
	"net/http"
	"strconv"

	"github.com/faebut/prozessor_gostack/internal/billing"
	"github.com/faebut/prozessor_gostack/internal/forms"
	"github.com/faebut/prozessor_gostack/internal/models"
	"github.com/faebut/prozessor_gostack/view/bills"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
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
	return rendering(w, r, bills.CustomBillForm(models.BillData{}, bills.FormErrors{}))
}

// HandleBillCreateNewItem returns an empty billing item
func (m *Repository) HandleBillCreateNewItem(w http.ResponseWriter, r *http.Request) error {
	var item models.BillItem

	item.ID = uuid.New().String()

	// return an empty form
	return rendering(w, r, bills.BillItem(item, []bills.ItemError{}))
}

func (m *Repository) HandleBillCreatePost(w http.ResponseWriter, r *http.Request) error {
	// parsForm.e form
	r.ParseForm()
	// show all values
	// for key, value := range r.Form {
	//
	// 	fmt.Printf("%s = %s\n", key, value)
	// }

	// create bill item
	var bill models.BillData

	// push all item fields to slices
	var item_id []string
	var item_position []int
	var item_description []string
	var item_count []float64
	var item_unit_price []float64

	for key, value := range r.Form {
		if key == "item_id" {
			for i := 0; i < len(value); i++ {
				item_id = append(item_id, value[i])
			}
		}
		if key == "item_position" {
			for i := 0; i < len(value); i++ {
				position, err := strconv.Atoi(value[i])
				if err != nil {
					return err
				}
				item_position = append(item_position, position)
			}
		}
		if key == "item_description" {
			for i := 0; i < len(value); i++ {
				item_description = append(item_description, value[i])
			}
		}
		if key == "item_count" {
			for i := 0; i < len(value); i++ {
				count, _ := strconv.ParseFloat(value[i], 64)
				item_count = append(item_count, count)
			}
		}
		if key == "item_unit_price" {
			for i := 0; i < len(value); i++ {
				price, _ := strconv.ParseFloat(value[i], 64)
				item_unit_price = append(item_unit_price, price)
			}
		}
	}
	// build items and push to slice
	var items []models.BillItem

	for i := 0; i < len(item_id); i++ {
		item := models.BillItem{
			ID:          item_id[i],
			Position:    item_position[i],
			Description: item_description[i],
			Count:       item_count[i],
			UnitPrice:   item_unit_price[i],
		}

		items = append(items, item)
	}

	// check if the necessary fields are set for the items (validation)
	var itemErrors []bills.ItemError

	for _, i := range items {
		isError := false
		var iError bills.ItemError

		if len(i.Description) < 1 {
			iError.Description = "Feld ist obligatorisch"
			isError = true
		}

		if i.Count == 0 {
			iError.Count = "Feld ist obligatorisch"
			isError = true
		}

		if i.UnitPrice == 0 {
			iError.UnitPrice = "Feld ist obligatorisch"
			isError = true
		}

		if isError {
			iError.ID = i.ID
			itemErrors = append(itemErrors, iError)
		}
	}

	bill.Items = items

	// get form field
	// first check if its a registered person or not for address
	extern := r.Form.Get("extern") == "on"

	// create a form to start validation
	form := forms.New(r.Form)
	formErrors := bills.FormErrors{}
	formErrors.Items = itemErrors

	// check if there are items and set error if not
	if len(bill.Items) == 0 {
		formErrors.Item = "Mindestens eine Position hinzufügen"
	}

	// set extern toggle state
	formErrors.Extern = extern

	// check validity of fields
	form.Required("message", "infos")
	formErrors.Message = form.Errors.Get("message")
	formErrors.Infos = form.Errors.Get("infos")

	// get fields
	bill.Message = r.Form.Get("message")
	bill.Infos = r.Form.Get("infos")

	if !extern {
		// check validity
		form.Required("member")
		formErrors.Member = form.Errors.Get("member")

		bill.Visitor.ID = r.Form.Get("member")
	} else {
		// check validity
		form.Required("street", "number", "postcode", "city")
		form.IsPostcode("postcode")
		// check if firma or name is set
		form.OneOf("firma", "firstname")
		form.OneOf("firma", "familyname")
		formErrors.Street = form.Errors.Get("street")
		formErrors.Number = form.Errors.Get("number")
		formErrors.Postcode = form.Errors.Get("postcode")
		formErrors.City = form.Errors.Get("city")
		formErrors.Firma = form.Errors.Get("firma")
		formErrors.FirstName = form.Errors.Get("firstname")
		formErrors.FamilyName = form.Errors.Get("familyname")

		// set fields
		bill.Firma = r.Form.Get("firma")
		bill.FirstName = r.Form.Get("firstname")
		bill.FamilyName = r.Form.Get("familyname")
		bill.Street = r.Form.Get("street")
		bill.Number = r.Form.Get("number")
		bill.Postcode = r.Form.Get("postcode")
		bill.City = r.Form.Get("city")
	}

	return rendering(w, r, bills.CustomBillForm(bill, formErrors))
}
