package handlers

import (
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/faebut/prozessor_gostack/internal/forms"
	"github.com/faebut/prozessor_gostack/internal/logger"
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
		slog.Error("could not get visitor from session", "path", r.URL.Path)
		m.App.Session.Put(r.Context(), "error", "Can't get visitor from session")

		return hxRedirect(w, r, "/admin/logout")
	}

	ressorts, err := m.DB.AllRessorts()
	if err != nil {
		// m.App.Session.Put(r.Context(), "error", "Can't get ressorts from database")
		return fmt.Errorf("failed to get ressorts: %w", err)
	}

	// remove Betriebsgruppe alt from list
	ressorts = ressorts[:len(ressorts)-1]

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
		return fmt.Errorf("failed to get machines: %w", err)
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
	// if there is an error, ignore it
	rID, _ := strconv.Atoi(r.Form.Get("ressort"))
	if rID > 0 {
		mfilter.RessortID = rID
	}

	machineList, _, err := m.DB.AllMachines(mfilter)
	if err != nil {
		return fmt.Errorf("failed to get machines: %w", err)
	}

	// render the template
	return rendering(w, r, machines.MachinesList(machineList))
}

// HandleMachinesCreateIndex handles the Machine creation Page
func (m *Repository) HandleMachinesCreateIndex(w http.ResponseWriter, r *http.Request) error {
	// send the data to the template
	ressorts, err := m.DB.AllRessorts()
	if err != nil {
		return fmt.Errorf("failed to get ressorts: %w", err)
	}

	// remove Betriebsgruppe alt
	ressorts = ressorts[:len(ressorts)-1]

	// render the template
	return rendering(w, r, machines.MachinesForm(models.Machine{}, ressorts, machines.MachineErrors{}))
}

// HandleMachinesCreate handles the Machine creation post
func (m *Repository) HandleMachinesCreate(w http.ResponseWriter, r *http.Request) error {
	err := r.ParseForm()
	if err != nil {
		return fmt.Errorf("failed to parse form: %w", err)
	}

	var machine models.Machine

	machine.MachineName = r.Form.Get("machine_name")
	ressortID, err := strconv.Atoi(r.Form.Get("ressort"))
	if err == nil {
		machine.RessortID = ressortID
	}
	machine.Location = r.Form.Get("location")

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
			return fmt.Errorf("failed to get ressorts: %w", err)
		}
		// remove Betriebsgruppe alt
		ressorts = ressorts[:len(ressorts)-1]
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
		return fmt.Errorf("failed to create machine: %w", err)
	}

	// log success
	logger.Success(r, m.App.Session,
		fmt.Sprintf("machine %s with id %s successfully created", machine.MachineName, uid))

	// if everything okay, go back to machines
	return hxRedirect(w, r, "/admin/machines")
}

// HandleMachinesEditIndex handles the Machine edit Page
func (m *Repository) HandleMachinesEditIndex(w http.ResponseWriter, r *http.Request) error {
	// send the data to the template
	ressorts, err := m.DB.AllRessorts()
	if err != nil {
		return fmt.Errorf("failed to get ressorts: %w", err)
	}

	// remove Betriebsgruppe alt
	ressorts = ressorts[:len(ressorts)-1]

	id := chi.URLParam(r, "id")

	machine, err := m.DB.MachineByID(id)
	if err != nil {
		return fmt.Errorf("failed to get machine: %w", err)
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
		return fmt.Errorf("failed to parse form: %w", err)
	}

	machine.MachineName = r.Form.Get("machine_name")
	ressortID, err := strconv.Atoi(r.Form.Get("ressort"))
	if err == nil {
		machine.RessortID = ressortID
	}
	machine.Location = r.Form.Get("location")

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
			return fmt.Errorf("failed to get ressorts: %w", err)
		}
		// remove Betriebsgruppe alt
		ressorts = ressorts[:len(ressorts)-1]
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
		return fmt.Errorf("failed to update machine: %w", err)
	}

	// log success
	logger.Success(r, m.App.Session,
		fmt.Sprintf("machine with id %s successfully updated", machine.ID))

	// if everything okay, go back to machines
	return hxRedirect(w, r, "/admin/machines")
}

// HandleMachinesQRCreate handles the Machine QR code generation
func (m *Repository) HandleMachinesQRCreate(w http.ResponseWriter, r *http.Request) error {
	// get id
	id := chi.URLParam(r, "id")

	machine, err := m.DB.MachineByID(id)
	if err != nil {
		return fmt.Errorf("failed to get machine: %w", err)
	}

	err = qrcodes.GenerateQRCode(machine)
	if err != nil {
		return fmt.Errorf("failed to generate qr code: %w", err)
	}

	// write changes to database
	machine.QR = "QR_" + machine.ID + ".png"
	err = m.DB.SetMachineQR(machine)
	if err != nil {
		return fmt.Errorf("failed to write qr code to db: %w", err)
	}

	// log success
	logger.Success(r, m.App.Session,
		fmt.Sprintf("qr code for machine id %s successfully created", id))

	// if everything okay, go back to machines
	return rendering(w, r, machines.QRcode(machine))
}

