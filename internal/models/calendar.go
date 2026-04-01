package models

import (
	"time"
)

var swissLoc = func() *time.Location {
	loc, err := time.LoadLocation("Europe/Zurich")
	if err != nil {
		return time.UTC
	}
	return loc
}()

type OpeningHour struct {
	ID          int
	StartTime   time.Time
	EndTime     time.Time
	Supervisors []Visitor
	Expertise   []string
	Column      int
	TotalCols   int
}

func (o *OpeningHour) GetStartTime() time.Time { return o.StartTime }
func (o *OpeningHour) GetEndTime() time.Time   { return o.EndTime }
func (o *OpeningHour) SetColumn(c int)         { o.Column = c }
func (o *OpeningHour) SetTotalCols(t int)      { o.TotalCols = t }
func (o *OpeningHour) GetColumn() int          { return o.Column }
func (o *OpeningHour) GetTotalCols() int       { return o.TotalCols }

type Event struct {
	Title     string    `yaml:"title"`
	Category  string    `yaml:"kategorie"`
	StartTime ParseDate `yaml:"event_start"`
	EndTime   ParseDate `yaml:"event_end"`
	Published bool      `yaml:"published"`
	URL       string
	Column    int
	TotalCols int
}

// type and func to parse Date from yaml
type ParseDate struct {
	time.Time
}

func (d *ParseDate) UnmarshalYAML(unmarshal func(any) error) error {
	var date string

	err := unmarshal(&date)
	if err != nil {
		return err
	}

	d.Time, err = time.Parse("2006-01-02 15:04", date)
	return err
}

func (e *Event) GetStartTime() time.Time {
	t := e.StartTime.Time
	return time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), 0, swissLoc)
}

func (e *Event) GetEndTime() time.Time {
	t := e.EndTime.Time
	return time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), 0, swissLoc)
}
func (e *Event) SetColumn(c int)    { e.Column = c }
func (e *Event) SetTotalCols(t int) { e.TotalCols = t }
func (e *Event) GetColumn() int     { return e.Column }
func (e *Event) GetTotalCols() int  { return e.TotalCols }

type CalendarEntry interface {
	GetStartTime() time.Time
	GetEndTime() time.Time
	SetColumn(int)
	SetTotalCols(int)
	GetColumn() int
	GetTotalCols() int
}
