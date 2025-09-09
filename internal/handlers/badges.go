package handlers

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/faebut/prozessor_gostack/internal/cardreader"
	"github.com/faebut/prozessor_gostack/internal/forms"
	"github.com/faebut/prozessor_gostack/internal/logger"
	"github.com/faebut/prozessor_gostack/internal/models"
	"github.com/faebut/prozessor_gostack/view/badges"
	"github.com/go-chi/chi/v5"
)

// HandleBadgesIndex is the badge page handler
func (m *Repository) HandleBadgesIndex(w http.ResponseWriter, r *http.Request) error {
	// get visitor
	visitor, ok := m.App.Session.Get(r.Context(), "visitor").(models.Visitor)
	if !ok {
		slog.Error("could not get visitor from session", "path", r.URL.Path)
		m.App.Session.Put(r.Context(), "error", "Can't get visitor from session")

		return hxRedirect(w, r, "/admin/logout")
	}

	return rendering(w, r, badges.Badges(visitor))
}

// HandleBadgesList gets and returns the unfiltered badges list
func (m *Repository) HandleBadgesList(w http.ResponseWriter, r *http.Request) error {
	// set initial params
	// need empty filter to get badges
	filter := models.NewBadgeFilter("")

	badgesList, _, err := m.DB.AllBadges(filter)
	if err != nil {
		// m.App.Session.Put(r.Context(), "error", "Can't get badges from database")
		return fmt.Errorf("failed to get badges: %w", err)
	}

	return rendering(w, r, badges.BadgesList(badgesList))
}

// BadgesList gets and returns the filtered badges list
func (m *Repository) HandleBadgesListFilter(w http.ResponseWriter, r *http.Request) error {
	// set initial params
	// need empty filter to get badges
	filterString := r.Form.Get("filter_string")
	filter := models.NewBadgeFilter(filterString)
	filter.InUse = r.Form.Get("inuse")
	filter.Type = r.Form.Get("type")

	badgesList, _, err := m.DB.AllBadges(filter)
	if err != nil {
		// m.App.Session.Put(r.Context(), "error", "Can't get badges from database")
		return fmt.Errorf("failed to get badges: %w", err)
	}

	return rendering(w, r, badges.BadgesList(badgesList))
}

// HandleBadgeEditIndex is the handler for the Badge edit form
func (m *Repository) HandleBadgeEditIndex(w http.ResponseWriter, r *http.Request) error {

	visitor, ok := m.App.Session.Get(r.Context(), "visitor").(models.Visitor)
	if !ok {
		slog.Error("could not get visitor from session", "path", r.URL.Path)
		m.App.Session.Put(r.Context(), "error", "Can't get visitor from session")

		return hxRedirect(w, r, "/admin/logout")
	}

	bID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		return fmt.Errorf("failed to convert string to int: %w", err)
	}

	badge, err := m.DB.BadgeByID(bID)
	if err != nil {
		return fmt.Errorf("failed to get badge: %w", err)
	}

	return rendering(w, r, badges.BadgeEdit(badge, badges.BadgeEditErrors{}, visitor, badges.BadgeEditFormParams{}))
}

// HandleBadgeReset resets the connection of a badge to internal or external visitors
func (m *Repository) HandleBadgeReset(w http.ResponseWriter, r *http.Request) error {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		return fmt.Errorf("failed to convert string to int: %w", err)
	}

	err = m.DB.ResetBadge(id)
	if err != nil {
		return fmt.Errorf("failed to reset badge: %w", err)
	}

	// log successful badge edit
	logger.Success(r, m.App.Session,
		fmt.Sprintf("badge with id %d successfully resetted", id))

	return hxRedirect(w, r, "/admin/badges")
}

