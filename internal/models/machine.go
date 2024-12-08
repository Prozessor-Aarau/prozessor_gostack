package models

import "time"

type Machine struct {
	ID           string
	RessortID    int
	RessortName  string
	SlackChannel string
	MachineName  string
	FAQs         []FAQ
	CreatedOn    time.Time
	UpdatedOn    time.Time
}

type FAQ struct {
	ID        int
	MachineID string
	Question  string
	Answer    string
	CreatedOn time.Time
	UpdatedOn time.Time
}

type MachineFilter struct {
	MachineName string
	RessortID   int
}
