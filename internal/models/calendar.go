package models

import "time"

type OpeningHour struct {
	ID          int
	StartTime   time.Time
	EndTime     time.Time
	Supervisors []Visitor
}

type CalEvent struct {
	ID        int
	Title     string
	Category  string
	StartTime time.Time
	EndTime   time.Time
	URL       string
}
