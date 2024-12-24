package handlers

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/faebut/prozessor_gostack/internal/cardreader"
	"github.com/faebut/prozessor_gostack/internal/forms"
	"github.com/faebut/prozessor_gostack/internal/models"
	"github.com/faebut/prozessor_gostack/internal/render"
	"github.com/faebut/prozessor_gostack/view/badges"
	"github.com/go-chi/chi/v5"
)

// HandleBadgesIndex is the badge page handler
func (m *Repository) HandleBadgesIndex(w http.ResponseWriter, r *http.Request) error {
	// send the data to the template
	visitor, ok := m.App.Session.Get(r.Context(), "visitor").(models.Visitor)
	if !ok {
		log.Println("could not get visitor from Session")
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
		log.Println("could not get badges from DB: ", err)
		m.DB.CreateLog(
			models.Log{
				Typ:     "error",
				Module:  "badges",
				Message: fmt.Sprintf("Can't get badges from database: %s", err),
			},
		)

		m.App.Session.Put(r.Context(), "error", "Can't get badges from database")
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
		log.Println("could not get badges from DB: ", err)
		m.DB.CreateLog(
			models.Log{
				Typ:     "error",
				Module:  "badges",
				Message: fmt.Sprintf("Can't get badges from database: %s", err),
			},
		)
		m.App.Session.Put(r.Context(), "error", "Can't get badges from database")
	}

	return rendering(w, r, badges.BadgesList(badgesList))
}

// Badge is the handler for a single Badge
func (m *Repository) Badge(w http.ResponseWriter, r *http.Request) {
	// set initial values
	super := false

	user, ok := m.App.Session.Get(r.Context(), "visitor").(models.Visitor)
	if !ok {
		log.Println("could not get visitor from Session")
		m.App.Session.Put(r.Context(), "error", "Can't get visitor from session")

		m.DB.CreateLog(
			models.Log{
				Typ:     "error",
				Module:  "badges",
				Message: "Can't get visitor from session",
			},
		)

		w.Header().Set("HX-Redirect", "/admin/logout")
		w.WriteHeader(http.StatusSeeOther)
		return
	}

	if user.RoleID == 4 {
		super = true
	}

	bID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		m.DB.CreateLog(
			models.Log{
				Typ:     "error",
				Module:  "badges",
				Message: fmt.Sprintf("Error converting badge ID to int: %s", err),
			},
		)
		log.Println("Error converting badge id to int", err)
	}

	badge, err := m.DB.BadgeByID(bID)
	if err != nil {
		m.DB.CreateLog(
			models.Log{
				Typ:     "error",
				Module:  "badges",
				Message: fmt.Sprintf("Error getting badge from database: %s", err),
			},
		)
		log.Println("Error getting badge from database:", err)
	}

	data := make(map[string]interface{})
	data["badge"] = badge
	data["super"] = super

	render.Template(w, r, "part_badge_edit.html", &models.TemplateData{
		Data: data,
		Form: forms.New(nil),
	})
}

// BadgeReset resets the connection of a badge to internal or external visitors
func (m *Repository) BadgeReset(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		log.Println("Error converting string to int:", err)
	}

	err = m.DB.ResetBadge(id)
	if err != nil {
		m.DB.CreateLog(
			models.Log{
				Typ:     "error",
				Module:  "badges",
				Message: fmt.Sprintf("Error resetting badge: %s", err),
			},
		)
		log.Println("Error resetting badge:", err)
	}

	// on success redirect
	m.DB.CreateLog(
		models.Log{
			Typ:     "info",
			Module:  "badges",
			Message: fmt.Sprintf("Badge with ID %d has been resetted", id),
		},
	)
	w.Header().Set("HX-Redirect", "/admin/l3/badges")
	w.WriteHeader(http.StatusSeeOther)
}

// BadgeUpdate updates the information and assignement of a badge
func (m *Repository) BadgeUpdate(w http.ResponseWriter, r *http.Request) {
	// set initial values
	super, err := strconv.ParseBool(r.Form.Get("super"))
	if err != nil {
		m.DB.CreateLog(
			models.Log{
				Typ:     "error",
				Module:  "badges",
				Message: fmt.Sprintf("Error converting string to bool: %s", err),
			},
		)
		log.Println("Error converting string to bool:", err)
	}

	var badge models.Badge
	// get id for badge
	bid, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		m.DB.CreateLog(
			models.Log{
				Typ:     "error",
				Module:  "badges",
				Message: fmt.Sprintf("Error converting string to int: %s", err),
			},
		)
		log.Println("Error converting string to int:", err)
	}
	badge.ID = bid

	form := forms.New(r.Form)
	// get editable fields
	badge.Number, err = strconv.Atoi(r.Form.Get("number"))
	if err != nil {
		m.DB.CreateLog(
			models.Log{
				Typ:     "error",
				Module:  "badges",
				Message: fmt.Sprintf("Error converting string to int: %s", err),
			},
		)
		log.Println("Error converting string to int:", err)
	}
	badge.UID = r.Form.Get("uid")

	inUse := form.BoxValue("inuse")
	extern := form.BoxValue("extern")

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
			badge.VisitorID = r.Form.Get("member")
		}

		// what to do if form not valid
		if !form.Valid() {
			// prepare data to send it to the templates
			data := make(map[string]interface{})
			data["super"] = super
			data["badge"] = badge

			// send the data to the template
			render.Template(w, r, "part_badge_edit.html", &models.TemplateData{
				Form: form,
				Data: data,
			})
			return
		}
		// write values to database
		err := m.DB.UpdateBadge(badge)

		// TODO: do better error handling
		if err != nil {
			m.DB.CreateLog(
				models.Log{
					Typ:     "error",
					Module:  "badges",
					Message: fmt.Sprintf("Error updating badge: %s", err),
				},
			)
			log.Println("Error updating badge:", err)
		}
	} else {
		if super {
			// write values to database
			err := m.DB.UpdateBadge(badge)

			// TODO: do better error handling
			if err != nil {
				m.DB.CreateLog(
					models.Log{
						Typ:     "error",
						Module:  "badges",
						Message: fmt.Sprintf("Error updating badge: %s", err),
					},
				)
				log.Println("Error updating badge:", err)
			}
		} else {
			err = m.DB.ResetBadge(badge.ID)
			if err != nil {
				m.DB.CreateLog(
					models.Log{
						Typ:     "error",
						Module:  "badges",
						Message: fmt.Sprintf("Error resetting badge: %s", err),
					},
				)
				log.Println("Error resetting badge:", err)
			}
		}
	}

	// on success redirect
	m.DB.CreateLog(
		models.Log{
			Typ:     "info",
			Module:  "badges",
			Message: fmt.Sprintf("Badge with Number %d has been updated", badge.Number),
		},
	)
	w.Header().Set("HX-Redirect", "/admin/l3/badges")
	w.WriteHeader(http.StatusSeeOther)
}

