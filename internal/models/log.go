package models

import "time"

type Log struct {
	ID        int
	Level     string
	Location  string
	Message   string
	Visitor   Visitor
	CreatedOn time.Time
}
