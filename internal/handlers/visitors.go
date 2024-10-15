package handlers

import (
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/faebut/prozessor_gostack/internal/forms"
	"github.com/faebut/prozessor_gostack/internal/models"
	"github.com/faebut/prozessor_gostack/internal/render"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// Visitors is the visitors page handler
func (m *Repository) Visitors(w http.ResponseWriter, r *http.Request) {
	user, ok := m.App.Session.Get(r.Context(), "visitor").(models.Visitor)
	if !ok {
		log.Println("could not get visitor from Session")
		m.App.Session.Put(r.Context(), "error", "Can't get visitor from session")
		http.Redirect(w, r, "/admin/logout", http.StatusSeeOther)
		return
	}

	data := make(map[string]interface{})
	data["user"] = user

	render.Template(w, r, "admin_visitors.html", &models.TemplateData{
		Data: data,
	})
}

// GetVisitors is the handler to get the visitors list
func (m *Repository) GetVisitors(w http.ResponseWriter, r *http.Request) {

	// set initial params
	page := 1
	limit := 10
	filter := models.NewVisitorSearch("")
	if r.URL.Query().Has("search") {
		filter = models.NewVisitorSearch(r.URL.Query().Get("search"))
	}
	filter.Sort = "family_name"
	filter.SortDir = "ASC"

	// check if a page param has been submitted and change to this value
	if r.URL.Query().Has("page") {
		page, _ = strconv.Atoi(r.URL.Query().Get("page"))
	}
	if r.URL.Query().Has("sort") {
		filter.Sort = r.URL.Query().Get("sort")
	}
	if r.URL.Query().Has("sortdir") {
		filter.SortDir = strings.ToUpper(r.URL.Query().Get("sortdir"))
	}
	if r.URL.Query().Has("abo") {
		filter.Abo, _ = strconv.ParseBool(r.URL.Query().Get("abo"))
	}
	if r.URL.Query().Has("valid") {
		filter.AboValidity = r.URL.Query().Get("valid")
	}
	if r.URL.Query().Has("member") {
		filter.Member, _ = strconv.ParseBool(r.URL.Query().Get("member"))
	}
	if r.URL.Query().Has("partner") {
		filter.Partner, _ = strconv.ParseBool(r.URL.Query().Get("partner"))
	}
	if r.URL.Query().Has("badge") {
		filter.Badge, _ = strconv.ParseBool(r.URL.Query().Get("badge"))
	}

	visitorList, count, err := m.DB.AllVisitors(page, limit, filter)
	if err != nil {
		log.Println("could not get visitors from DB: ", err)
		m.App.Session.Put(r.Context(), "error", "Can't get visitors from database")
	}

	data := make(map[string]interface{})
	data["visitorlist"] = visitorList
	data["count"] = count
	data["pages"] = int(math.Ceil(float64(count) / float64(limit)))
	data["onPage"] = page
	data["filter"] = filter

	render.Template(w, r, "part_visitorlist.html", &models.TemplateData{
		Data: data,
	})
}

// PostGetVisitors gets the filter values and returns the filtered visitor list
func (m *Repository) PostGetVisitors(w http.ResponseWriter, r *http.Request) {

	// set initial page param
	page := 1
	limit := 10
	// check if a page param has been submitted and change to this value
	if r.URL.Query().Has("page") {
		page, _ = strconv.Atoi(r.URL.Query().Get("page"))
	}
	searchTerm := r.Form.Get("search_term")

	form := forms.New(r.PostForm)
	filter := models.NewVisitorSearch(searchTerm)
	filter.Sort = r.Form.Get("sort")
	filter.SortDir = r.Form.Get("sort_dir")
	filter.Abo = form.BoxValue("hasabo")
	filter.AboValidity = r.Form.Get("abovalidity")
	filter.Badge = form.BoxValue("badge")
	filter.Member = form.BoxValue("member")
	filter.Partner = form.BoxValue("partner")

	visitorList, count, err := m.DB.AllVisitors(page, limit, filter)
	if err != nil {
		log.Println("could not get visitors from DB: ", err)
		m.App.Session.Put(r.Context(), "error", "Can't get visitors from database")
	}

	data := make(map[string]interface{})
	data["visitorlist"] = visitorList
	data["count"] = count
	data["pages"] = int(math.Ceil(float64(count) / float64(limit)))
	data["onPage"] = page
	data["filter"] = filter

	render.Template(w, r, "part_visitorlist.html", &models.TemplateData{
		Data: data,
	})
}

// VisitorsSelectBox returns a select box with the filtered and selected visitors
func (m *Repository) VisitorsSelectBox(w http.ResponseWriter, r *http.Request) {
	visID := chi.URLParam(r, "id")

	searchTerm := r.Form.Get("search_term")

	filter := models.NewVisitorSearch(searchTerm)

	// get filtered visitors from DB
	visitors, _, err := m.DB.AllVisitors(1, 0, filter)
	if err != nil {
		log.Println("could not get visitors from DB: ", err)
		m.App.Session.Put(r.Context(), "error", "Can't get visitors from database")
	}

	data := make(map[string]interface{})
	data["visitors"] = visitors
	data["visid"] = visID

	render.Template(w, r, "comp_visitors_selectbox.html", &models.TemplateData{
		Data: data,
	})
}

// EditVisitor is the edit visitor page handler
func (m *Repository) EditVisitor(w http.ResponseWriter, r *http.Request) {
	// set initial values
	admin := false
	super := false
	user, ok := m.App.Session.Get(r.Context(), "visitor").(models.Visitor)
	if !ok {
		log.Println("could not get user from Session")
		m.App.Session.Put(r.Context(), "error", "Can't get user from session")
		http.Redirect(w, r, "/admin/logout", http.StatusSeeOther)
		return
	}

	if user.RoleID >= 3 {
		admin = true
	}

	if user.RoleID >= 4 {
		super = true
	}

	userID := chi.URLParam(r, "id")

	// Get the queryparams to send back to the right page
	var ref string
	if r.URL.Query().Has("ref") {
		ref = r.URL.Query().Get("ref")
	} else {
		ref = ""
	}

	// if there has been a id provided, set user ID to that. Else take the one from the logged in user
	if len(userID) == 0 {
		userID = user.ID
	}

	// get visitor to edit from DB
	visitor, err := m.DB.VisitorByID(userID)
	if err != nil {
		log.Println("could not get visitor from DB: ", err)
		m.App.Session.Put(r.Context(), "error", "Can't get visitors from database")
	}

	// get ressorts from DB
	ressorts, err := m.DB.AllRessorts()
	if err != nil {
		log.Println("could not get ressorts from DB: ", err)
		m.App.Session.Put(r.Context(), "error", "Can't get ressorts from database")
	}
	// get roles from DB
	roles, err := m.DB.AllRoles()
	if err != nil {
		log.Println("could not get roles from DB: ", err)
	}

	exploded := strings.Split(r.RequestURI, "/")
	route := exploded[2]

	data := make(map[string]interface{})
	data["user"] = user
	data["visitor"] = visitor
	data["route"] = route
	data["ref"] = ref
	data["admin"] = admin
	data["super"] = super
	data["roles"] = roles
	data["ressorts"] = ressorts

	render.Template(w, r, "admin_profile.html", &models.TemplateData{
		Data: data,
		Form: forms.New(nil),
	})
}

// UpdateProfile updates the visitors profile
func (m *Repository) PostEditVisitor(w http.ResponseWriter, r *http.Request) {
	// initial values
	admin := false
	user, ok := m.App.Session.Get(r.Context(), "visitor").(models.Visitor)
	if !ok {
		log.Println("could not get visitor from Session")
		m.App.Session.Put(r.Context(), "error", "Can't get visitor from session")
		http.Redirect(w, r, "/admin/logout", http.StatusSeeOther)
		return
	}

	if user.RoleID >= 3 {
		admin = true
	}

	// get visitor from DB for comparison
	userID := chi.URLParam(r, "id")

	// if there has been a id provided, set user ID to that. Else take the one from the logged in user
	if len(userID) == 0 {
		userID = user.ID
	}

	// get visitor to edit from DB
	visitor, err := m.DB.VisitorByID(userID)
	if err != nil {
		log.Println("could not get visitor from DB: ", err)
		m.App.Session.Put(r.Context(), "error", "Can't get visitors from database")
	}

	// parse form and check for size of picture
	err = r.ParseMultipartForm(2 << 20)
	if err != nil {
		log.Println("Error parsing multipart form: ", err)
		m.App.Session.Put(r.Context(), "error", "kann Formular nicht verarbeiten")
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// parse birthdate to work with formfield and convert to string in visitor struct
	bdate := r.Form.Get("birthdate")
	if len(bdate) > 0 {
		visitor.Birthdate, err = time.Parse("2006-01-02", bdate)
		if err != nil {
			log.Println("Error parsing birthdate: ", err)
			return
		}
	}

	// get ressorts from DB
	ressorts, err := m.DB.AllRessorts()
	if err != nil {
		log.Println("could not get ressorts from DB: ", err)
		m.App.Session.Put(r.Context(), "error", "Can't get ressorts from database")
	}

	// get default fields
	ref := r.Form.Get("ref")
	route := r.Form.Get("route")

	// get editable fields
	visitor.FirstName = r.Form.Get("first_name")
	visitor.FamilyName = r.Form.Get("family_name")
	visitor.Address = r.Form.Get("address")
	visitor.Postcode = r.Form.Get("postcode")
	visitor.City = r.Form.Get("city")
	newEmail := strings.TrimSpace(r.Form.Get("email"))
	visitor.Phone = strings.TrimSpace(r.Form.Get("phone"))
	visitor.Infos = strings.TrimSpace(r.Form.Get("infos"))

	ressortString := r.Form.Get("ressort")

	if len(ressortString) > 0 {
		ressort, err := strconv.Atoi(ressortString)
		if err != nil {
			log.Println("error converting string to int", err)
		}
		visitor.Ressort.ID = ressort
	}

	// validation
	form := forms.New(r.PostForm)
	form.Required("first_name", "family_name", "birthdate", "address", "postcode", "city")
	form.OneOf("email", "phone")
	form.IsEmail("email")
	form.IsPhone("phone")
	form.IsPassword("password")
	form.IsMatch("password_check", "password")

	// // UniqueMail checks if the mail address is unique
	if !form.IsEmpty("email") {
		if visitor.Email != newEmail {
			validmail, _ := m.DB.CheckEmailUnique(newEmail)
			if !validmail {
				form.Errors.Add("email", "Email Adresse wurde schon verwendet")
			}
			visitor.Email = newEmail
		}
	}

	// reformat to real bool and add to visitor model
	visitor.Newsletter = form.BoxValue("newsletter")

	// check for new avatar
	// Retrieve the file from form data
	file, handler, err := r.FormFile("avatar")
	if err != nil {
		if err != http.ErrMissingFile {
			log.Println("Error retrieving file")
			m.App.Session.Put(r.Context(), "error", "Error retrieving File")
			http.Redirect(w, r, "/admin/profile", http.StatusSeeOther)
			return
		}
	} else {
		defer file.Close()

		// Generate a unique filename to prevent overwriting and conflicts
		uuid, err := uuid.NewRandom()
		if err != nil {
			log.Println("Error generating uuid")
			m.App.Session.Put(r.Context(), "error", "Error generating uuid")
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		filename := uuid.String() + filepath.Ext(handler.Filename) // Append the file extension

		// Create the full path for saving the file
		filePath := filepath.Join("static/uploads", filename)

		// Save the file to the server
		dst, err := os.Create(filePath)
		if err != nil {
			log.Println("error saving File")
			form.Errors.Add("avatar", "Error saving file createpath")
			http.Redirect(w, r, "/admin/profile", http.StatusSeeOther)
			return
		}
		defer dst.Close()
		if _, err = io.Copy(dst, file); err != nil {
			log.Println("error saving File copyfile")
			form.Errors.Add("avatar", "Error saving file")
			http.Redirect(w, r, "/admin/profile", http.StatusSeeOther)
			return
		}

		//Delete current image from the initial fetch of the user
		if visitor.Avatar != "" {
			oldAvatarPath := filepath.Join("static/uploads", visitor.Avatar)

			//Check if the oldPath is not the same as the new path
			if oldAvatarPath != filePath {
				if err := os.Remove(oldAvatarPath); err != nil {
					log.Println("Warning: failed to delete old avatar file: ", err)
				}
			}
		}

		// Update the user's avatar in the visitor struct, it will be passed with the rest
		visitor.Avatar = filename
	}

	// delete avatar if toggle is there
	if form.BoxValue("delete_avatar") {
		AvatarPath := filepath.Join("static/uploads", visitor.Avatar)
		if err := os.Remove(AvatarPath); err != nil {
			log.Printf("Warning: failed to delete avatar file: %s\n", err)
		}
		visitor.Avatar = ""
	}

	// prepare data to send it to the templates
	data := make(map[string]interface{})
	data["visitor"] = visitor
	data["user"] = user
	data["ref"] = ref
	data["route"] = route
	data["ressorts"] = ressorts
	data["admin"] = admin

	// what to do if form not valid
	if !form.Valid() {
		// send the data to the template
		w.WriteHeader(http.StatusSeeOther)
		render.Template(w, r, "admin_profile.html", &models.TemplateData{
			Form: form,
			Data: data,
		})
		return
	}

	// if checks passed get new password if the field is not empty
	pw := r.Form.Get("password")
	if len(pw) > 7 {
		hashedPassword, err := bcrypt.GenerateFromPassword(
			[]byte(pw),
			bcrypt.DefaultCost,
		)
		if err != nil {
			log.Println("Error creating hash :", err)
			return
		}
		visitor.Password = string(hashedPassword)
		visitor.Login = true
	}

	// what to do if form valid
	err = m.DB.UpdateVisitor(visitor)
	if err != nil {
		log.Println(err)
		http.Redirect(w, r, "/admin", http.StatusSeeOther)
		return
	}

	// get where user came from and set redirect path to this place
	redirectPath := "/admin"
	if ref != "" {
		redirectPath = "/admin/" + ref
	}

	// on success redirect
	http.Redirect(w, r, redirectPath, http.StatusSeeOther)
}
