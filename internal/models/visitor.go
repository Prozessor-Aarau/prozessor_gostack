package models

import (
	"time"
)

type Visitor struct {
	ID           string
	FirstName    string
	FamilyName   string
	Birthdate    time.Time
	Address      string
	Postcode     string
	City         string
	Email        string
	Phone        string
	Infos        string
	Password     string
	Ressort      Ressort
	RoleID       int
	Newsletter   bool
	Avatar       string
	AboLast      Abonnement
	AboLastValid bool
	AboAll       []Abonnement
	Badge        Badge
	Partner      []Partner
	CreatedOn    time.Time
	UpdatedOn    time.Time
}
