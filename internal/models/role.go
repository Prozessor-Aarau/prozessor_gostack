package models

import "time"

type Role struct {
	ID        int
	RoleName  string
	CreatedOn time.Time
	UpdatedOn time.Time
}
