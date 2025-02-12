package models

import (
	"time"
)

type BillData struct {
	ID            int
	Purpose       string
	Visitor       Visitor
	WithAddress   bool
	Firma         string
	FirstName     string
	FamilyName    string
	Street        string
	Postcode      string
	City          string
	Number        string
	Amount        float64
	Infos         string
	Items         []BillItem
	WithMessage   bool
	Message       string
	WithReference bool
	Reference     string
	FileName      string
	Paid          bool
	CreatedOn     time.Time
	UpdatedOn     time.Time
}

type BillItem struct {
	ID          string
	Position    int
	Description string
	UnitPrice   float64
	Count       float64
	Price       float64
	CreatedOn   time.Time
	UpdatedOn   time.Time
}