// HandleBadgeUpdate updates the information and assignement of a badge
func (m *Repository) HandleBadgeUpdate(w http.ResponseWriter, r *http.Request) error {
	visitor, ok := m.App.Session.Get(r.Context(), "visitor").(models.Visitor)
	if !ok {
		slog.Error("could not get visitor from session", "path", r.URL.Path)
		m.App.Session.Put(r.Context(), "error", "Can't get visitor from session")

		return hxRedirect(w, r, "/admin/logout")
	}

	var badge models.Badge
	var formErrors badges.BadgeEditErrors
	// get id for badge
	bid, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		return fmt.Errorf("failed to convert string to int: %w", err)
	}
	badge.ID = bid

	form := forms.New(r.Form)
	// get editable fields
	badge.Number, err = strconv.Atoi(r.Form.Get("number"))
	if err != nil {
		return fmt.Errorf("failed to convert string to int: %w", err)
	}
	badge.UID = r.Form.Get("uid")

	var params badges.BadgeEditFormParams
	inUse := form.BoxValue("inuse")
	extern := form.BoxValue("extern")
	params.InUse = inUse
	params.Extern = extern

	if inUse {
		// set values from form
		if extern {
			// handle form fields only when extern is set
			form.Required("ext_firstname", "ext_familyname", "ext_purpose")
			form.OneOf("ext_email", "ext_phone")
			form.IsEmail("ext_email")
			form.IsPhone("ext_phone")

			badge.ExternFirstName = r.Form.Get("ext_firstname")
			badge.ExternFamilyName = r.Form.Get("ext_familyname")
			badge.ExternCity = r.Form.Get("ext_city")
			badge.ExternPhone = r.Form.Get("ext_phone")
			badge.ExternEmail = r.Form.Get("ext_email")
			badge.ExternPurpose = r.Form.Get("ext_purpose")
		} else {
			form.Required("member")
			badge.VisitorID = r.Form.Get("member")
		}

		// what to do if form not valid
		if !form.Valid() {
			formErrors.ExternFirstName = form.Errors.Get("ext_firstname")
			formErrors.ExternFamilyName = form.Errors.Get("ext_familyname")
			formErrors.ExternPurpose = form.Errors.Get("ext_purpose")
			formErrors.ExternPhone = form.Errors.Get("ext_phone")
			formErrors.ExternEmail = form.Errors.Get("ext_email")
			if len(form.Errors.Get("member")) > 0 {
				formErrors.Member = "Bitte ein Mitglied auswählen"
			}

			// send the data to the template
			return rendering(w, r, badges.BadgeEdit(badge, formErrors, visitor, params))
		}
		// write values to database
		err := m.DB.UpdateBadge(badge)

		if err != nil {
			return fmt.Errorf("failed to update badge with id %d: %w", badge.ID, err)
		}

		// log success
		logger.Success(r, m.App.Session,
			fmt.Sprintf("badge with ID %d successfully updated", badge.ID))
	} else {

		// TODO: is this still valid? does not check for a specific permission but for the role ID, does it really matter? If it is not in use, it should be resetted anyways

		// what to do if not in use. Do Update with new values, because of what a superuser can do or reset if not superuser
		if visitor.RoleID >= 8 {
			// write values to database
			err := m.DB.UpdateBadge(badge)

			if err != nil {
				return fmt.Errorf("failed to update badge with id %d: %w", badge.ID, err)
			}

			// log success
			logger.Success(r, m.App.Session,
				fmt.Sprintf("badge with ID %d successfully updated", badge.ID))
		} else {
			err = m.DB.ResetBadge(badge.ID)
			if err != nil {
				return fmt.Errorf("failed to reset badge with id %d: %w", badge.ID, err)
			}

			// log success
			logger.Success(r, m.App.Session,
				fmt.Sprintf("badge with ID %d successfully resetted", badge.ID))
		}
	}

	// on success redirect
	return hxRedirect(w, r, "/admin/badges")
}

// HandleBadgeCreateIndex is the handler for the Badge creation form
func (m *Repository) HandleBadgeCreateIndex(w http.ResponseWriter, r *http.Request) error {
	// return an empty form
	return rendering(w, r, badges.BadgeCreate(models.Badge{}, badges.BadgeCreateErrors{}))
}

// HandleBadgeCreate is the handler for posting the badge creation form
func (m *Repository) HandleBadgeCreate(w http.ResponseWriter, r *http.Request) error {
	// create empty badge
	var badge models.Badge
	var formErrors badges.BadgeCreateErrors

	form := forms.New(r.Form)
	form.Required("number", "uid")
	form.IsHigher("number", 0)

	badge.UID = r.Form.Get("uid")
	number, err := strconv.Atoi(r.Form.Get("number"))
	if err != nil {
		return fmt.Errorf("failed to convert string to int: %w", err)
	}
	badge.Number = number

	numValid, uidValid, err := m.DB.UniqueBadge(badge.Number, badge.UID)
	if err != nil {
		slog.Error("error getting uniqueness of id and num", err)
	}

	if !numValid {
		form.Errors.Add("number", "Badge mit dieser Nummer exisitert schon")
	}
	if !uidValid {
		form.Errors.Add("uid", "Badge mit dieser UID exisitert schon")
	}

	if !form.Valid() {
		formErrors.Number = form.Errors.Get("number")
		formErrors.UID = form.Errors.Get("uid")
		return rendering(w, r, badges.BadgeCreate(badge, formErrors))
	}

	badgeID, err := m.DB.CreateBadge(badge)
	if err != nil {
		return fmt.Errorf("failed to create badge: %w", err)
	}

	// log success
	logger.Success(r, m.App.Session,
		fmt.Sprintf("badge with ID %d successfully created", badgeID))

	// redirect on success
	return hxRedirect(w, r, "/admin/badges")
}

// HandleBadgeDelete is the handler for deleting a badge
func (m *Repository) HandleBadgeDelete(w http.ResponseWriter, r *http.Request) error {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		return fmt.Errorf("failed to convert string to int: %w", err)
	}

	err = m.DB.DeleteBadge(id)
	if err != nil {
		return fmt.Errorf("failed to delete badge: %w", err)
	}

	// log success
	logger.Success(r, m.App.Session,
		fmt.Sprintf("badge with ID %d successfully deleted", id))

	// redirect on success
	return hxRedirect(w, r, "/admin/badges")
}

// HandleBadgeRead handles a read request by badge reader and returns the uid
func (m *Repository) HandleBadgeRead(w http.ResponseWriter, r *http.Request) error {
	// set timeout for card reading
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// start reader
	myReader, err := cardreader.NewCardReader()
	if err != nil {
		slog.Error(fmt.Errorf("failed to start card reader: %w", err).Error())
		return rendering(w, r, badges.BadgeReturnUID("", "noreader"))
	}

	// read card
	res, err := myReader.ReadCard(ctx)
	if err != nil {
		if err.Error() == "Deadline Exceeded" {
			return rendering(w, r, badges.BadgeReturnUID("", "deadline"))
		} else {
			return fmt.Errorf("failed to handle card reader: %w", err)
		}
	}

	return rendering(w, r, badges.BadgeReturnUID(res.UID, ""))
}
