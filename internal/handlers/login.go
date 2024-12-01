package handlers

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/faebut/prozessor_gostack/internal/forms"
	"github.com/faebut/prozessor_gostack/internal/models"
	"github.com/faebut/prozessor_gostack/internal/render"
	"golang.org/x/crypto/bcrypt"
)

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

// CreateLoginMail creates a new UUID saves it to the db and sends a Mail with
// a link to the new user
func (m *Repository) CreateLoginMail(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "kann Formular nicht verarbeiten")
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	var visitor models.Visitor

	visitor.FirstName = r.Form.Get("first_name")
	visitor.FamilyName = r.Form.Get("family_name")
	visitor.Email = strings.TrimSpace(r.Form.Get("email"))

	// validation
	if len(visitor.Email) < 3 {
		fmt.Fprint(
			w,
			`<span class="label-text text-error">Zuerst muss eine E-Mail adresse registriert sein!!</span>`,
		)
		return
	}

	uid, err := m.DB.CreateRegisterCode(visitor)
	if err != nil {
		fmt.Fprintf(
			w,
			`<span class="label-text text-error">Probleme beim generieren des Codes: %s</span>`,
			err,
		)
		return
	}

	byteTemplate, err := os.ReadFile("./email-templates/register-link.html")
	if err != nil {
		log.Fatal(err)
	}

	registerLink := fmt.Sprintf(
		"http://localhost:4000/createlogin?email=%s&uid=%s",
		visitor.Email,
		uid,
	)

	emailTemplate := string(byteTemplate)
	msgToSend := strings.Replace(emailTemplate, "[%firstname%]", visitor.FirstName, 1)
	msgToSend = strings.Replace(msgToSend, "[%link%]", registerLink, 1)

	msg := models.MailData{
		From:    "info@prozessor-aarau.ch",
		To:      visitor.Email,
		Subject: "Prozessor Registrierungs-Link",
		Content: msgToSend,
	}

	m.App.MailChan <- msg

	log.Printf("E-Mail an %s %s versendet", visitor.FirstName, visitor.FamilyName)
	// send the data to the template
	fmt.Fprint(
		w,
		`<span class="label-text text-success">Login Mail versendet!</span>`,
	)

}

// Login is the login page handlerh http.HandlerFunc
func (m *Repository) CreateLogin(w http.ResponseWriter, r *http.Request) {
	email := r.URL.Query().Get("email")
	uid := r.URL.Query().Get("uid")

	// prepare data to send it to the templates
	data := make(map[string]interface{})

	visitor, err := m.DB.CheckCreateLogin(email, uid)
	if err != nil {
		data["error"] = "Code nicht mehr gültig. Bitte wende dich an den Vorstand!"
	}

	data["visitor"] = visitor

	// send the data to the template
	render.Template(w, r, "admin_createlogin.html", &models.TemplateData{
		Data: data,
		Form: forms.New(nil),
	})
}

// PostLogin is the post login page handler
func (m *Repository) PostCreateLogin(w http.ResponseWriter, r *http.Request) {

	// validation
	form := forms.New(r.PostForm)
	form.Required("password")
	form.Required("password_check")
	form.IsPassword("password")
	form.IsMatch("password_check", "password")

	var visitor models.Visitor

	visitor.Email = r.Form.Get("email")
	visitor.Password = r.Form.Get("password")
	visitor.FirstName = r.Form.Get("first_name")
	visitor.FamilyName = r.Form.Get("family_name")

	// prepare data to send it to the templates
	data := make(map[string]interface{})

	// what to do if form not valid
	if !form.Valid() {
		// send the data to the template
		w.WriteHeader(http.StatusSeeOther)
		render.Template(w, r, "admin_createlogin.html", &models.TemplateData{
			Form: form,
			Data: data,
		})
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword(
		[]byte(visitor.Password),
		bcrypt.DefaultCost,
	)
	if err != nil {
		log.Println("Error creating hash :", err)
		return
	}
	visitor.Password = string(hashedPassword)
	visitor.Login = true

	// put new password in db and set to login
	err = m.DB.CreatePassword(visitor)
	if err != nil {
		data["error"] = fmt.Sprint("Error putting password in db:", err)
	}

	// what to do if form not valid
	if data["error"] != nil {
		// send the data to the template
		w.WriteHeader(http.StatusSeeOther)
		render.Template(w, r, "admin_createlogin.html", &models.TemplateData{
			Data: data,
			Form: form,
		})
		return
	}

	// redirect
	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}

// Logout logs a user out
func (m *Repository) Logout(w http.ResponseWriter, r *http.Request) {
	_ = m.App.Session.Destroy(r.Context())
	_ = m.App.Session.RenewToken(r.Context())

	http.Redirect(w, r, "/login", http.StatusSeeOther)
}
