package handlers

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/faebut/prozessor_gostack/internal/forms"
	"github.com/faebut/prozessor_gostack/internal/models"
	"github.com/faebut/prozessor_gostack/view/machines"
	"github.com/go-chi/chi/v5"
)

// HandleMachinesIndex handles the Machines Page
func (m *Repository) HandleMachinesIndex(w http.ResponseWriter, r *http.Request) error {
	// send the data to the template
	visitor, ok := m.App.Session.Get(r.Context(), "visitor").(models.Visitor)
	if !ok {
		log.Println("could not get visitor from Session")
		m.App.Session.Put(r.Context(), "error", "Can't get visitor from session")

		return hxRedirect(w, r, "/admin/logout")
	}

	ressorts, err := m.DB.AllRessorts()
	if err != nil {
		log.Println("could not get ressorts from DB: ", err)
		m.DB.CreateLog(
			models.Log{
				Typ:     "error",
				Module:  "machines",
				Message: fmt.Sprintf("Can't get ressorts from database: %s", err),
			},
		)

		m.App.Session.Put(r.Context(), "error", "Can't get ressorts from database")
	}

	machineList, _, err := m.DB.AllMachines(models.MachineFilter{
		MachineName: "",
		RessortID:   0,
	})

	log.Println("machines:", machineList)

	if err != nil {
		log.Println(err)
		return nil
	}

	// render the template
	return rendering(w, r, machines.Machines(visitor, machineList, ressorts))
}

// HandleMachinesCreateIndex handles the Machine creation Page
func (m *Repository) HandleMachinesCreateIndex(w http.ResponseWriter, r *http.Request) error {
	// send the data to the template
	ressorts, err := m.DB.AllRessorts()
	if err != nil {
		log.Println("could not get ressorts from DB: ", err)
		m.DB.CreateLog(
			models.Log{
				Typ:     "error",
				Module:  "machines",
				Message: fmt.Sprintf("Can't get ressorts from database: %s", err),
			},
		)

		m.App.Session.Put(r.Context(), "error", "Can't get badges from database")
	}

	// render the template
	return rendering(w, r, machines.MachinesForm(models.Machine{}, ressorts, machines.MachineErrors{}))
}

// HandleMachinesCreate handles the Machine creation post
func (m *Repository) HandleMachinesCreate(w http.ResponseWriter, r *http.Request) error {
	err := r.ParseForm()
	if err != nil {
		log.Println(err)
		// TODO better error handling
		return nil
	}

	var machine models.Machine

	machine.MachineName = r.Form.Get("machine_name")
	ressortID, err := strconv.Atoi(r.Form.Get("ressort"))
	if err == nil {
		machine.RessortID = ressortID
	}
	log.Println("ressort: ", ressortID)

	// validation
	form := forms.New(r.PostForm)
	form.Required("machine_name")
	if machine.RessortID < 1 {
		form.Errors.Add("ressort", "Feld ist obligatorisch")
	}

	// what to do if form not valid
	if !form.Valid() {
		ressorts, err := m.DB.AllRessorts()
		if err != nil {
			log.Println("could not get ressorts from DB: ", err)
			m.DB.CreateLog(
				models.Log{
					Typ:     "error",
					Module:  "machines",
					Message: fmt.Sprintf("Can't get ressorts from database: %s", err),
				},
			)

			m.App.Session.Put(r.Context(), "error", "Can't get badges from database")
		}
		// return form with errors and allready filled fields
		return rendering(w, r, machines.MachinesForm(
			machine,
			ressorts,
			machines.MachineErrors{
				MachineName: form.Errors.Get("machine_name"),
				Ressort:     form.Errors.Get("ressort"),
			},
		))
	}

	// post machine to DB
	uid, err := m.DB.CreateMachine(machine)
	if err != nil {
		log.Println(err)
		return nil
	}

	log.Println("new Machine: ", uid)

	// if everything okay, go back to machines
	return hxRedirect(w, r, "/admin/machines")
}

// HandleMachinesEditIndex handles the Machine edit Page
func (m *Repository) HandleMachinesEditIndex(w http.ResponseWriter, r *http.Request) error {
	// send the data to the template
	ressorts, err := m.DB.AllRessorts()
	if err != nil {
		log.Println("could not get ressorts from DB: ", err)
		m.DB.CreateLog(
			models.Log{
				Typ:     "error",
				Module:  "machines",
				Message: fmt.Sprintf("Can't get ressorts from database: %s", err),
			},
		)
	}

	id := chi.URLParam(r, "id")

	machine, err := m.DB.MachineByID(id)
	if err != nil {
		log.Println("could not get machine from DB: ", err)
		m.DB.CreateLog(
			models.Log{
				Typ:     "error",
				Module:  "machines",
				Message: fmt.Sprintf("Can't get machine from database: %s", err),
			},
		)
	}

	// render the template
	return rendering(w, r, machines.MachinesForm(machine, ressorts, machines.MachineErrors{}))
}

// HandleMachinesEdit handles the Machine edit post
func (m *Repository) HandleMachinesEdit(w http.ResponseWriter, r *http.Request) error {
	var machine models.Machine
	// get id
	id := chi.URLParam(r, "id")
	machine.ID = id

	err := r.ParseForm()
	if err != nil {
		log.Println(err)
		// TODO better error handling
		return nil
	}

	machine.MachineName = r.Form.Get("machine_name")
	ressortID, err := strconv.Atoi(r.Form.Get("ressort"))
	if err == nil {
		machine.RessortID = ressortID
	}
	log.Println("ressort: ", ressortID)

	// validation
	form := forms.New(r.PostForm)
	form.Required("machine_name")
	if machine.RessortID < 1 {
		form.Errors.Add("ressort", "Feld ist obligatorisch")
	}

	// what to do if form not valid
	if !form.Valid() {
		ressorts, err := m.DB.AllRessorts()
		if err != nil {
			log.Println("could not get ressorts from DB: ", err)
			m.DB.CreateLog(
				models.Log{
					Typ:     "error",
					Module:  "machines",
					Message: fmt.Sprintf("Can't get ressorts from database: %s", err),
				},
			)

			m.App.Session.Put(r.Context(), "error", "Can't get machine from database")
		}
		// return form with errors and allready filled fields
		return rendering(w, r, machines.MachinesForm(
			machine,
			ressorts,
			machines.MachineErrors{
				MachineName: form.Errors.Get("machine_name"),
				Ressort:     form.Errors.Get("ressort"),
			},
		))
	}

	// post machine to DB
	err = m.DB.UpdateMachine(machine)
	if err != nil {
		log.Println(err)
		return nil
	}

	// if everything okay, go back to machines
	return hxRedirect(w, r, "/admin/machines")
}
