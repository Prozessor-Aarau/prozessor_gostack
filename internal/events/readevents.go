package events

import (
	"io"
	"log/slog"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/adrg/frontmatter"
	"github.com/faebut/prozessor_gostack/internal/logger"
	"github.com/faebut/prozessor_gostack/internal/models"
	"github.com/jlaffaye/ftp"
	"github.com/joho/godotenv"
)

// GetEvents returns the events parsed from the website
func GetEvents() ([]models.Event, error) {
	var events []models.Event

	err := godotenv.Load()
	if err != nil {
		logger.Fatal("loading .env file", slog.String("error", err.Error()))
	}

	c, err := ftp.Dial(os.Getenv("FTP_HOST"), ftp.DialWithTimeout(5*time.Second))
	if err != nil {
		return nil, err
	}

	err = c.Login(os.Getenv("FTP_USER"), os.Getenv("FTP_PASS"))
	if err != nil {
		return nil, err
	}

	// walk through folder
	walker := c.Walk("/")

	// current Date for comparison
	currentDate := time.Now()

	// limit search for the next two months. If it's the beginning of the month
	// use this and next month else the next two months
	y, m, d := currentDate.Date()
	if d < 10 {
		m = m + 2
	} else {
		m = m + 3
	}
	maxDate := time.Date(y, m, 0, 0, 0, 0, 0, time.Local)

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
			if e.Published && e.EventTime.After(currentDate) && e.EventTime.Before(maxDate) {
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
