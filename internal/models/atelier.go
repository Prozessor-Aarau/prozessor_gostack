package models

import "time"

type Atelier struct {
	ID          int
	AtelierName string
	Description string
	Active      bool
	Icon        string
	CreatedOn   time.Time
	UpdatedOn   time.Time
}
