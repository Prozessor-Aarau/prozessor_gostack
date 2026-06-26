package handlers

import (
	"fmt"
	"log"
	"log/slog"
	"math"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/faebut/prozessor_gostack/internal/billing"
	"github.com/faebut/prozessor_gostack/internal/forms"
	"github.com/faebut/prozessor_gostack/internal/helpers"
	"github.com/faebut/prozessor_gostack/internal/labels"
	"github.com/faebut/prozessor_gostack/internal/logger"
	"github.com/faebut/prozessor_gostack/internal/mailtemplates"
	"github.com/faebut/prozessor_gostack/internal/models"
	"github.com/faebut/prozessor_gostack/view/layout"
	"github.com/faebut/prozessor_gostack/view/logistics"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
)

// HandleLogisticsIndex is the logistics page handler
func (m *Repository) HandleLogisticsIndex(w http.ResponseWriter, r *http.Request) error {
	// get visitor
	visitor, ok := m.App.Session.Get(r.Context(), "visitor").(models.Visitor)
	if !ok {
		slog.Error("could not get visitor from session", "path", r.URL.Path)
		m.App.Session.Put(r.Context(), "error", "Can't get visitor from session")

		return hxRedirect(w, r, "/admin/logout")
	}

	// get categories to pass on
	categories, err := m.DB.AllLogisticCategories()
	if err != nil {
		return fmt.Errorf("failed to get categories: %w", err)
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
		return fmt.Errorf("failed to get logistics: %w", err)
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
		return fmt.Errorf("failed to get logistics: %w", err)
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
		return fmt.Errorf("failed to convert string to int: %w", err)
	}

	l, err := m.DB.LogisticsByID(lID)
	if err != nil {
		return fmt.Errorf("failed to get logistics: %w", err)
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
		return fmt.Errorf("failed to get categories: %w", err)
	}
	// return an empty form
	return rendering(w, r, logistics.LogisticsForm(models.Logistic{}, categories, logistics.FormErrors{}, false))
}

// TODO check if there is allready a bill made
// TODO check what to do?
// do I want to delete the existing bill, do I want to create a new one?
// should check, if the duration is changed or the type

// HandleLogisticsCreateAndEdit is the handler for posting the logistic creation and edit form
func (m *Repository) HandleLogisticsCreateAndEdit(w http.ResponseWriter, r *http.Request) error {
	categories, err := m.DB.AllLogisticCategories()
	// TODO check if enddate is after startdate
	if err != nil {
		return fmt.Errorf("failed to get categories: %w", err)
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
			return fmt.Errorf("failed to convert string to int: %w", err)
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
		// parse dates
		start, err := time.Parse("2006-01-02", r.Form.Get("startdate"))
		if err != nil {
			return fmt.Errorf("failed to parse date: %w", err)
		}
		newLogistic.StartDate = start
		end, err := time.Parse("2006-01-02", r.Form.Get("enddate"))
		if err != nil {
			return fmt.Errorf("failed to parse date: %w", err)
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
			return fmt.Errorf("failed to parse bool: %w", err)
		}
		newLogistic.Member = member

		newLogistic.Price = calcPrice(newLogistic.StartDate, newLogistic.EndDate, newLogistic.UnitPrice)

	} else {
		// check needed fields
		form.Required("number")
		number, err := strconv.Atoi(r.Form.Get("number"))
		if err != nil {
			return fmt.Errorf("failed to parse number: %w", err)
		}
		newLogistic.Number = number

		if number < 1 || number > 199 {
			form.Errors.Add("number", "Nummer muss zwischen 1 und 199 liegen")
		}
		catID, err := strconv.Atoi(r.Form.Get("fix_cat"))
		if err != nil {
			return fmt.Errorf("failed to convert string to int: %w", err)
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
				return fmt.Errorf("failed to parse membership: %w", err)
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

	// define action
	var action string

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
			return fmt.Errorf("failed to update logistic: %w", err)
		}
		action = "updated"
	} else {
		// create a new logistic place
		err = m.DB.CreateLogistic(newLogistic)
		if err != nil {
			return fmt.Errorf("failed to create logistic: %w", err)
		}
		action = "created"
	}

	// log success
	visitor, err := m.DB.VisitorByID(newLogistic.Visitor.ID)
	if err == nil {
		logger.Success(r, m.App.Session,
			fmt.Sprintf("logistic for visitor %s %s successfully %s", visitor.FirstName, visitor.FamilyName, action))
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
		timeFactor = math.Round(diffMonths)
	}

	return unit * timeFactor
}

