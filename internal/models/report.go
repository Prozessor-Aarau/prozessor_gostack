package models

type Report struct {
	MachineID      string
	MachineName    string
	SlackChannel   string
	RessortID      int
	RessortName    string
	Fatal          bool
	ReportInfo     string
	ReportReporter string
}
