package models

import "time"

type Log struct {
	ID        int
	Typ       string
	Module    string
	Message   string
	VisitorID string
	CreatedOn time.Time
}
