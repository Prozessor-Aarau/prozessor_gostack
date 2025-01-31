package models

import "time"

type Report struct {
	ID             int
	MachineID      string
	MachineName    string
	SlackChannel   string
	RessortID      int
	RessortName    string
	Fatal          bool
	ReportInfo     string
	ReportReporter string
	Solved         bool
	UpdatedOn      time.Time
	CreatedOn      time.Time
}
