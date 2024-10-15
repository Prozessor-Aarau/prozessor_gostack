package models

import "time"

type Abonnement struct {
	ID                int
	BuyDate           string
	PartialPayment    bool
	Reminder          time.Time
	VisitorID         string
	VisitorFirstName  string
	VisitorFamilyName string
	Price             int
	Member            bool
	reminder          string
	CreatedOn         time.Time
	UpdatedOn         time.Time
}
