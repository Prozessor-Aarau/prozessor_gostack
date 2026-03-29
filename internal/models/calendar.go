package models

import "time"

type OpeningHour struct {
	ID          int
	StartTime   time.Time
	EndTime     time.Time
	Supervisors []Visitor
	Column      int
	TotalCols   int
}

func (o *OpeningHour) GetStartTime() time.Time {
	return o.StartTime
}

func (o *OpeningHour) GetEndTime() time.Time {
	return o.EndTime
}

func (o *OpeningHour) SetColumn(c int) {
	o.Column = c
}

func (o *OpeningHour) SetTotalCols(t int) {
	o.TotalCols = t
}

type CalEvent struct {
	ID        int
	Title     string
	Category  string
	StartTime time.Time
	EndTime   time.Time
	URL       string
	Column    int
	TotalCols int
}

func (e *CalEvent) GetStartTime() time.Time {
	return e.StartTime
}

func (e *CalEvent) GetEndTime() time.Time {
	return e.EndTime
}

func (e *CalEvent) SetColumn(c int) {
	e.Column = c
}

func (e *CalEvent) SetTotalCols(t int) {
	e.TotalCols = t
}

type CalendarEntry interface {
	GetStartTime() time.Time
	GetEndTime() time.Time
	SetColumn(int)
	SetTotalCols(int)
}
