package models

import (
	"time"
)

type ChartVisitCount struct {
	Date       time.Time
	Total      int
	Daily      int
	Abonnement int
	Member     int
}
