package events

import (
	"io"
	"sort"
	"strings"
	"time"

	"github.com/adrg/frontmatter"
	"github.com/faebut/prozessor_gostack/internal/models"
	"github.com/jlaffaye/ftp"
)

// GetEvents returns the events parsed from the website
func GetEvents() ([]models.Event, error) {
	var events []models.Event

	c, err := ftp.Dial("ftp.prozessor-aarau.ch:21", ftp.DialWithTimeout(5*time.Second))
	if err != nil {
		return nil, err
	}

	err = c.Login("kalender@prozessor-aarau.ch", "eCNHMvAvczBGJU#Bv#6j")
	if err != nil {
		return nil, err
	}

	// walk through folder
	walker := c.Walk("/")

	// current Date for comparison
	currentDate := time.Now()

	for walker.Next() {
		var e models.Event

		// check if it's a MarkDown file
		if strings.HasSuffix(walker.Path(), ".md") {
			// fill Event
			// split path by / and get first part for filepath
			path := strings.Split(walker.Path(), "/")[1]
			e.URL = "https://prozessor-aarau.ch/kalender/" + path

			// read file and parse to event
			r, err := c.Retr(walker.Path())
			if err != nil {
				return nil, err
			}

			file, err := io.ReadAll(r)
			if err != nil {
				return nil, err
			}

			_, err = frontmatter.MustParse(strings.NewReader(string(file)), &e)
			if err != nil {
				return nil, err
			}

			// close the file
			if err = r.Close(); err != nil {
				return nil, err
			}

			// if it is in the future and published, add to slice
			// TODO: Filter for specific timeframe. (eg. max 2 months)
			if e.Published && e.EventTime.After(currentDate) {
				events = append(events, e)
			}
		}
	}

	if walker.Err() != nil && walker.Err() != io.EOF {
		return nil, err
	}

	// close connection
	if err = c.Quit(); err != nil {
		return nil, err
	}

	// Sort Events
	sort.Slice(events, func(i, j int) bool {
		return events[i].EventTime.Time.Before(events[j].EventTime.Time)
	})

	return events, nil
}
