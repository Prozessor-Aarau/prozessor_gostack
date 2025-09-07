package handlers

import (
	"log"
	"net/http"
	"strconv"

	"github.com/faebut/prozessor_gostack/internal/forms"
	"github.com/faebut/prozessor_gostack/internal/models"
	"github.com/faebut/prozessor_gostack/internal/qrcodes"
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
	ressorts = ressorts[:len(ressorts)-1]

	if err != nil {
		// m.App.Session.Put(r.Context(), "error", "Can't get ressorts from database")
		return err
	}

	// render the template
	return rendering(w, r, machines.Machines(visitor, ressorts))
}

// HandleMachinesList handles the machines list
func (m *Repository) HandleMachinesList(w http.ResponseWriter, r *http.Request) error {

	machineList, _, err := m.DB.AllMachines(models.MachineFilter{
		MachineName: "",
		RessortID:   0,
	})

	if err != nil {
		log.Println(err)
		return nil
	}

	// render the template
	return rendering(w, r, machines.MachinesList(machineList))
}

// HandleMachinesListFilter handles the machines list search filters
func (m *Repository) HandleMachinesListFilter(w http.ResponseWriter, r *http.Request) error {

	// initial values
	var mfilter models.MachineFilter
	mfilter.MachineName = ""
	mfilter.RessortID = 0

	// get data from Form and set
	f := r.Form.Get("filter_string")
	if len(f) > 0 {
		mfilter.MachineName = f
	}
	rID, err := strconv.Atoi(r.Form.Get("ressort"))
	if err != nil {
		log.Println("no ressort set")
	}
	if rID > 0 {
		mfilter.RessortID = rID
	}

	machineList, _, err := m.DB.AllMachines(mfilter)

	if err != nil {
		log.Println(err)
		return nil
	}

	// render the template
	return rendering(w, r, machines.MachinesList(machineList))
}