// HandleLogisticsDelete is the handler for deleting a logistic place
func (m *Repository) HandleLogisticsDelete(w http.ResponseWriter, r *http.Request) error {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		return fmt.Errorf("failed to convert string to int: %w", err)
	}

	l, err := m.DB.LogisticsByID(id)
	if err != nil {
		return fmt.Errorf("failed to get logistic: %w", err)
	}

	// check if there is a label that has to be deleted
	if len(l.Label) > 0 {
		err := labels.DeleteLabel(l)
		if err != nil {
			return fmt.Errorf("failed to delete label: %w", err)
		}
	}

	// delete the logistic place
	err = m.DB.DeleteLogistic(id)
	if err != nil {
		return fmt.Errorf("failed to delete logistic: %w", err)
	}

	// log success
	logger.Success(r, m.App.Session,
		fmt.Sprintf("logistics with id %d belonging to %s %s successfully deleted", id, l.Visitor.FirstName, l.Visitor.FamilyName))

	// redirect on success
	return hxRedirect(w, r, "/admin/logistics")
}

// HandleLogisticsLabelCreate handles the Logistic label generation
func (m *Repository) HandleLogisticsLabelCreate(w http.ResponseWriter, r *http.Request) error {
	// get id
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))

	l, err := m.DB.LogisticsByID(id)
	if err != nil {
		return fmt.Errorf("failed to get logistic: %w", err)
	}

	// generate file
	filename, err := labels.GenerateLabel(l)
	if err != nil {
		return fmt.Errorf("failed to generate label: %w", err)
	}

	// write changes to database
	l.Label = filename
	err = m.DB.SetLogisticLabel(l)
	if err != nil {
		return fmt.Errorf("failed to write label to DB: %w", err)
	}

	// log success
	logger.Success(r, m.App.Session,
		fmt.Sprintf("label for logistic with id %d successfully created", id))

	// if everything okay, go back to machines
	return rendering(w, r, logistics.Label(l))
}

// HandleLogisticsLabelDelete handles the Logistic label deletion
func (m *Repository) HandleLogisticsLabelDelete(w http.ResponseWriter, r *http.Request) error {
	// get id and logistic
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))
	l, err := m.DB.LogisticsByID(id)
	if err != nil {
		return fmt.Errorf("failed to get logistic: %w", err)
	}

	err = labels.DeleteLabel(l)
	if err != nil {
		return fmt.Errorf("failed to delete label: %w", err)
	}

	// write changes to database
	err = m.DB.DeleteLogisticLabel(l)
	if err != nil {
		return fmt.Errorf("failed to delete label from DB: %w", err)
	}

	l.Label = ""

	// log success
	logger.Success(r, m.App.Session,
		fmt.Sprintf("label for logistic with id %d successfully deleted", id))

	// if everything okay, go back to machines
	return rendering(w, r, logistics.Label(l))
}

// HandleLogisticsLabelPrint handles the logistic label printing
func (m *Repository) HandleLogisticsLabelPrint(w http.ResponseWriter, r *http.Request) error {
	// get id and logistic
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))
	l, err := m.DB.LogisticsByID(id)
	if err != nil {
		return fmt.Errorf("failed to get logistic: %w", err)
	}

	// print the QR-Code
	err = labels.PrintLabel(l)
	if err != nil {
		slog.Error(fmt.Errorf("could not print label: %w", err).Error())
		return rendering(w, r, logistics.ErrorButton("Druckerfehler!"))
	}

	return rendering(w, r, logistics.SucessButton(l))
}

