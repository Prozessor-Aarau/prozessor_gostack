package models

import "time"

type Ressort struct {
	ID          int
	RessortName string
	CreatedOn   time.Time
	UpdatedOn   time.Time
}
