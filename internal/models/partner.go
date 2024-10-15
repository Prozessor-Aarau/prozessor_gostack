package models

import "time"

type Partner struct {
	ID                int
	PartnerName       string
	ContactFirstName  string
	ContactFamilyName string
	Email             string
	Address           string
	Postcode          string
	City              string
	Description       string
	Active            bool
	CreatedOn         time.Time
	UpdatedOn         time.Time
}
