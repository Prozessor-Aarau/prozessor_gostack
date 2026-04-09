package calendar

import (
	"fmt"
	"time"

	"github.com/faebut/prozessor_gostack/internal/models"
)

var weekdayDE = map[time.Weekday]string{
	time.Monday:    "Mo",
	time.Tuesday:   "Di",
	time.Wednesday: "Mi",
	time.Thursday:  "Do",
	time.Friday:    "Fr",
	time.Saturday:  "Sa",
	time.Sunday:    "So",
}

var weekdayDELong = map[time.Weekday]string{
	time.Monday:    "Montag",
	time.Tuesday:   "Dienstag",
	time.Wednesday: "Mittwoch",
	time.Thursday:  "Donnerstag",
	time.Friday:    "Freitag",
	time.Saturday:  "Samstag",
	time.Sunday:    "Sonntag",
}

func WeekdayDE(t time.Time) string {
	return weekdayDELong[t.In(SwissLoc).Weekday()]
}

var SwissLoc = func() *time.Location {
	loc, err := time.LoadLocation("Europe/Zurich")
	if err != nil {
		return time.UTC
	}
	return loc
}()

func CalDayHeader(t time.Time) string {
	return fmt.Sprintf("%s %s", weekdayDE[t.Weekday()], t.Format("02.01"))
}

func IsToday(t time.Time) bool {
	now := time.Now().In(SwissLoc)
	y1, m1, d1 := now.Date()
	y2, m2, d2 := t.In(SwissLoc).Date()
	return y1 == y2 && m1 == m2 && d1 == d2
}

func CalSlotLabel(slot, startHour, slotMinutes int) string {
	totalMinutes := startHour*60 + slot*slotMinutes
	return fmt.Sprintf("%02d:%02d", totalMinutes/60, totalMinutes%60)
}

func calStartSlot(t time.Time, startHour, slotMinutes int) int {
	local := t.In(SwissLoc)
	return ((local.Hour()-startHour)*60 + local.Minute()) / slotMinutes
}

func calDurationSlots(start, end time.Time, slotMinutes int) int {
	slots := int(end.Sub(start).Minutes()) / slotMinutes
	if slots < 1 {
		return 1
	}
	return slots
}

func CalEntriesForDay(entries []models.CalendarEntry, day time.Time) []models.CalendarEntry {
	y, m, d := day.In(SwissLoc).Date()
	var result []models.CalendarEntry
	for _, e := range entries {
		ey, em, ed := e.GetStartTime().In(SwissLoc).Date()
		if ey == y && em == m && ed == d {
			result = append(result, e)
		}
	}
	return result
}

func CalEventStyle(entry models.CalendarEntry, startHour, slotMinutes int) string {
	startSlot := calStartSlot(entry.GetStartTime(), startHour, slotMinutes)
	durationSlots := calDurationSlots(entry.GetStartTime(), entry.GetEndTime(), slotMinutes)
	return fmt.Sprintf("--start-slot:%d;--duration-slots:%d;--col:%d;--total-cols:%d",
		startSlot, durationSlots, entry.GetColumn(), entry.GetTotalCols())
}
