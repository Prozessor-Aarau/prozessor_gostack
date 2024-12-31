package models

import "time"

type Role struct {
	ID        int
	RoleName  string
	CreatedOn time.Time
	UpdatedOn time.Time
	// persmissions
	ViewVisitors          bool
	EditVisitor           bool
	EditVisitorMembership bool
	EditVisitorRole       bool
	ViewMachines          bool
	ViewBadges            bool
	AssignBadge           bool
	EditBadge             bool
	ViewAbonnements       bool
	AddAbonnement         bool
	EditAbonnement        bool
	ViewReminders         bool
	AddReminder           bool
	ViewNewsletters       bool
	AddNewsletter         bool
	ViewLogistics         bool
	EditLogistics         bool
	ViewStatistics        bool
	ViewLogs              bool
	ViewFinances          bool
	AddFinances           bool
	EditFinances          bool
}
