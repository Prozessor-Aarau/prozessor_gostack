package models

import "time"

type ReminderVisitor struct {
	Visitor     Visitor
	RessortID   int
	BadgeNumber int
	BuyDate     time.Time
	EndDate     time.Time
	Valid       bool
	Difference  int
	Reminder    time.Time
}

type Reminder struct {
	Count      int
	Progress   int
	Email      string
	FirstName  string
	FamilyName string
}