// HandleLogisticsCreateBillIndex is the handler for the logistic bill creation form
func (m *Repository) HandleLogisticsCreateBillIndex(w http.ResponseWriter, r *http.Request) error {
	return rendering(w, r, logistics.LogisticsBill())
}

// HandleLogisticsCreateBillIndex is the handler for the logistic bill creation form
func (m *Repository) HandleLogisticsBillList(w http.ResponseWriter, r *http.Request) error {
	// get visitor ID
	vID := chi.URLParam(r, "id")

	// get logistics by visitor id
	logList, err := m.DB.LogisticsByVisitorID(vID)
	if err != nil {
		return fmt.Errorf("failed to get logistics: %w", err)
	}

	// return an empty form
	return rendering(w, r, logistics.BillList(logList, vID))
}

// HandleLogisticsCreateBillIndex is the handler for the logistic bill creation form
func (m *Repository) HandleLogisticsCreateBillPost(w http.ResponseWriter, r *http.Request) error {
	sending := r.URL.Query().Has("send")

	vID := r.Form.Get("vid")

	if len(vID) < 1 {
		return fmt.Errorf("no visitor ID given")
	}

	// get visitor
	visitor, err := m.DB.VisitorByID(vID)
	if err != nil {
		return fmt.Errorf("failed to get visitor: %w", err)
	}

	// get logistics by visitor id
	logList, err := m.DB.LogisticsByVisitorID(vID)
	if err != nil {
		return fmt.Errorf("failed to get logistic: %w", err)
	}

	// check for all ids if the checkbox is checked and add visitor to new slice
	form := forms.New(r.Form)
	var selectedLogistics []models.Logistic
	var total float64
	var numString string
	var items []models.BillItem
	for _, l := range logList {
		if form.BoxValue(strconv.Itoa(l.ID)) {
			selectedLogistics = append(selectedLogistics, l)
			// calculate total
			total += l.Price
			numString += "[" + strconv.Itoa(l.Number) + "]"

			var unitprice float64
			if l.Member {
				unitprice = l.Category.PriceMember
			} else {
				unitprice = l.Category.Price
			}

			item := models.BillItem{
				ID:          uuid.NewString(),
				Position:    l.Number,
				Description: l.Category.CatName + ": " + l.StartDate.Format("02.01.2006") + " - " + l.EndDate.Format("02.01.2006"),
				UnitPrice:   unitprice,
				Count:       calcPrice(l.StartDate, l.EndDate, 1),
			}

			items = append(items, item)
		}
	}

	// create bill
	file, billID, err := billing.CreateBill(
		models.BillData{
			Date:    time.Now(),
			Purpose: "lagerplatz",
			Visitor: visitor,
			Amount:  total,
			Infos: fmt.Sprintf(
				"Lagerplätze für %s %s, Plätze: %s",
				visitor.FirstName,
				visitor.FamilyName,
				numString,
			),
			Message: fmt.Sprintf(
				"Lagerplätze %s",
				numString,
			),
			Items: items,
		},
		m.DB,
	)
	if err != nil {
		return err
	}

	// create slice of int of ids from selected logs
	var logSlice []int
	for _, l := range selectedLogistics {
		logSlice = append(logSlice, l.ID)
	}

	// write bill id to the logistics
	err = m.DB.UpdateLogisticsBillID(billID, logSlice)
	if err != nil {
		return fmt.Errorf("failed to update logistics bill id: %w", err)
	}

	// if mail has to be send create mail and send it
	if sending {
		// get email template
		template, err := mailtemplates.LogisticsBill(visitor, selectedLogistics)
		if err != nil {
			return fmt.Errorf("failed to get email template: %w", err)
		}

		// create mail
		msg := models.MailData{
			To:      visitor.Email,
			Subject: "Rechnung Lagerplätze",
			Content: template,
			File:    file,
		}

		// send mail
		m.App.MailChan <- msg
	}

	// message string for success
	var messageString string
	var action string
	if sending {
		messageString = "Rechnung erfolgreich versendet."
		action = "sent"
	} else {
		messageString = "Rechnung erfolgreich gespeichert."
		action = "saved"
	}

	// log success
	logger.Success(r, m.App.Session,
		fmt.Sprintf("bill for %s %s successfully %s", visitor.FirstName, visitor.FamilyName, action))

	return rendering(w, r, layout.Success(messageString))
}

