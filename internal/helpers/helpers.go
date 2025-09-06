package helpers

import (
	"fmt"
	"log"
	"math"
	"net/http"
	"runtime/debug"
	"strconv"
	"strings"
	"time"

	"github.com/faebut/prozessor_gostack/internal/config"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

var app *config.AppConfig

// NewHelpers sets the config for the helpers package
func NewHelpers(a *config.AppConfig) {
	app = a
}

func ClientError(w http.ResponseWriter, status int) {
	app.InfoLog.Println("Client error with status of", status)
	http.Error(w, http.StatusText(status), status)
}

func ServerError(w http.ResponseWriter, err error) {
	trace := fmt.Sprintf("%s\n%s", err.Error(), debug.Stack())
	app.ErrorLog.Println(trace)
	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

func IsAuthenticated(r *http.Request) bool {
	exists := app.Session.Exists(r.Context(), "user_id")
	return exists
}

// helper functions

// Iterate returns a slice of ints, starting at 1, going to count
func Iterate(count int) []int {
	var items []int
	for i := range count {
		items = append(items, i+1)
	}
	return items
}

// HumanDate returns time in YYYY-MM-DD format
func HumanDate(t time.Time) string {
	return t.Format("2006-01-02")
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
