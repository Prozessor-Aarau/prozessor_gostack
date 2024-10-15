package handlers

import (
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/faebut/prozessor_gostack/internal/config"
	"github.com/faebut/prozessor_gostack/internal/driver"
	"github.com/faebut/prozessor_gostack/internal/forms"
	"github.com/faebut/prozessor_gostack/internal/models"
	"github.com/faebut/prozessor_gostack/internal/render"
	"github.com/faebut/prozessor_gostack/internal/repository"
)

// Repo the repository used by the handlers
var Repo *Repository

// Repository is the repository type
type Repository struct {
	App *config.AppConfig
	DB  repository.DatabaseRepo
}

// NewRepo creates a new repository
func NewRepo(a *config.AppConfig, db *driver.DB) *Repository {
	return &Repository{
		App: a,
		DB:  repository.NewPostgresRepo(db.SQL, a),
	}
}

// NewHandlers sets the repository for the handlers
func NewHandlers(r *Repository) {
	Repo = r
}

// Home is the home page handler
func (m *Repository) Home(w http.ResponseWriter, r *http.Request) {
	// redirect
	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}

// Admin is the admin page handler
func (m *Repository) Admin(w http.ResponseWriter, r *http.Request) {
	// send the data to the template
	user, ok := m.App.Session.Get(r.Context(), "visitor").(models.Visitor)
	if !ok {
		log.Println("could not get visitor from Session")
		m.App.Session.Put(r.Context(), "error", "Can't get visitor from session")

		http.Redirect(w, r, "/admin/logout", http.StatusSeeOther)
		return
	}

	data := make(map[string]interface{})
	data["user"] = user

	render.Template(w, r, "admin.html", &models.TemplateData{
		Data: data,
	})
}

// Login is the login page handler
func (m *Repository) Login(w http.ResponseWriter, r *http.Request) {
	// send the data to the template
	render.Template(w, r, "login.html", &models.TemplateData{
		Form: forms.New(nil),
	})
}

// PostLogin is the home page handler
func (m *Repository) PostLogin(w http.ResponseWriter, r *http.Request) {
	_ = m.App.Session.RenewToken(r.Context())

	err := r.ParseForm()
	if err != nil {
		log.Println(err)
	}

	email := r.Form.Get("email")
	password := r.Form.Get("password")

	// prepare data to send it to the templates
	data := make(map[string]interface{})

	form := forms.New(r.PostForm)
	form.Required("email", "password")
	form.IsEmail("email")

	data["email"] = email
	if !form.Valid() {
		render.Template(w, r, "login.html", &models.TemplateData{
			Form: form,
			Data: data,
		})
		return
	}

	id, visitor, err := m.DB.Authenticate(email, password)
	if err != nil {
		log.Println(err)

		// what to do if form not valid
		data["error"] = "E-Mail Adresse oder Passwort falsch"

		// send the data to the template
		w.WriteHeader(http.StatusSeeOther)
		render.Template(w, r, "login.html", &models.TemplateData{
			Form: form,
			Data: data,
		})
		return
	}

	m.App.Session.Put(r.Context(), "user_id", id)
	m.App.Session.Put(r.Context(), "visitor", visitor)
	m.App.Session.Put(r.Context(), "access", visitor.RoleID)
	m.App.Session.Put(r.Context(), "flash", "Logged in successfully")
	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}

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
