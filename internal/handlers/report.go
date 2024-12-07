package handlers

import (
	"log"
	"net/http"
	"strconv"

	"github.com/faebut/prozessor_gostack/internal/forms"
	"github.com/faebut/prozessor_gostack/view/report"
)

// HandleRegisterIndex handles the Register Page
func (m *Repository) HandleReportIndex(w http.ResponseWriter, r *http.Request) error {
	// render the template
	return rendering(w, r, report.Report())
}

// HandleReportCreate is the post register handler
func (m *Repository) HandleReportCreate(w http.ResponseWriter, r *http.Request) error {
	err := r.ParseForm()
	if err != nil {
		log.Println(err)
		// TODO better error handling
		return nil
	}

	var reportForm report.ReportParams

	form := forms.New(r.PostForm)
	reportForm.Fatal = form.BoxValue("fatal")

	reportForm.ReportInfo = r.Form.Get("report_infos")
	reportForm.ReportReporter = r.Form.Get("report_reporter")

	// validation
	form.Required("report_infos")

	// what to do if form not valid
	if !form.Valid() {
		// return form with errors and allready filled fields
		return rendering(w, r, report.ReportForm(reportForm, report.ReportErrors{
			ReportInfo: form.Errors.Get("report_infos"),
		}))
	}

	// what to do if form valid
	// post report to database
	log.Println(reportForm)

	// post report to slack

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