// HandleLogisticsRequestIndex is the handler for the logistics request index page
func (m *Repository) HandleLogisticsRequestIndex(w http.ResponseWriter, r *http.Request) error {
	// get categories to pass on
	categories, err := m.DB.AllLogisticCategories()
	if err != nil {
		return fmt.Errorf("failed to get categories: %w", err)
	}

	// return the request index page
	return rendering(w, r, logistics.Request(categories))
}

// HandleLogisticsRequestForm handles the logistics request form
func (m *Repository) HandleLogisticsRequestForm(w http.ResponseWriter, r *http.Request) error {
	err := r.ParseForm()
	if err != nil {
		return fmt.Errorf("failed to parse form: %w", err)
	}

	// build request params
	var params logistics.RequestParams

	params.Email = strings.TrimSpace(r.Form.Get("email"))
	params.Reason = strings.TrimSpace(r.Form.Get("reason"))
	params.Category.ID, err = strconv.Atoi(r.Form.Get("category"))
	if err != nil {
		return fmt.Errorf("parse category: %w", err)
	}

	// parse dates
	startDate := r.Form.Get("startdate")
	if len(startDate) > 0 {
		params.StartDate, err = time.Parse("2006-01-02", startDate)
		if err != nil {
			return fmt.Errorf("parse startdate: %w", err)
		}
	}

	endDate := r.Form.Get("enddate")
	if len(endDate) > 0 {
		params.EndDate, err = time.Parse("2006-01-02", endDate)
		if err != nil {
			return fmt.Errorf("parse startdate: %w", err)
		}
	}

	// validation
	form := forms.New(r.PostForm)
	form.Required("email", "reason", "startdate", "enddate", "category")
	form.IsEmail("email")
	form.IsAfter("startdate", "enddate")

	// UniqueMail checks if the mail address is already registered
	if !form.IsEmpty("email") {
		uniquemail, _ := m.DB.CheckEmailUnique(params.Email)
		if uniquemail {
			form.Errors.Add("email", "E-Mail Addresse ist nicht registriert. Verwende die gleiche Addresse wie bei der Registrierung.")
		}
	}

	// check if duration is within allowed range
	var response logistics.RequestResponseContent
	response.Duration = calcPrice(params.StartDate, params.EndDate, 1)
	if response.Duration > 12 {
		form.Errors.Add("enddate", "Es sind maximal 12 Monate erlaubt.")
	}

	// get categories
	categories, err := m.DB.AllLogisticCategories()
	if err != nil {
		return fmt.Errorf("failed to get categories: %w", err)
	}

	// set chosen category to the whole category
	for _, cat := range categories {
		if cat.ID == params.Category.ID {
			params.Category = cat
		}
	}

	// check if visitor is member if selected "Kleingut"
	var visitor models.Visitor
	if !(len(form.Errors.Get("email")) > 0) {

		visitor, err = m.DB.VisitorByEmail(params.Email)
		if err != nil {
			return fmt.Errorf("get visitor: %w", err)
		}

		if params.Category.CatName == "Kleingut" {
			if !(visitor.Ressort.ID > 0) {
				form.Errors.Add("category", "Kleingut ist für Vereinsmitglieder reserviert. Wähle Halbe Palette oder ein grösseres Format.")
			}
		}
	}

	// what to do if form not valid
	if !form.Valid() {
		log.Println("not valid")

		// return form with errors and allready filled fields
		var errs logistics.RequestErrors
		form.MapErrors(&errs)
		return rendering(w, r, logistics.RequestForm(
			params,
			categories,
			errs,
		))
	}

	// build response
	response.FirstName = visitor.FirstName

	// calculate unit price
	var unitPrice float64

	if visitor.Ressort.ID > 0 {
		unitPrice = params.Category.PriceMember
	} else {
		unitPrice = params.Category.Price
	}

	// get estimated price
	response.EstPrice = calcPrice(params.StartDate, params.EndDate, unitPrice)

	// set price to true if it can be calculated
	if params.Category.ID < 4 {
		response.Price = true
	}

	// build logistic item
	newLogistic := models.Logistic{
		Category:       params.Category,
		StartDate:      params.StartDate,
		EndDate:        params.EndDate,
		Visitor:        visitor,
		Member:         visitor.Ressort.ID > 0,
		Reason:         params.Reason,
		RequestPending: true,
	}

	// number
	switch params.Category.CatName {
	case "Kleingut":
		newLogistic.Number = 200
	case "Sperrgut":
		newLogistic.Number = 300
	default:
		newLogistic.Number = 100
		newLogistic.UnitPrice = unitPrice
		newLogistic.Price = response.EstPrice
	}

	// create db entry
	err = m.DB.CreateLogistic(newLogistic)
	if err != nil {
		return fmt.Errorf("failed to create logistic: %w", err)
	} else {
		logger.Success(r, m.App.Session,
			fmt.Sprintf("logistic for visitor %s %s successfully requested", visitor.FirstName, visitor.FamilyName))
	}

	// create slack message
	err = godotenv.Load()
	if err != nil {
		logger.Fatal("loading .env file", slog.String("error", err.Error()))
	}

	channelID := os.Getenv("LOGISTIC_SLACK_CHANNEL")

	// construct message string out of all information
	message := fmt.Sprintf(
		"<!channel> *Neue Anfrage via QR-Code:*\n\n*Zeitpunkt*\n%s\n\n*Informationen*\n%s\n\n*Einheit*\n%s\n\n*Von*\n%s\n\n*Bis*\n%s\n\n*Dauer in Monaten*\n%s\n\n*Kontaktangaben*\n%s",
		time.Now().Format("01.02.2006 15:04"),
		params.Reason,
		params.Category.CatName,
		params.StartDate.Format("02.01.2006"),
		params.EndDate.Format("02.01.2006"),
		helpers.FormatFloat(response.Duration, 1),
		fmt.Sprintf(
			"%s %s: %s",
			visitor.FirstName,
			visitor.FamilyName,
			visitor.Email,
		),
	)

	err = m.App.Slack.SendSlack(channelID, message)
	if err != nil {
		return fmt.Errorf("creating slack message: %w", err)
	}

	// store confirmation data in session for PRG redirect
	m.App.Session.Put(r.Context(), "logistic_request_confirmation", models.LogisticRequestSession{
		FirstName:    response.FirstName,
		Email:        params.Email,
		Reason:       params.Reason,
		CategoryName: params.Category.CatName,
		StartDate:    params.StartDate,
		EndDate:      params.EndDate,
		Duration:     response.Duration,
		Price:        response.Price,
		EstPrice:     response.EstPrice,
	})

	return hxRedirect(w, r, "/logistics-request/confirmation")
}

// HandleLogisticsRequestConfirmation renders the confirmation page after a successful request
func (m *Repository) HandleLogisticsRequestConfirmation(w http.ResponseWriter, r *http.Request) error {
	session, ok := m.App.Session.Get(r.Context(), "logistic_request_confirmation").(models.LogisticRequestSession)
	if !ok {
		return hxRedirect(w, r, "/logistics-request")
	}

	m.App.Session.Remove(r.Context(), "logistic_request_confirmation")

	response := logistics.RequestResponseContent{
		FirstName: session.FirstName,
		Price:     session.Price,
		EstPrice:  session.EstPrice,
		Duration:  session.Duration,
	}

	params := logistics.RequestParams{
		Email:     session.Email,
		Reason:    session.Reason,
		StartDate: session.StartDate,
		EndDate:   session.EndDate,
	}
	params.Category.CatName = session.CategoryName

	return rendering(w, r, logistics.RequestConfirmationPage(response, params))
}
