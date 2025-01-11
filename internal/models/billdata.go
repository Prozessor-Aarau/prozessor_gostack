package models

type BillData struct {
	ID            int
	Purpose       string
	Visitor       Visitor
	WithAddress   bool
	Street        string
	Number        string
	Amount        float64
	Infos         string
	Products      []string
	WithMessage   bool
	Message       string
	WithReference bool
	Reference     string
	FileName      string
}
