package handlers

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/faebut/prozessor_gostack/internal/forms"
	"github.com/faebut/prozessor_gostack/internal/models"
	"github.com/faebut/prozessor_gostack/view/auth"
	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
)

// GetLoginIndex is the login index page handler
func (m *Repository) HandleLoginIndex(w http.ResponseWriter, r *http.Request) error {
	// render the template
	return rendering(w, r, auth.Login())
}

// HandleLoginCreate handles the login form post request
func (m *Repository) HandleLoginCreate(w http.ResponseWriter, r *http.Request) error {
	_ = m.App.Session.RenewToken(r.Context())

	err := r.ParseForm()
	if err != nil {
		log.Println(err)
	}

	email := r.Form.Get("email")
	password := r.Form.Get("password")

	// check form fields
	form := forms.New(r.PostForm)
	form.Required("email", "password")
	form.IsEmail("email")

	if !form.Valid() {
		return rendering(w, r, auth.LoginForm(
			auth.LoginParams{
				Email:    email,
				Password: password,
			},
			auth.LoginErrors{
				Email:    form.Errors.Get("email"),
				Password: form.Errors.Get("password"),
			},
		))
	}

	id, visitor, err := m.DB.Authenticate(email, password)
	if err != nil {
		log.Println(err)

		// send the data to the template
		return rendering(w, r, auth.LoginForm(
			auth.LoginParams{
				Email:    email,
				Password: password,
			},
			auth.LoginErrors{
				Email:              form.Errors.Get("email"),
				Password:           form.Errors.Get("password"),
				InvalidCredentials: "E-Mail Adresse oder Passwort ist falsch",
			},
		))
	}

	m.App.Session.Put(r.Context(), "user_id", id)
	m.App.Session.Put(r.Context(), "visitor", visitor)
	m.App.Session.Put(r.Context(), "access", visitor.RoleID)
	m.App.Session.Put(r.Context(), "flash", "Logged in successfully")

	return hxRedirect(w, r, "/admin")
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

	// handling roles
	visitor.RoleID = 1
	rIDString := r.Form.Get("role")
	if len(rIDString) > 0 {
		roleID, err := strconv.Atoi(rIDString)
		if err != nil {
			fmt.Fprintf(
				w,
				`<span class="label-text text-error">Probleme beim generieren des Codes: %s</span>`,
				err,
			)
			return
		}
		visitor.RoleID = roleID
	}

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

	err = godotenv.Load()
	if err != nil {
		log.Fatal("error loading .env file")
	}

	baseUrl := os.Getenv("BASE_URL")

	registerLink := fmt.Sprintf(
		"%s/createlogin?email=%s&uid=%s",
		baseUrl,
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

// HandleCreateLoginIndex handles the CreateLogin index page
func (m *Repository) HandleCreateLoginIndex(w http.ResponseWriter, r *http.Request) error {
	email := r.URL.Query().Get("email")
	uid := r.URL.Query().Get("uid")

	// prepare data to send it to the templates
	var errors auth.CreateLoginErrors
	var params auth.CreateLoginParams

	visitor, err := m.DB.CheckCreateLogin(email, uid)
	if err != nil {
		errors.AppError = "Code nicht mehr gültig. Bitte wende dich an den Vorstand!"
	}

	params.FirstName = visitor.FirstName
	params.Email = visitor.Email

	return rendering(w, r, auth.CreateLogin(params, errors))
}

// HandleCreateLogin handles the CreateLogin post form
func (m *Repository) HandleCreateLogin(w http.ResponseWriter, r *http.Request) error {

	// validation
	form := forms.New(r.PostForm)
	form.Required("password")
	form.Required("password_check")
	form.IsPassword("password")
	form.IsMatch("password_check", "password")

	var visitor auth.CreateLoginParams

	visitor.Email = r.Form.Get("email")
	visitor.Password = r.Form.Get("password")
	visitor.FirstName = r.Form.Get("first_name")

	// prepare data to send it to the templates
	var errors auth.CreateLoginErrors

	// what to do if form not valid
	if !form.Valid() {
		errors.Password = form.Errors.Get("password")
		errors.PasswordCheck = form.Errors.Get("password_check")
		// send the data to the template
		return rendering(w, r, auth.CreateLoginForm(visitor, errors))
	}

	hashedPassword, err := bcrypt.GenerateFromPassword(
		[]byte(visitor.Password),
		bcrypt.DefaultCost,
	)
	if err != nil {
		errors.AppError = "Error creating password hash"
		// send the data to the template
		return rendering(w, r, auth.CreateLoginForm(visitor, errors))
	}
	var newVisitor models.Visitor
	newVisitor.Email = visitor.Email
	newVisitor.Password = string(hashedPassword)

	// put new password in db and set to login
	err = m.DB.CreatePassword(newVisitor)
	if err != nil {
		errors.AppError = "Error putting password in db"
		// send the data to the template
		return rendering(w, r, auth.CreateLoginForm(visitor, errors))
	}

	// redirect
	return hxRedirect(w, r, "/admin")
}

// Logout logs a user out
func (m *Repository) Logout(w http.ResponseWriter, r *http.Request) {
	_ = m.App.Session.Destroy(r.Context())
	_ = m.App.Session.RenewToken(r.Context())

	http.Redirect(w, r, "/login", http.StatusSeeOther)
}
