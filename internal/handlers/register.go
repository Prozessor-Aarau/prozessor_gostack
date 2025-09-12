package handlers

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/faebut/prozessor_gostack/internal/forms"
	"github.com/faebut/prozessor_gostack/internal/logger"
	"github.com/faebut/prozessor_gostack/internal/models"
	"github.com/faebut/prozessor_gostack/view/register"
)

// HandleRegisterIndex handles the Register Page
func (m *Repository) HandleRegisterIndex(w http.ResponseWriter, r *http.Request) error {
	// render the template
	return rendering(w, r, register.Register())
}

// PostRegister is the post registration handler
func (m *Repository) HandleRegisterCreate(w http.ResponseWriter, r *http.Request) error {
	err := r.ParseForm()
	if err != nil {
		return fmt.Errorf("failed to parse form: %w", err)
	}

	var visitor register.RegisterParams

	bdate := r.Form.Get("birthdate")
	if len(bdate) > 0 {
		visitor.Birthdate, err = time.Parse("2006-01-02", bdate)
		if err != nil {
			return fmt.Errorf("failed to parse date: %w", err)
		}
	}

	visitor.FirstName = r.Form.Get("first_name")
	visitor.FamilyName = r.Form.Get("family_name")
	visitor.Address = r.Form.Get("address")
	visitor.Postcode = r.Form.Get("postcode")
	visitor.City = r.Form.Get("city")
	visitor.Email = strings.TrimSpace(r.Form.Get("email"))
	visitor.Phone = strings.TrimSpace(r.Form.Get("phone"))
	visitor.Infos = strings.TrimSpace(r.Form.Get("infos"))

	// validation
	form := forms.New(r.PostForm)
	form.Required("first_name", "family_name", "birthdate", "address", "postcode", "city")
	form.OneOf("email", "phone")
	form.IsEmail("email")
	form.IsPhone("phone")

	// UniqueMail checks if the mail address is unique
	if !form.IsEmpty("email") {
		validmail, _ := m.DB.CheckEmailUnique(visitor.Email)
		if !validmail {
			form.Errors.Add("email", "Email Adresse wurde schon verwendet")
		}
	}

	// reformat to real bool and add to visitor model
	visitor.Newsletter = form.BoxValue("newsletter")

	// check consent box and add error if necessary
	visitor.Consent = form.BoxValue("consent")
	if !visitor.Consent {
		form.Errors.Add("consent", "Checkbox muss ausgewählt werden")
	}

	// prepare data to send it to the templates
	data := make(map[string]any)
	data["visitor"] = visitor

	// what to do if form not valid
	if !form.Valid() {
		// return form with errors and allready filled fields
		return rendering(w, r, register.RegisterForm(
			visitor,
			register.RegisterErrors{
				FirstName:  form.Errors.Get("first_name"),
				FamilyName: form.Errors.Get("family_name"),
				Birthdate:  form.Errors.Get("birthdate"),
				Phone:      form.Errors.Get("phone"),
				Email:      form.Errors.Get("email"),
				Address:    form.Errors.Get("address"),
				Postcode:   form.Errors.Get("postcode"),
				City:       form.Errors.Get("city"),
				Consent:    form.Errors.Get("consent"),
			},
		))
	}

	// what to do if form valid
	// post new visitor to db
	newVisitorID, err := m.DB.CreateVisitor(models.Visitor{
		FirstName:  visitor.FirstName,
		FamilyName: visitor.FamilyName,
		Birthdate:  visitor.Birthdate,
		Phone:      visitor.Phone,
		Email:      visitor.Email,
		Address:    visitor.Address,
		Postcode:   visitor.Postcode,
		City:       visitor.City,
		Infos:      visitor.Infos,
		Newsletter: visitor.Newsletter,
	})
	if err != nil {
		return fmt.Errorf("failed to create visitor: %w", err)
	}

	// log success
	logger.Success(r, m.App.Session,
		fmt.Sprintf("new visitor %s %s successfully created with id: %s", visitor.FirstName, visitor.FamilyName, newVisitorID))

	return hxRedirect(w, r, "/registersuccess?name="+visitor.FirstName)
}

// HandleRegisterSuccess handles the Register Success Page
func (m *Repository) HandleRegisterSuccess(w http.ResponseWriter, r *http.Request) error {
	// get param from Query to pass it on
	name := r.URL.Query().Get("name")

	// render the template
	return rendering(w, r, register.RegisterSuccess(name))
}
