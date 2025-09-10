package handlers

import (
	"fmt"
	"log/slog"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/faebut/prozessor_gostack/internal/billing"
	"github.com/faebut/prozessor_gostack/internal/forms"
	"github.com/faebut/prozessor_gostack/internal/logger"
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
		slog.Error("could not get visitor from session", "path", r.URL.Path)
		m.App.Session.Put(r.Context(), "error", "Can't get visitor from session")

		return hxRedirect(w, r, "/admin/logout")
	}

	categories, err := m.DB.GetBillCategories()
	if err != nil {
		return fmt.Errorf("failed to get categories: %w", err)
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
		return fmt.Errorf("failed to get bills: %w", err)
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
		return fmt.Errorf("failed to get bills: %w", err)
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
		return fmt.Errorf("failed to convert string to int: %w", err)
	}

	bill, err := m.DB.GetBillByID(id)
	if err != nil {
		return fmt.Errorf("failed to get bill: %w", err)
	}

	// check if there is an invoice that has to be deleted
	if len(bill.FileName) > 0 {
		err := billing.DeleteBill(bill)
		if err != nil {
			return fmt.Errorf("failed to to delete invoice from filesystem: %w", err)
		}
	}

	// delete the delete
	err = m.DB.DeleteBill(id)
	if err != nil {
		return fmt.Errorf("failed to to delete bill: %w", err)
	}

	logger.Success(r, m.App.Session,
		fmt.Sprintf("bill with id %d successfully deleted", id))

	// return empty string
	fmt.Fprint(w, "")
	return nil
}

// HandleBillingPayment is the handler for changing the payment status
func (m *Repository) HandleBillingPayment(w http.ResponseWriter, r *http.Request) error {
	billStringID := r.Form.Get("id")

	billID, err := strconv.Atoi(billStringID)
	if err != nil {
		return fmt.Errorf("failed to convert string to int: %w", err)
	}

	bill, err := m.DB.GetBillByID(billID)
	if err != nil {
		return fmt.Errorf("failed to get bill: %w", err)
	}

	bill.Paid = !bill.Paid

	err = m.DB.SetBillPayment(bill)
	if err != nil {
		return fmt.Errorf("failed to set payment: %w", err)
	}

	// log success
	var status string
	if bill.Paid {
		status = "paid"
	} else {
		status = "unpaid"
	}

	logger.Success(r, m.App.Session,
		fmt.Sprintf("payment status of bill with id %d successfully set to %s", bill.ID, status))

	return rendering(w, r, bills.PayButton(bill))
}

// HandleBillCreateIndex is the handler for the Bill creation form
func (m *Repository) HandleBillCreateIndex(w http.ResponseWriter, r *http.Request) error {
	// get categories
	categories, err := m.DB.GetBillCategories()
	if err != nil {
		return fmt.Errorf("failed to get categories: %w", err)
	}
	// return an empty form
	return rendering(w, r, bills.CustomBillForm(models.BillData{}, bills.FormErrors{}, categories))
}

// HandleBillCreateNewItem returns an empty billing item
func (m *Repository) HandleBillCreateNewItem(w http.ResponseWriter, r *http.Request) error {
	var item models.BillItem

	item.ID = uuid.New().String()

	// return an empty form
	return rendering(w, r, bills.BillItem(item, []bills.ItemError{}))
}

