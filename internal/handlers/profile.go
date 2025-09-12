package handlers

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/faebut/prozessor_gostack/internal/forms"
	"github.com/faebut/prozessor_gostack/internal/logger"
	"github.com/faebut/prozessor_gostack/internal/models"
	"github.com/faebut/prozessor_gostack/view/profile"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// HandleProfileIndex is the profile/visitor edit page handler
func (m *Repository) HandleProfileIndex(w http.ResponseWriter, r *http.Request) error {
	visitor, ok := m.App.Session.Get(r.Context(), "visitor").(models.Visitor)
	if !ok {
		slog.Error("could not get visitor from session", "path", r.URL.Path)
		m.App.Session.Put(r.Context(), "error", "Can't get visitor from session")

		return hxRedirect(w, r, "/admin/logout")
	}

	// get id of the profile, that we want to edit
	profID := chi.URLParam(r, "id")

	// Get the queryparams to send back to the right page
	var referer string
	if r.URL.Query().Has("ref") {
		referer = r.URL.Query().Get("ref")
	} else {
		referer = ""
	}

	// if there has been a id provided, set user ID to that. Else take the one from the logged in user
	if len(profID) == 0 {
		profID = visitor.ID
	}

	// get visitor to edit from DB
	profileToEdit, err := m.DB.VisitorByID(profID)
	if err != nil {
		return fmt.Errorf("failed to get visitor: %w", err)
	}

	// get ressorts from DB
	ressorts, err := m.DB.AllRessorts()
	if err != nil {
		return fmt.Errorf("failed to get ressorts: %w", err)
	}

	// get roles from DB
	roles, err := m.DB.AllRoles()
	if err != nil {
		return fmt.Errorf("failed to get roles: %w", err)
	}

	// get route from request url
	exploded := strings.Split(r.RequestURI, "/")
	route := exploded[2]

	return rendering(w, r, profile.Profile(
		visitor,
		profileToEdit,
		roles,
		ressorts,
		route,
		referer,
	))
}

