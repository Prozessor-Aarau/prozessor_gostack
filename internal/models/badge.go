package models

import "time"

type Badge struct {
	ID                int
	Number            int
	UID               string
	VisitorID         string
	VisitorFirstName  string
	VisitorFamilyName string
	VisitorAboValid   bool
	ExternFirstName   string
	ExternFamilyName  string
	ExternCity        string
	ExternPhone       string
	ExternEmail       string
	ExternPurpose     string
	CreatedOn         time.Time
	UpdatedOn         time.Time
}
