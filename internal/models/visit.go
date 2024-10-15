package models

import (
	"time"
)

type Visit struct {
	ID         int
	Abonnement bool
	DateIn     time.Time
	DateOut    time.Time
	Partner    Partner
	Visitor    Visitor
	Ressort    Ressort
	Ateliers   []Atelier
	CreatedOn  time.Time
	UpdatedOn  time.Time
}
