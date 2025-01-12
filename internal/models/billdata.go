package models

import "time"

type BillData struct {
	ID            int
	Purpose       string
	Visitor       Visitor
	WithAddress   bool
	FirstName     string
	FamilyName    string
	Street        string
	Postcode      string
	City          string
	Number        string
	Amount        float64
	Infos         string
	Products      []string
	WithMessage   bool
	Message       string
	WithReference bool
	Reference     string
	FileName      string
	Paid          bool
	CreatedOn     time.Time
	UpdatedOn     time.Time
}
