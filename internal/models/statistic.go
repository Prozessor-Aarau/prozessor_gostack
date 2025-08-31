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

type ChartVisitCounts []ChartVisitCount

func (c ChartVisitCounts) ChartAxis() []string {
	var list []string
	for _, item := range c {
		list = append(list, item.Date.Format("2006-01-02"))
	}
	return list
}

func (c ChartVisitCounts) ChartDailies() []int {
	var list []int
	for _, item := range c {
		list = append(list, item.Daily)
	}
	return list
}

func (c ChartVisitCounts) ChartAbonnements() []int {
	var list []int
	for _, item := range c {
		list = append(list, item.Abonnement)
	}
	return list
}

func (c ChartVisitCounts) ChartMembers() []int {
	var list []int
	for _, item := range c {
		list = append(list, item.Member)
	}
	return list
}
