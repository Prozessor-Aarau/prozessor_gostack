package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/faebut/prozessor_gostack/internal/forms"
	"github.com/faebut/prozessor_gostack/internal/logger"
	"github.com/faebut/prozessor_gostack/internal/models"
	"github.com/faebut/prozessor_gostack/view/report"
)

// HandleRegisterIndex handles the Register Page
func (m *Repository) HandleReportIndex(w http.ResponseWriter, r *http.Request) error {
	id := r.URL.Query().Get("machine")

	// get machine to pass it to the template
	machine, err := m.DB.MachineByID(id)
	if err != nil {
		return fmt.Errorf("failed to get machine: %w", err)
	}

	// render the template
	return rendering(w, r, report.Report(machine))
}

// HandleReportCreate is the post register handler
func (m *Repository) HandleReportCreate(w http.ResponseWriter, r *http.Request) error {
	err := r.ParseForm()
	if err != nil {
		return fmt.Errorf("failed to parse form: %w", err)
	}

	var reportForm models.Report

	form := forms.New(r.PostForm)
	reportForm.Fatal = form.BoxValue("fatal")

	reportForm.ReportInfo = r.Form.Get("report_infos")
	reportForm.ReportReporter = r.Form.Get("report_reporter")

	reportForm.MachineID = r.Form.Get("machine_id")
	// get machine to pass it to the template
	machine, err := m.DB.MachineByID(reportForm.MachineID)
	if err != nil {
		return fmt.Errorf("failed to get machine: %w", err)
	}

	// assign machine to reportForm
	reportForm.MachineName = machine.MachineName
	reportForm.SlackChannel = machine.SlackChannel
	reportForm.RessortID = machine.RessortID
	reportForm.RessortName = machine.RessortName
	// validation
	form.Required("report_infos")

	// what to do if form not valid
	if !form.Valid() {
		// return form with errors and allready filled fields
		return rendering(w, r, report.ReportForm(
			reportForm,
			report.ReportErrors{
				ReportInfo: form.Errors.Get("report_infos"),
			},
		))
	}

	// what to do if form valid
	// post report to database
	reportID, err := m.DB.CreateReport(reportForm)
	if err != nil {
		return fmt.Errorf("failed to create report: %w", err)
	}

	// post report to slack
	err = m.App.Slack.SendSlackReport(reportForm)
	if err != nil {
		return fmt.Errorf("failed to send report to slack: %w", err)
	}

	// log success
	logger.Success(r, m.App.Session,
		fmt.Sprintf("new report with id %d for machine %s successfully created", reportID, machine.MachineName))

	fatalstring := strconv.FormatBool(reportForm.Fatal)
	return hxRedirect(w, r, "/reportsuccess?fatal="+fatalstring)
}

// HandleReportSuccess handles the Report Success Page
func (m *Repository) HandleReportSuccess(w http.ResponseWriter, r *http.Request) error {
	// get param from Query to pass it on
	var fatal bool
	ftext := r.URL.Query().Get("fatal")

	if ftext == "true" {
		fatal = true
	} else {
		fatal = false
	}

	// render the template
	return rendering(w, r, report.ReportSuccess(fatal))
}
