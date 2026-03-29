package events

import (
	"sort"

	"github.com/faebut/prozessor_gostack/internal/models"
)

func LayoutEvents(events []models.CalendarEntry) []models.CalendarEntry {
	sort.Slice(events, func(i, j int) bool {
		return events[i].GetStartTime().Before(events[j].GetStartTime())
	})

	groups := [][]models.CalendarEntry{}
	for _, e := range events {
		placed := false
		for i, g := range groups {
			for _, ge := range g {
				if e.GetStartTime().Before(ge.GetEndTime()) && e.GetEndTime().After(ge.
					GetStartTime()) {
					groups[i] = append(groups[i], e)
					placed = true
					break
				}
			}
			if placed {
				break
			}
		}
		if !placed {
			groups = append(groups, []models.CalendarEntry{e})
		}
	}

	result := []models.CalendarEntry{}
	for _, g := range groups {
		for col, e := range g {
			e.SetColumn(col)
			e.SetTotalCols(len(g))
			result = append(result, e)
		}
	}
	return result
}
