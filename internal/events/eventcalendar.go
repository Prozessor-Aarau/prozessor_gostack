package events

import (
	"sort"
	"time"

	"github.com/faebut/prozessor_gostack/internal/models"
)

func LayoutEvents(events []models.CalendarEntry) []models.CalendarEntry {
	sort.Slice(events, func(i, j int) bool {
		return events[i].GetStartTime().Before(events[j].GetStartTime())
	})

	// first pass: assign columns using per-column end-time tracking
	colEnds := []time.Time{}
	for _, e := range events {
		assigned := -1
		for col, end := range colEnds {
			if !e.GetStartTime().Before(end) {
				assigned = col
				colEnds[col] = e.GetEndTime()
				break
			}
		}
		if assigned == -1 {
			assigned = len(colEnds)
			colEnds = append(colEnds, e.GetEndTime())
		}
		e.SetColumn(assigned)
	}

	// second pass: find TotalCols for each event — the max number of
	// simultaneously active columns across each event's time range
	for _, e := range events {
		maxCols := 1
		for _, other := range events {
			if other.GetStartTime().Before(e.GetEndTime()) && other.GetEndTime().After(e.GetStartTime()) {
				if other.GetColumn()+1 > maxCols {
					maxCols = other.GetColumn() + 1
				}
			}
		}
		e.SetTotalCols(maxCols)
	}

	return events
}
