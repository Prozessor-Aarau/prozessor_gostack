package handlers

import (
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/faebut/prozessor_gostack/internal/forms"
	"github.com/faebut/prozessor_gostack/internal/models"
	"github.com/faebut/prozessor_gostack/internal/render"
)

// Register is the register page handler
func (m *Repository) Register(w http.ResponseWriter, r *http.Request) {
	visitor := models.Visitor{}
	visitor.Newsletter = true

	data := make(map[string]interface{})
	data["visitor"] = visitor

	// send the data to the template
	render.Template(w, r, "selfregister.html", &models.TemplateData{
		Form: forms.New(nil),
		Data: data,
	})
}

// PostRegister is the post registration handler
func (m *Repository) PostRegister(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "kann Formular nicht verarbeiten")
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	var visitor models.Visitor

	bdate := r.Form.Get("birthdate")
	if len(bdate) > 0 {
		visitor.Birthdate, err = time.Parse("2006-01-02", bdate)
		if err != nil {
			log.Println("Error parsing birthdate: ", err)
			return
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
	consent := form.BoxValue("consent")
	if !consent {
		form.Errors.Add("consent", "Checkbox muss ausgewählt werden")
	}

	// prepare data to send it to the templates
	data := make(map[string]interface{})
	data["visitor"] = visitor

	// what to do if form not valid
	if !form.Valid() {
		data["consent"] = consent

		// send the data to the template
		w.WriteHeader(http.StatusSeeOther)
		render.Template(w, r, "selfregister.html", &models.TemplateData{
			Form: form,
			Data: data,
		})
		return
	}

	// what to do if form valid
	newVisitorID, err := m.DB.CreateVisitor(visitor)
	if err != nil {
		log.Println(err)
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	log.Println("Visitor inserted, new ID: ", newVisitorID)
	// send the data to the template
	w.WriteHeader(http.StatusSeeOther)
	render.Template(w, r, "selfregistersuccess.html", &models.TemplateData{
		Data: data,
	})

}
