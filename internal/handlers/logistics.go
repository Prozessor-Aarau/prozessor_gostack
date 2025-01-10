package handlers

import (
	"fmt"
	"log"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/faebut/prozessor_gostack/internal/forms"
	"github.com/faebut/prozessor_gostack/internal/labels"
	"github.com/faebut/prozessor_gostack/internal/models"
	"github.com/faebut/prozessor_gostack/view/logistics"
	"github.com/go-chi/chi/v5"
)

// HandleLogisticsIndex is the logistics page handler
func (m *Repository) HandleLogisticsIndex(w http.ResponseWriter, r *http.Request) error {
	// get visitor
	visitor, ok := m.App.Session.Get(r.Context(), "visitor").(models.Visitor)
	if !ok {
		log.Println("could not get visitor from Session")
		return hxRedirect(w, r, "/admin/logout")
	}

	// get categories to pass on
	categories, err := m.DB.AllLogisticCategories()

	if err != nil {
		return err
	}

	return rendering(w, r, logistics.Logistics(visitor, categories))
}

// HandleLogisticsList gets and returns the unfiltered logistics list
func (m *Repository) HandleLogisticsList(w http.ResponseWriter, r *http.Request) error {
	// set initial params
	page := 1
	filter := models.NewLogisticFilter("")

	// check if page param has been submitted and change to this values
	if r.URL.Query().Has("page") {
		page, _ = strconv.Atoi(r.URL.Query().Get("page"))
	}
	if r.URL.Query().Has("search") {
		filter = models.NewLogisticFilter(r.URL.Query().Get("search"))
	}
	if r.URL.Query().Has("validity") {
		filter.Validity = r.URL.Query().Get("validity")
	}
	if r.URL.Query().Has("inuse") {
		filter.InUse = r.URL.Query().Get("inuse")
	}
	if r.URL.Query().Has("type") {
		filter.Type, _ = strconv.Atoi(r.URL.Query().Get("type"))
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

// HandleLogisticsListFilter gets and returns the filtered logistics list
func (m *Repository) HandleLogisticsListFilter(w http.ResponseWriter, r *http.Request) error {
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

// HandleBadgeEditIndex is the handler for the Badge edit form
func (m *Repository) HandleLogisticsEditIndex(w http.ResponseWriter, r *http.Request) error {
	// get categories to pass on
	categories, err := m.DB.AllLogisticCategories()

	if err != nil {
		return err
	}

	lID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		log.Println("Error converting badge id to int", err)
		return err
	}

	l, err := m.DB.LogisticsByID(lID)
	if err != nil {
		log.Println("Error getting logistic from database:", err)
		return err
	}

	return rendering(w, r, logistics.LogisticsForm(
		l,
		categories,
		logistics.FormErrors{},
		len(l.Visitor.ID) > 0,
	))
}

// HandleLogisticsCreateIndex is the handler for the logistic creation form
func (m *Repository) HandleLogisticsCreateIndex(w http.ResponseWriter, r *http.Request) error {
	categories, err := m.DB.AllLogisticCategories()

	if err != nil {
		return err
	}
	// return an empty form
	return rendering(w, r, logistics.LogisticsForm(models.Logistic{}, categories, logistics.FormErrors{}, false))
}

// HandleLogisticsCreateAndEdit is the handler for posting the logistic creation and edit form
func (m *Repository) HandleLogisticsCreateAndEdit(w http.ResponseWriter, r *http.Request) error {
	categories, err := m.DB.AllLogisticCategories()

	if err != nil {
		return err
	}
	// create empty logistic
	var newLogistic models.Logistic
	var formErrors logistics.FormErrors

	form := forms.New(r.Form)
	// get checkboxes
	assigned := form.BoxValue("assign")
	freePlace := form.BoxValue("free")

	// check needed values for both usecases
	if freePlace {
		catID, err := strconv.Atoi(r.Form.Get("free_cat"))
		if err != nil {
			return err
		}
		newLogistic.Category.ID = catID
		// get category and set number
		if newLogistic.Category.ID == 4 {
			newLogistic.Number = 200
		} else {
			newLogistic.Number = 300
		}

		// check assignment
		form.Required("startdate", "enddate", "member", "free_price")
		newLogistic.Visitor.ID = r.Form.Get("member")
		unitPrice, err := strconv.ParseFloat(r.Form.Get("free_price"), 64)
		if err != nil {
			form.Errors.Add("free_price", "Nur Zahlen im Format 0.00 sind erlaubt")
		}
		newLogistic.UnitPrice = unitPrice
		newLogistic.Paid = form.BoxValue("paid")
		// parse dates
		start, err := time.Parse("2006-01-02", r.Form.Get("startdate"))
		if err != nil {
			log.Println("cannot parse date", err)
		}
		newLogistic.StartDate = start
		end, err := time.Parse("2006-01-02", r.Form.Get("enddate"))
		if err != nil {
			log.Println("cannot parse date", err)
		}
		newLogistic.EndDate = end

		if !form.Valid() {
			// get errors and then render form again
			formErrors.StartDate = form.Errors.Get("startdate")
			formErrors.EndDate = form.Errors.Get("enddate")
			if len(form.Errors.Get("member")) > 0 {
				formErrors.Visitor = "Bitte eine/n Besucher/in auswählen"
			}
			formErrors.UnitPrice = form.Errors.Get("free_price")
			assigned = true
			return rendering(w, r, logistics.LogisticsForm(newLogistic, categories, formErrors, assigned))
		}

		// get membership
		member, err := strconv.ParseBool(r.Form.Get("membership"))
		if err != nil {
			log.Println("cannot parse bool", err)
		}
		newLogistic.Member = member

		newLogistic.Price = calcPrice(newLogistic.StartDate, newLogistic.EndDate, newLogistic.UnitPrice)

	} else {
		// check needed fields
		form.Required("number")
		number, err := strconv.Atoi(r.Form.Get("number"))
		if err != nil {
			log.Println("can not parse number", err)
		}
		newLogistic.Number = number

		if number < 1 || number > 199 {
			form.Errors.Add("number", "Nummer muss zwischen 1 und 199 liegen")
		}
		catID, err := strconv.Atoi(r.Form.Get("fix_cat"))
		if err != nil {
			return err
		}
		newLogistic.Category.ID = catID

		if assigned {
			// check assignment
			form.Required("startdate", "enddate", "member")
			newLogistic.Visitor.ID = r.Form.Get("member")

			// parse dates
			start, _ := time.Parse("2006-01-02", r.Form.Get("startdate"))
			newLogistic.StartDate = start
			end, _ := time.Parse("2006-01-02", r.Form.Get("enddate"))
			newLogistic.EndDate = end
			newLogistic.Paid = form.BoxValue("paid")

			if !form.Valid() {
				// get errors and then render form again
				formErrors.StartDate = form.Errors.Get("startdate")
				formErrors.EndDate = form.Errors.Get("enddate")
				if len(form.Errors.Get("member")) > 0 {
					formErrors.Visitor = "Bitte eine/n Besucher/in auswählen"
				}
				formErrors.Number = form.Errors.Get("number")
				return rendering(w, r, logistics.LogisticsForm(newLogistic, categories, formErrors, assigned))
			}

			// get membership
			member, err := strconv.ParseBool(r.Form.Get("membership"))
			if err != nil {
				log.Println("cannot parse bool", err)
			}
			newLogistic.Member = member

			// get right price for members and non-members
			for _, c := range categories {
				if c.ID == newLogistic.Category.ID {
					if member {
						newLogistic.UnitPrice = float64(c.PriceMember)
					} else {
						newLogistic.UnitPrice = float64(c.Price)
					}
				}
			}

			newLogistic.Price = calcPrice(newLogistic.StartDate, newLogistic.EndDate, newLogistic.UnitPrice)
		} else {
			// creation of empty new fixed logistics place

			if !form.Valid() {
				formErrors.Number = form.Errors.Get("number")
				return rendering(w, r, logistics.LogisticsForm(newLogistic, categories, formErrors, assigned))
			}
		}
	}

	id := r.Form.Get("id")
	newLogistic.ID, _ = strconv.Atoi(id)
	if newLogistic.ID > 0 {

		// check if there is a label and if it is true, delete the old label
		label := r.Form.Get("label")
		if len(label) > 0 {
			newLogistic.Label = label
			err := labels.DeleteLabel(newLogistic)
			if err != nil {
				return err
			}
			err = m.DB.DeleteLogisticLabel(newLogistic)
			if err != nil {
				return err
			}
			newLogistic.Label = ""
		}

		// update existing Logistic
		err = m.DB.UpdateLogistic(newLogistic)
		if err != nil {
			return err
		}
	} else {
		// create a new logistic place
		err = m.DB.CreateLogistic(newLogistic)
		if err != nil {
			return err
		}
	}

	// redirect on success
	return hxRedirect(w, r, "/admin/logistics")
}

// calcPrice calculates the price for a given duration of logistic rent and returns a price
func calcPrice(start time.Time, end time.Time, unit float64) float64 {
	diff := end.Sub(start)
	diffMonths := diff.Hours() / 24 / 30
	var timeFactor float64

	if diffMonths <= 0.5 {
		timeFactor = 0.5
	} else {
		timeFactor = math.Ceil(diffMonths)
	}

	return unit * timeFactor
}

// HandleLogisticsDelete is the handler for deleting a logistic place
func (m *Repository) HandleLogisticsDelete(w http.ResponseWriter, r *http.Request) error {
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

// HandleLogisticsLabelCreate handles the Logistic label generation
func (m *Repository) HandleLogisticsLabelCreate(w http.ResponseWriter, r *http.Request) error {
	// get id
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))

	l, err := m.DB.LogisticsByID(id)
	if err != nil {
		log.Println("could not get logistic from DB: ", err)
		return err
	}

	// generate file
	filename, err := labels.GenerateLabel(l)
	if err != nil {
		log.Println("error creating label", err)
		return err
	}

	// write changes to database
	l.Label = filename
	err = m.DB.SetLogisticLabel(l)
	if err != nil {
		log.Println("could not write Label to DB: ", err)
		return nil
	}

	// if everything okay, go back to machines
	return rendering(w, r, logistics.Label(l))
}

// HandleLogisticsLabelDelete handles the Logistic label deletion
func (m *Repository) HandleLogisticsLabelDelete(w http.ResponseWriter, r *http.Request) error {
	// get id and logistic
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))
	l, err := m.DB.LogisticsByID(id)
	if err != nil {
		log.Println("could not get logistic from DB: ", err)
		return err
	}

	err = labels.DeleteLabel(l)

	if err != nil {
		log.Println("could not Delete label: ", err)
		return err
	}

	// write changes to database
	err = m.DB.DeleteLogisticLabel(l)
	if err != nil {
		log.Println("could not delete Label from DB: ", err)
		return err
	}

	l.Label = ""

	// if everything okay, go back to machines
	return rendering(w, r, logistics.Label(l))
}

// HandleLogisticsLabelPrint handles the logistic label printing
func (m *Repository) HandleLogisticsLabelPrint(w http.ResponseWriter, r *http.Request) error {
	// get id and logistic
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))
	l, err := m.DB.LogisticsByID(id)
	if err != nil {
		log.Println("could not get logistic from DB: ", err)
		return err
	}

	// print the QR-Code
	err = labels.PrintLabel(l)
	if err != nil {
		log.Println(err)
		return rendering(w, r, logistics.ErrorButton("Druckerfehler!"))
	}

	return rendering(w, r, logistics.SucessButton(l))
}
