package render

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"log"
	"math"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/faebut/prozessor_gostack/internal/config"
	"github.com/faebut/prozessor_gostack/internal/models"
	"github.com/justinas/nosurf"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

var functions = template.FuncMap{
	"humanDate":   HumanDate,
	"formatDate":  FormatDate,
	"realDate":    RealDate,
	"timeElapsed": TimeElapsed,
	"today":       Today,
	"iterate":     Iterate,
	"add":         Add,
	"sub":         Sub,
	"initials":    Initials,
	"toToday":     ToToday,
	"mkSlice":     MkSlice,
	"contains":    Contains,
}

var app *config.AppConfig
var pathToTemplates = "./templates"

// TODO remove helper functions from here and put it into helpers
// Add two ints together
func Add(a, b int) int {
	return a + b
}

// subtracts one int from an int
func Sub(a, b int) int {
	return a - b
}

// Iterate returns a slice of ints, starting at 1, going to count
func Iterate(count int) []int {
	var i int
	var items []int
	for i = 0; i < count; i++ {
		items = append(items, i+1)
	}
	return items
}

// NewRenderer sets the config for the template package
func NewRenderer(a *config.AppConfig) {
	app = a
}

// HumanDate returns time in YYYY-MM-DD format
func HumanDate(t time.Time) string {
	return t.Format("2006-01-02")
}

// FormatDate returns a formatted date
func FormatDate(t time.Time, f string) string {
	return t.Format(f)
}

// RealDate returns if a time is before a certain time
func RealDate(t time.Time) bool {
	compareDate := time.Date(1910, 1, 1, 0, 0, 0, 0, time.UTC)
	return !t.Before(compareDate)
}

// format Numbers using the swiss locale
var swissPrinter = message.NewPrinter(language.Make("de-CH"))

// FormatInt returns a swiss formated int
func FormatInt(num int) string {
	return swissPrinter.Sprintf("%d", num)
}

// FormatFloat returns a swiss formated float with a specified number of decimals
func FormatFloat(num float64, decimals int) string {
	format := fmt.Sprintf("%%.%df", decimals)
	return swissPrinter.Sprintf(format, num)
}

// FormatCurrency returns a number as formated swiss currency
func FormatCurrency(num float64) string {
	return swissPrinter.Sprintf("CHF %.2f", num)
}

// TimeElapsed returns the elapsed hours and minutes in 00:00 format from a timestamp to now
func TimeElapsed(t time.Time) string {
	// if less than 9 put a 0 in front of number
	doubleDigit := func(d int) string {
		if d < 10 {
			return fmt.Sprintf("0%d", d)
		} else {
			return strconv.Itoa(d)
		}
	}

	timeD := time.Since(t)
	hours := int(math.Floor(timeD.Hours()))
	minutes := int(timeD.Minutes()) - (hours * 60)

	diffOut := fmt.Sprintf("%s:%s", doubleDigit(hours), doubleDigit(minutes))

	return diffOut
}

// Today returns the formatted today date
func Today() string {
	return time.Now().Format("2. Jan 2006")
}

// Initials returns first Letters of String as new string
func Initials(s ...string) string {
	var initials string
	for _, v := range s {
		// get first Character and make it uppercase
		firstLetter := strings.ToUpper(v[0:1])

		// append it to Initials
		initials += firstLetter
	}
	return initials
}

// ToToday returns the amount of days to today

func ToToday(s string) int {
	today := time.Now()
	date, err := time.Parse("2006-01-02", s)
	if err != nil {
		log.Println("Error converting time-string: ", err)
	}

	date = date.AddDate(1, 0, 0)

	diffDays := date.Sub(today).Hours() / 24

	return int(diffDays)
}

func MkSlice(args ...interface{}) []interface{} {
	return args
}

func Contains(num int, slice []int) bool {
	for _, a := range slice {
		if a == num {
			return true
		}
	}
	return false
}

func AddDefaultData(td *models.TemplateData, r *http.Request) *models.TemplateData {
	// place to add data that every page needs
	td.Flash = app.Session.PopString(r.Context(), "flash")
	td.Error = app.Session.PopString(r.Context(), "error")
	td.Warning = app.Session.PopString(r.Context(), "warning")
	td.CSRFToken = nosurf.Token(r)
	if app.Session.Exists(r.Context(), "user_id") {
		td.IsAuthenticated = 1
	}
	return td
}

// Template renders a template
func Template(w http.ResponseWriter, r *http.Request, tmpl string, td *models.TemplateData) error {
	var tc map[string]*template.Template

	if app.UseCache {
		// get the template cache from the app config
		tc = app.TemplateCache
	} else {
		tc, _ = CreateTemplateCache()
	}

	// get requested template from cache
	t, ok := tc[tmpl]
	if !ok {
		return errors.New("could not get template from cache")
	}

	buf := new(bytes.Buffer)

	td = AddDefaultData(td, r)
	_ = t.Execute(buf, td)

	// render the template

	_, err := buf.WriteTo(w)
	if err != nil {
		log.Println("Error writing template to browser", err)
		return err
	}
	return nil
}

// CreateTemplateCache creates a template cache as a map
func CreateTemplateCache() (map[string]*template.Template, error) {
	myCache := map[string]*template.Template{}

	// get all of the files named *.page.tmpl.html from ./templates
	pages, err := filepath.Glob(fmt.Sprintf("%s/views/*.html", pathToTemplates))

	if err != nil {
		return myCache, err
	}

	// range through all files ending with .page.html
	for _, page := range pages {
		name := filepath.Base(page)
		ts, err := template.New(name).Funcs(functions).ParseFiles(page)

		if err != nil {
			return myCache, err
		}

		matches, err := filepath.Glob(fmt.Sprintf("%s/layout/*.html", pathToTemplates))

		if err != nil {
			return myCache, err
		}

		if len(matches) > 0 {
			ts, err = ts.ParseGlob(fmt.Sprintf("%s/layout/*.html", pathToTemplates))
			if err != nil {
				return myCache, err
			}
		}

		myCache[name] = ts
	}
	return myCache, nil
}