// HandleMachinesQRDelete handles the Machine QR code deletion
func (m *Repository) HandleMachinesQRDelete(w http.ResponseWriter, r *http.Request) error {
	// get id
	id := chi.URLParam(r, "id")

	machine, err := m.DB.MachineByID(id)
	if err != nil {
		return fmt.Errorf("failed to get machine: %w", err)
	}

	err = qrcodes.RemoveQRCode(machine)
	if err != nil {
		return fmt.Errorf("failed to remove qr code: %w", err)
	}

	// write changes to database
	err = m.DB.DeleteMachineQR(machine)
	if err != nil {
		return fmt.Errorf("failed to remove qr code from db: %w", err)
	}

	machine.QR = ""

	// log success
	logger.Success(r, m.App.Session,
		fmt.Sprintf("qr code for machine id %s successfully removed", id))

	// if everything okay, go back to machines
	return rendering(w, r, machines.QRcode(machine))
}

// HandleMachinesQRPrint handles the Machine QR code deletion
func (m *Repository) HandleMachinesQRPrint(w http.ResponseWriter, r *http.Request) error {
	// get id
	id := chi.URLParam(r, "id")

	machine, err := m.DB.MachineByID(id)
	if err != nil {
		return fmt.Errorf("failed to get machine: %w", err)
	}

	// print the QR-Code
	err = qrcodes.PrintQRCode(machine)
	if err != nil {
		slog.Error(fmt.Errorf("failed to print qr code: %w", err).Error())
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
		return fmt.Errorf("failed to get machine: %w", err)
	}

	if len(machine.QR) > 0 {
		err = qrcodes.RemoveQRCode(machine)
		if err != nil {
			return fmt.Errorf("failed to delete machine qr code: %w", err)
		}
	}

	err = m.DB.DeleteMachineById(machine.ID)
	if err != nil {
		return fmt.Errorf("failed to delete machine: %w", err)
	}

	// log success
	logger.Success(r, m.App.Session,
		fmt.Sprintf("machine with id %s successfully deleted", id))

	// if everything okay, go back to machines
	return hxRedirect(w, r, "/admin/machines")
}

// HandleMachinesQRDelete handles the Machine QR code deletion
func (m *Repository) HandleMachineUpdateReportStatus(w http.ResponseWriter, r *http.Request) error {
	// get url params
	machineID := chi.URLParam(r, "id")

	reportID, err := strconv.Atoi(chi.URLParam(r, "reportid"))
	if err != nil {
		return fmt.Errorf("failed to convert string to int: %w", err)
	}

	solved, err := strconv.ParseBool(chi.URLParam(r, "solved"))
	if err != nil {
		return fmt.Errorf("failed to parse bool: %w", err)
	}

	// set new solved-status in DB
	err = m.DB.UpdateReportStatus(solved, reportID)
	if err != nil {
		return fmt.Errorf("failed to set new report status: %w", err)
	}

	// get updated list of reports
	reports, err := m.DB.ReportsByMachineID(machineID)
	if err != nil {
		return fmt.Errorf("failed to get reports: %w", err)
	}

	// log success
	var status string
	if solved {
		status = "solved"
	} else {
		status = "open"
	}
	logger.Success(r, m.App.Session,
		fmt.Sprintf("status of report with id %d successfully changed to %s", reportID, status))

	// render new list to frontend
	return rendering(w, r, machines.ReportList(reports))
}

// HandleMachinesQRDelete handles the Machine QR code deletion
func (m *Repository) HandleMachineDeleteReport(w http.ResponseWriter, r *http.Request) error {
	// get url params
	machineID := chi.URLParam(r, "id")

	reportID, err := strconv.Atoi(chi.URLParam(r, "reportid"))
	if err != nil {
		return fmt.Errorf("failed to convert string to int: %w", err)
	}

	// set new solved-status in DB
	err = m.DB.DeleteReport(reportID)
	if err != nil {
		return fmt.Errorf("failed to delete report: %w", err)
	}

	// get updated list of reports
	reports, err := m.DB.ReportsByMachineID(machineID)
	if err != nil {
		return fmt.Errorf("failed to get reports: %w", err)
	}

	// log success
	logger.Success(r, m.App.Session,
		fmt.Sprintf("report with id %d successfully deleted", reportID))

	// render new list to frontend
	return rendering(w, r, machines.ReportList(reports))
}
