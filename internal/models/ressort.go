package models

import "time"

type Ressort struct {
	ID           int
	RessortName  string
	SlackChannel string
	CreatedOn    time.Time
	UpdatedOn    time.Time
}