// HandleMachinesCreateIndex handles the Machine creation Page
func (m *Repository) HandleMachinesCreateIndex(w http.ResponseWriter, r *http.Request) error {
	// send the data to the template
	ressorts, err := m.DB.AllRessorts()
	ressorts = ressorts[:len(ressorts)-1]

	if err != nil {
		// m.App.Session.Put(r.Context(), "error", "Can't get badges from database")
		return err
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

	// validation
	form := forms.New(r.PostForm)
	form.Required("machine_name")
	if machine.RessortID < 1 {
		form.Errors.Add("ressort", "Feld ist obligatorisch")
	}

	// what to do if form not valid
	if !form.Valid() {
		ressorts, err := m.DB.AllRessorts()
		ressorts = ressorts[:len(ressorts)-1]
		if err != nil {
			// m.App.Session.Put(r.Context(), "error", "Can't get badges from database")
			return err
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
	ressorts = ressorts[:len(ressorts)-1]

	if err != nil {
		return err
	}

	id := chi.URLParam(r, "id")

	machine, err := m.DB.MachineByID(id)
	if err != nil {
		return err
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

	// validation
	form := forms.New(r.PostForm)
	form.Required("machine_name")
	if machine.RessortID < 1 {
		form.Errors.Add("ressort", "Feld ist obligatorisch")
	}

	// what to do if form not valid
	if !form.Valid() {
		ressorts, err := m.DB.AllRessorts()
		ressorts = ressorts[:len(ressorts)-1]

		if err != nil {
			// m.App.Session.Put(r.Context(), "error", "Can't get machine from database")
			return err
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

// HandleMachinesQRCreate handles the Machine QR code generation
func (m *Repository) HandleMachinesQRCreate(w http.ResponseWriter, r *http.Request) error {
	// get id
	id := chi.URLParam(r, "id")

	machine, err := m.DB.MachineByID(id)
	if err != nil {
		return err
	}

	err = qrcodes.GenerateQRCode(machine)

	if err != nil {
		return err
	}

	// write changes to database
	machine.QR = "QR_" + machine.ID + ".png"
	err = m.DB.SetMachineQR(machine)
	if err != nil {
		log.Println("could not write QR Code to DB: ", err)
		return nil
	}

	// if everything okay, go back to machines
	return rendering(w, r, machines.QRcode(machine))
}

// HandleMachinesQRDelete handles the Machine QR code deletion
func (m *Repository) HandleMachinesQRDelete(w http.ResponseWriter, r *http.Request) error {
	// get id
	id := chi.URLParam(r, "id")

	machine, err := m.DB.MachineByID(id)
	if err != nil {
		return err
	}

	err = qrcodes.RemoveQRCode(machine)

	if err != nil {
		return err
	}

	// write changes to database
	err = m.DB.DeleteMachineQR(machine)
	if err != nil {
		return err
	}

	machine.QR = ""

	// if everything okay, go back to machines
	return rendering(w, r, machines.QRcode(machine))
}

// HandleMachinesQRPrint handles the Machine QR code deletion
func (m *Repository) HandleMachinesQRPrint(w http.ResponseWriter, r *http.Request) error {
	// get id
	id := chi.URLParam(r, "id")

	machine, err := m.DB.MachineByID(id)
	if err != nil {
		return err
	}

	// print the QR-Code
	err = qrcodes.PrintQRCode(machine)
	if err != nil {
		// TODO: add logging to db here and do something else in the frontend
		return rendering(w, r, machines.ErrorButton("Druckerfehler!"))
	}

	return rendering(w, r, machines.SucessButton(machine))
}

// HandleMachinesQRDelete handles the Machine QR code deletion
func (m *Repository) HandleMachinesDelete(w http.ResponseWriter, r *http.Request) error {
	// get id
	id := chi.URLParam(r, "id")

	machine, err := m.DB.MachineByID(id)
	if err != nil {
		log.Println("could not get machine from DB: ", err)
		return nil
	}

	if len(machine.QR) > 0 {
		err = qrcodes.RemoveQRCode(machine)

		if err != nil {
			log.Println("could not Delete QR Code: ", err)
			return nil
		}
	}

	err = m.DB.DeleteMachineById(machine.ID)
	if err != nil {
		log.Println("could not delete machine in DB: ", err)
		return nil
	}

	// if everything okay, go back to machines
	return hxRedirect(w, r, "/admin/machines")
}

// HandleMachinesQRDelete handles the Machine QR code deletion
func (m *Repository) HandleMachineUpdateReportStatus(w http.ResponseWriter, r *http.Request) error {
	// get url params
	machineID := chi.URLParam(r, "id")

	reportID, err := strconv.Atoi(chi.URLParam(r, "reportid"))
	if err != nil {
		return err
	}

	solved, err := strconv.ParseBool(chi.URLParam(r, "solved"))
	if err != nil {
		return err
	}

	// set new solved-status in DB
	err = m.DB.UpdateReportStatus(solved, reportID)
	if err != nil {
		return err
	}

	// get updated list of reports
	reports, err := m.DB.ReportsByMachineID(machineID)
	if err != nil {
		return err
	}

	// render new list to frontend
	return rendering(w, r, machines.ReportList(reports))
}

// HandleMachinesQRDelete handles the Machine QR code deletion
func (m *Repository) HandleMachineDeleteReport(w http.ResponseWriter, r *http.Request) error {
	// get url params
	machineID := chi.URLParam(r, "id")

	reportID, err := strconv.Atoi(chi.URLParam(r, "reportid"))
	if err != nil {
		return err
	}

	// set new solved-status in DB
	err = m.DB.DeleteReport(reportID)
	if err != nil {
		return err
	}

	// get updated list of reports
	reports, err := m.DB.ReportsByMachineID(machineID)
	if err != nil {
		return err
	}

	// render new list to frontend
	return rendering(w, r, machines.ReportList(reports))
}