// Badge is the handler for a single Badge
func (m *Repository) BadgeCreate(w http.ResponseWriter, r *http.Request) {
	// create empty badge
	badge := models.Badge{}

	data := make(map[string]interface{})
	data["badge"] = badge

	render.Template(w, r, "part_badge_create.html", &models.TemplateData{
		Data: data,
		Form: forms.New(nil),
	})
}

// Badge is the handler for a single Badge
func (m *Repository) PostBadgeCreate(w http.ResponseWriter, r *http.Request) {
	// create empty badge
	var badge models.Badge

	form := forms.New(r.Form)
	form.Required("number", "uid")
	form.IsHigher("number", 0)

	badge.UID = r.Form.Get("uid")
	number, err := strconv.Atoi(r.Form.Get("number"))
	if err != nil {
		m.DB.CreateLog(
			models.Log{
				Typ:     "error",
				Module:  "badges",
				Message: fmt.Sprintf("Error converting string to int: %s", err),
			},
		)
		log.Println("Error converting string to int:", err)
	}
	badge.Number = number

	numValid, uidValid, err := m.DB.UniqueBadge(badge.Number, badge.UID)

	if !numValid {
		form.Errors.Add("number", "Badge mit dieser Nummer exisitert schon")
	}
	if !uidValid {
		form.Errors.Add("uid", "Badge mit dieser UID exisitert schon")
	}

	if !form.Valid() {
		data := make(map[string]interface{})
		data["badge"] = badge

		render.Template(w, r, "part_badge_create.html", &models.TemplateData{
			Data: data,
			Form: form,
		})
		return
	}

	id, err := m.DB.CreateBadge(badge)
	if err != nil {
		m.DB.CreateLog(
			models.Log{
				Typ:     "error",
				Module:  "badges",
				Message: fmt.Sprintf("Error converting string to int: %s", err),
			},
		)
		log.Println("Error converting string to int:", err)
		return
	} else {
		m.DB.CreateLog(
			models.Log{
				Typ:     "info",
				Module:  "badges",
				Message: fmt.Sprintf("Created new badge with id: %d", id),
			},
		)
	}

	// redirect on success
	w.Header().Set("HX-Redirect", "/admin/l3/badges")
	w.WriteHeader(http.StatusSeeOther)
}

// BadgeDelete is the handler for deleting a badge
func (m *Repository) BadgeDelete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		m.DB.CreateLog(
			models.Log{
				Typ:     "error",
				Module:  "badges",
				Message: fmt.Sprintf("Error converting string to int: %s", err),
			},
		)
		log.Println("Error converting string to int:", err)
		return
	}

	err = m.DB.DeleteBadge(id)
	if err != nil {
		m.DB.CreateLog(
			models.Log{
				Typ:     "error",
				Module:  "badges",
				Message: fmt.Sprintf("Error deleting badge: %s", err),
			},
		)
		log.Println("Error deleting badge:", err)
		return
	} else {
		m.DB.CreateLog(
			models.Log{
				Typ:     "info",
				Module:  "badges",
				Message: fmt.Sprintf("Badge with id %d has been deleted", id),
			},
		)
	}
	// redirect on success
	w.Header().Set("HX-Redirect", "/admin/l3/badges")
	w.WriteHeader(http.StatusSeeOther)
}

func (m *Repository) BadgeRead(w http.ResponseWriter, r *http.Request) {
	// set timeout for card reading
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	myReader, err := cardreader.NewCardReader()
	if err != nil {
		log.Println(err)
		fmt.Fprint(w, `<span class="text-error">Kein Lesegerät verbunden!</span>`)
		return
	}

	res, err := myReader.ReadCard(ctx)
	if err != nil {
		if err.Error() == "Deadline Exceeded" {
			fmt.Fprint(
				w,
				`<span class="text-error">Dies hat zu lange gedauert, versuchs nochmals!</span>`,
			)
		} else {
			log.Println("error handling for other errors", err)
		}
	} else {
		// uid to template data and return in new file
		data := make(map[string]interface{})
		data["uid"] = res.UID

		w.Header().Set("HX-Retarget", "#badge")
		render.Template(w, r, "comp_return_uid.html", &models.TemplateData{
			Data: data,
		})
	}
}
