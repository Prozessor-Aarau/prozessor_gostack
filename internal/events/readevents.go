package events

import (
	"io"
	"log"
	"strings"
	"time"

	"github.com/adrg/frontmatter"
	"github.com/jlaffaye/ftp"
)

type Event struct {
	Title     string    `yaml:"title"`
	Kategorie string    `yaml:"kategorie"`
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

func main() {
	var events []Event

	c, err := ftp.Dial("ftp.prozessor-aarau.ch:21", ftp.DialWithTimeout(5*time.Second))
	if err != nil {
		log.Fatal(err)
	}

	err = c.Login("kalender@prozessor-aarau.ch", "eCNHMvAvczBGJU#Bv#6j")
	if err != nil {
		log.Fatal(err)
	}

	// walk through folder
	walker := c.Walk("/")

	// current Date for comparison
	currentDate := time.Now()

	for walker.Next() {
		var e Event

		// check if it's a MarkDown file
		if strings.HasSuffix(walker.Path(), ".md") {
			// fill Event
			// split path by / and get first part for filepath
			path := strings.Split(walker.Path(), "/")[1]
			e.URL = path

			// read file and parse to event
			r, err := c.Retr(walker.Path())
			if err != nil {
				log.Println(err)
			}

			file, err := io.ReadAll(r)
			if err != nil {
				log.Fatal(err)
			}

			_, err = frontmatter.MustParse(strings.NewReader(string(file)), &e)
			if err != nil {
				log.Println(err)
			}

			// close the file
			if err = r.Close(); err != nil {
				log.Fatal(err)
			}

			// if it is in the future and published, add to slice
			if e.Published && e.EventTime.After(currentDate) {
				events = append(events, e)
			}
		}
	}

	if walker.Err() != nil && walker.Err() != io.EOF {
		log.Fatal(err)
	}

	// close connection
	if err = c.Quit(); err != nil {
		log.Fatal(err)
	}

	log.Printf("%v \n", events)
}
