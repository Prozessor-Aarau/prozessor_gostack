package models

import "time"

type Event struct {
	Title     string    `yaml:"title"`
	Category  string    `yaml:"kategorie"`
	EventTime ParseDate `yaml:"event_start"`
	Published bool      `yaml:"published"`
	URL       string
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