// UpdateProfile updates the visitors profile
func (m *Repository) HandleProfileEdit(w http.ResponseWriter, r *http.Request) error {
	visitor, ok := m.App.Session.Get(r.Context(), "visitor").(models.Visitor)
	if !ok {
		slog.Error("could not get visitor from session", "path", r.URL.Path)
		m.App.Session.Put(r.Context(), "error", "Can't get visitor from session")

		return hxRedirect(w, r, "/admin/logout")
	}

	// prepare for formErrors
	var formErrors profile.ProfileErrors

	// get id of the profile, that we want to edit
	profID := chi.URLParam(r, "id")

	// if there has been a id provided, set user ID to that. Else take the one from the logged in user
	if len(profID) == 0 {
		profID = visitor.ID
	}

	// get visitor to edit from DB
	profileToEdit, err := m.DB.VisitorByID(profID)
	if err != nil {
		return fmt.Errorf("failed to get visitor: %w", err)
	}

	// parse form and check for size of picture
	err = r.ParseMultipartForm(2 << 20)
	if err != nil {
		return fmt.Errorf("failed to parse multipart form: %w", err)
	}

	// parse birthdate to work with formfield and convert to string in visitor struct
	bdate := r.Form.Get("birthdate")
	if len(bdate) > 0 {
		profileToEdit.Birthdate, err = time.Parse("2006-01-02", bdate)
		if err != nil {
			return fmt.Errorf("failed to parse birthdate: %w", err)
		}
	}

	// get ressorts from DB
	ressorts, err := m.DB.AllRessorts()
	if err != nil {
		return fmt.Errorf("failed to get ressorts: %w", err)
	}

	// get roles from DB
	roles, err := m.DB.AllRoles()
	if err != nil {
		return fmt.Errorf("failed to get roles: %w", err)
	}

	// get default fields
	referer := r.Form.Get("ref")
	route := r.Form.Get("route")

	// get editable fields
	profileToEdit.FirstName = r.Form.Get("first_name")
	profileToEdit.FamilyName = r.Form.Get("family_name")
	profileToEdit.Address = r.Form.Get("address")
	profileToEdit.Postcode = r.Form.Get("postcode")
	profileToEdit.City = r.Form.Get("city")
	newEmail := strings.TrimSpace(r.Form.Get("email"))
	profileToEdit.Phone = strings.TrimSpace(r.Form.Get("phone"))
	profileToEdit.Infos = strings.TrimSpace(r.Form.Get("infos"))

	ressort, err := strconv.Atoi(r.Form.Get("ressort"))
	if err != nil {
		return fmt.Errorf("failed to convert string to int: %w", err)
	}
	profileToEdit.Ressort.ID = ressort

	roleID, err := strconv.Atoi(r.Form.Get("role"))
	if err != nil {
		return fmt.Errorf("failed to convert string to int: %w", err)
	}
	profileToEdit.RoleID = roleID

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
		if profileToEdit.Email != newEmail {
			validmail, _ := m.DB.CheckEmailUnique(newEmail)
			if !validmail {
				form.Errors.Add("email", "Email Adresse wurde schon verwendet")
			}
			profileToEdit.Email = newEmail
		}
	}

	// reformat to real bool and add to visitor model
	profileToEdit.Newsletter = form.BoxValue("newsletter")

	// check for new avatar
	// Retrieve the file from form data
	file, handler, err := r.FormFile("avatar")
	if err != nil {
		if err != http.ErrMissingFile {
			slog.Error(fmt.Errorf("failed to retrieve file: %w", err).Error())
			formErrors.AppError = "Error retrieving file"
			return rendering(w, r, profile.ProfileForm(
				visitor,
				profileToEdit,
				formErrors,
				roles,
				ressorts,
				route,
				referer,
			))
		}
	} else {
		defer file.Close()

		// Generate a unique filename to prevent overwriting and conflicts
		uuid, err := uuid.NewRandom()
		if err != nil {
			slog.Error(fmt.Errorf("failed to generate uuid: %w", err).Error())
			formErrors.AppError = "Error generating uuid"
			return rendering(w, r, profile.ProfileForm(
				visitor,
				profileToEdit,
				formErrors,
				roles,
				ressorts,
				route,
				referer,
			))
		}
		filename := uuid.String() + filepath.Ext(handler.Filename) // Append the file extension

		// check if uploads directory exists, if not create it
		if _, err := os.Stat("static/uploads"); err != nil {
			os.Mkdir("static/uploads/", os.ModePerm)
		}

		// Create the full path for saving the file
		filePath := filepath.Join("static/uploads", filename)

		// Save the file to the server
		dst, err := os.Create(filePath)
		if err != nil {
			slog.Error(fmt.Errorf("failed to create file: %w", err).Error())
			formErrors.AppError = "Error create file"
			return rendering(w, r, profile.ProfileForm(
				visitor,
				profileToEdit,
				formErrors,
				roles,
				ressorts,
				route,
				referer,
			))
		}
		defer dst.Close()
		if _, err = io.Copy(dst, file); err != nil {
			slog.Error(fmt.Errorf("failed to write to file: %w", err).Error())
			formErrors.AppError = "Error writing to file"
			return rendering(w, r, profile.ProfileForm(
				visitor,
				profileToEdit,
				formErrors,
				roles,
				ressorts,
				route,
				referer,
			))
		}

		//Delete current image from the initial fetch of the user
		if profileToEdit.Avatar != "" {
			oldAvatarPath := filepath.Join("static/uploads", profileToEdit.Avatar)

			//Check if the oldPath is not the same as the new path
			if oldAvatarPath != filePath {
				if err := os.Remove(oldAvatarPath); err != nil {
					return fmt.Errorf("failed to delete old file: %w", err)
				}
			}
		}

		// Update the user's avatar in the visitor struct, it will be passed with the rest
		profileToEdit.Avatar = filename
	}

	// delete avatar if toggle is there
	if form.BoxValue("delete_avatar") {
		AvatarPath := filepath.Join("static/uploads", profileToEdit.Avatar)
		if err := os.Remove(AvatarPath); err != nil {
			return fmt.Errorf("failed to delete old file: %w", err)
		}
		profileToEdit.Avatar = ""
	}

	// write errors to object
	formErrors.FirstName = form.Errors.Get("first_name")
	formErrors.FamilyName = form.Errors.Get("family_name")
	formErrors.Birthdate = form.Errors.Get("birthdate")
	formErrors.Phone = form.Errors.Get("phone")
	formErrors.Email = form.Errors.Get("email")
	formErrors.Address = form.Errors.Get("address")
	formErrors.Postcode = form.Errors.Get("postcode")
	formErrors.City = form.Errors.Get("city")
	formErrors.Password = form.Errors.Get("password")
	formErrors.PasswordCheck = form.Errors.Get("password_check")
	formErrors.Avatar = form.Errors.Get("avatar")

	// what to do if form not valid
	if !form.Valid() {
		// send the data to the template
		return rendering(w, r, profile.ProfileForm(
			visitor,
			profileToEdit,
			formErrors,
			roles,
			ressorts,
			route,
			referer,
		))
	}

	// if checks passed get new password if the field is not empty
	pw := r.Form.Get("password")
	if len(pw) > 7 {
		hashedPassword, err := bcrypt.GenerateFromPassword(
			[]byte(pw),
			bcrypt.DefaultCost,
		)
		if err != nil {
			formErrors.AppError = "Error creating password hash"
			slog.Error(fmt.Errorf("failed to create password hash: %w", err).Error())
			// send the data to the template
			return rendering(w, r, profile.ProfileForm(
				visitor,
				profileToEdit,
				formErrors,
				roles,
				ressorts,
				route,
				referer,
			))
		}
		profileToEdit.Password = string(hashedPassword)
	}

	// what to do if form valid
	err = m.DB.UpdateVisitor(profileToEdit)
	if err != nil {
		formErrors.AppError = err.Error()
		slog.Error(fmt.Errorf("failed to update profile: %w", err).Error())
		// send the data to the template
		return rendering(w, r, profile.ProfileForm(
			visitor,
			profileToEdit,
			formErrors,
			roles,
			ressorts,
			route,
			referer,
		))
	}

	// get where user came from and set redirect path to this place
	redirectPath := "/admin"
	if referer != "" {
		redirectPath = "/admin/" + referer
	}

	// if it is the own profile, update the session cookie
	if route == "profile" {
		m.App.Session.Put(r.Context(), "visitor", profileToEdit)
	}

	// log success
	logger.Success(r, m.App.Session,
		fmt.Sprintf("profile for %s %s successfully updated", profileToEdit.FirstName, profileToEdit.FamilyName))

	// on success redirect
	return hxRedirect(w, r, redirectPath)
}