func (m *Repository) HandleBillCreateEditPost(w http.ResponseWriter, r *http.Request) error {
	// set initial value
	edit := false
	// parse form
	r.ParseForm()

	// get categories
	categories, err := m.DB.GetBillCategories()
	if err != nil {
		return fmt.Errorf("failed to get categories: %w", err)
	}

	// create bill item
	var bill models.BillData

	// get route from request url
	exploded := strings.Split(r.RequestURI, "/")
	if exploded[3] == "edit" {
		id, err := strconv.Atoi(chi.URLParam(r, "id"))
		if err != nil {
			return fmt.Errorf("failed to convert string to int: %w", err)
		}
		bill.ID = id
		edit = true
	}

	// push all item fields to slices
	var item_id []string
	var item_position []int
	var item_description []string
	var item_count []float64
	var item_unit_price []float64

	for key, value := range r.Form {
		if key == "item_id" {
			for i := range value {
				item_id = append(item_id, value[i])
			}
		}
		if key == "item_position" {
			for i := range value {
				position, err := strconv.Atoi(value[i])
				if err != nil {
					return err
				}
				item_position = append(item_position, position)
			}
		}
		if key == "item_description" {
			for i := range value {
				item_description = append(item_description, value[i])
			}
		}
		if key == "item_count" {
			for i := range value {
				count, _ := strconv.ParseFloat(value[i], 64)
				item_count = append(item_count, count)
			}
		}
		if key == "item_unit_price" {
			for i := range value {
				price, _ := strconv.ParseFloat(value[i], 64)
				item_unit_price = append(item_unit_price, price)
			}
		}
	}
	// build items and push to slice
	form := forms.New(r.Form)
	var items []models.BillItem

	for i := range item_id {
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
			form.Errors.Add("item", "Feld ist obligatorisch")
		}
	}

	bill.Items = items

	// get form field
	// first check if its a registered person or not for address
	extern := r.Form.Get("extern") == "on"

	// create a form to start validation
	formErrors := bills.FormErrors{}
	formErrors.Items = itemErrors
	// set edit if it's edit
	formErrors.Edit = edit

	// check if there are items and set error if not
	if len(bill.Items) == 0 {
		form.Errors.Add("items", "Mindestens eine Position hinzufügen")
		formErrors.Item = "Mindestens eine Position hinzufügen"
	}

	// set extern toggle state
	formErrors.Extern = extern

	// check validity of fields
	form.Required("purpose", "message", "infos", "date")
	formErrors.Message = form.Errors.Get("message")
	formErrors.Infos = form.Errors.Get("infos")
	formErrors.Date = form.Errors.Get("date")
	formErrors.Purpose = form.Errors.Get("purpose")

	// get fields
	bill.Date, _ = time.Parse("2006-01-02", r.Form.Get("date"))
	bill.Message = r.Form.Get("message")
	bill.Infos = r.Form.Get("infos")
	bill.Purpose = strings.ToLower(r.Form.Get("purpose"))

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

	// what to do if form not valid
	if !form.Valid() {
		// send the data to the template
		return rendering(w, r, bills.CustomBillForm(bill, formErrors, categories))
	}

	// what to do if everything is okay
	// build BillData first
	if !extern {
		// get all of visitor
		visitor, err := m.DB.VisitorByID(bill.Visitor.ID)
		if err != nil {
			return fmt.Errorf("failed to get visitor: %w", err)
		}
		bill.Visitor = visitor
	}

	// get total
	var total float64
	for _, i := range bill.Items {
		price := (math.Round(i.Count*100*i.UnitPrice) / 100)

		total += price
	}
	bill.Amount = total

	// set bill to custom
	bill.Custom = true

	// create bill
	_, bill.ID, err = billing.CreateBill(bill, m.DB)
	if err != nil {
		return fmt.Errorf("failed to create bill: %w", err)
	}

	// log success
	logger.Success(r, m.App.Session,
		fmt.Sprintf("bill with id %d for visitor %s %s successfully created/updated", bill.ID, bill.Visitor.FirstName, bill.Visitor.FamilyName))

	return hxRedirect(w, r, "/admin/billing")
}

// HandleBillEditIndex is the handler for the Bill creation form
func (m *Repository) HandleBillEditIndex(w http.ResponseWriter, r *http.Request) error {
	bID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		return fmt.Errorf("failed to convert string to int: %w", err)
	}

	bill, err := m.DB.GetBillByID(bID)
	if err != nil {
		return fmt.Errorf("failed to get bill: %w", err)
	}

	// get categories
	categories, err := m.DB.GetBillCategories()
	if err != nil {
		return fmt.Errorf("failed to get categories: %w", err)
	}
	// return an empty form
	return rendering(
		w,
		r,
		bills.CustomBillForm(
			bill,
			bills.FormErrors{Edit: true, Extern: len(bill.Visitor.ID) < 1},
			categories,
		),
	)
}
