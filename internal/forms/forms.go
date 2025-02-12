package forms

import (
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"unicode"
)

// Form creates a custom form struct, embeds a url.Values object
type Form struct {
	url.Values
	Errors errors
}

// regex expressions
var EmailRx = regexp.MustCompile(`[A-Z0-9a-z._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,64}`)
var PhoneRx = regexp.MustCompile(`^[+0-9 ]{10,18}$`)

// Valid returns true if there are no errors, otherwise false
func (f *Form) Valid() bool {
	return len(f.Errors) == 0
}

// New initializes a form struct
func New(data url.Values) *Form {
	return &Form{
		data,
		errors(map[string][]string{}),
	}
}

// Required checks for required fields
func (f *Form) Required(fields ...string) {
	for _, field := range fields {
		value := f.Get(field)
		if strings.TrimSpace(value) == "" {
			f.Errors.Add(field, "Feld ist obligatorisch")
		}
	}
}

// Has checks if form field is in post and not empty
func (f *Form) Has(field string) bool {
	x := f.Get(field)

	return x != ""
}

// MinLength checks for string minimum length
func (f *Form) MinLength(field string, length int) bool {
	x := f.Get(field)
	if len(x) < length {
		f.Errors.Add(field, fmt.Sprintf("This field must be at least %d characters long", length))
		return false
	}
	return true
}

// checks if field is empty
func (f *Form) IsEmpty(field string) bool {
	return strings.TrimSpace(f.Get(field)) == ""
}

// IsEmail checks for valid email address
func (f *Form) IsEmail(field string) {

	if !f.IsEmpty(field) {
		if !EmailRx.MatchString(f.Get(field)) {
			f.Errors.Add(field, "korrekte E-Mailadresse angeben")
		}
	}
}

// IsPhone checks for valid phone number
func (f *Form) IsPhone(field string) {
	if strings.TrimSpace(f.Get(field)) != "" {
		if !PhoneRx.MatchString(f.Get(field)) {
			f.Errors.Add(field, "korrekte Telefonnummer angeben")
		}
	}
}

// OneOf checks if at least one of the fields is filled
func (f *Form) OneOf(fields ...string) {

	var exists = func(check []string) bool {
		for _, field := range check {
			value := f.Get(field)
			if strings.TrimSpace(value) != "" {
				return true
			}
		}
		return false
	}

	if !exists(fields) {
		for _, field := range fields {
			f.Errors.Add(field, "Mindestens eine Angabe notwendig")
		}
	}
}

// BoxValue checks value of checkbox and returns boolean
func (f *Form) BoxValue(field string) bool {
	var boxValue bool

	switch strings.ToLower(f.Get(field)) {
	case "yes", "on":
		boxValue = true
	case "no", "off":
		boxValue = false
	default:
		boxValue = false
	}
	return boxValue
}

// IsPassword checks for valid password string
func (f *Form) IsPassword(field string) {

	pwcheck := func(s string) bool {
		var (
			hasMinLen  = false
			hasUpper   = false
			hasLower   = false
			hasNumber  = false
			hasSpecial = false
		)
		if len(s) >= 8 {
			hasMinLen = true
		}
		for _, char := range s {
			switch {
			case unicode.IsUpper(char):
				hasUpper = true
			case unicode.IsLower(char):
				hasLower = true
			case unicode.IsNumber(char):
				hasNumber = true
			case unicode.IsPunct(char) || unicode.IsSymbol(char):
				hasSpecial = true
			}
		}
		return hasMinLen && hasUpper && hasLower && hasNumber && hasSpecial
	}

	if !f.IsEmpty(field) {
		if !pwcheck(f.Get(field)) {
			f.Errors.Add(
				field,
				"Passwort muss mindestens 8 Zeichen lang sein, einen Gross- und einen Kleinbuchstaben, eine Zahl sowie ein Sonderzeichen enthalten.",
			)
		}
	}
}

// matches checks for valid password string
func (f *Form) IsMatch(field, comparison string) {

	if !f.IsEmpty(field) || !f.IsEmpty(comparison) {
		if f.Get(field) != f.Get(comparison) {
			f.Errors.Add(
				field,
				"Passwörter stimmen nicht überein",
			)
		}
	}
}

// IsHigher checks, if number is higher
func (f *Form) IsHigher(field string, value int) {
	if !f.IsEmpty(field) {
		number, _ := strconv.Atoi(f.Get(field))
		if number <= value {
			f.Errors.Add(
				field,
				fmt.Sprintf("Nummer muss grösser als %d sein", value),
			)
		}
	}
}

// IsPostcode checks, if number is CH postcode
func (f *Form) IsPostcode(field string) {
	if !f.IsEmpty(field) {
		number, _ := strconv.Atoi(f.Get(field))
		if number < 1000 || number > 9999 {
			f.Errors.Add(
				field,
				"Gültige CH Postleitzahl angeben",
			)
		}
	}
}
