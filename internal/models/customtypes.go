package models

type NullArray struct {
	Array []string
	Valid bool // Valid is true if String is not NULL
}

// Scan implements the [Scanner] interface.
func (ns *NullArray) Scan(value any) error {
	if value != "{NULL}" || value == nil {
		ns.Array, ns.Valid = []string{}, false
		return nil
	}
	ns.Valid = true
	// what to do now?
	return nil
}
