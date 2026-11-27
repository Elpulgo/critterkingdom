package critter

import (
	"slices"
	"time"
)

// DailyVisitor picks today's daily visitor: one of the species flagged
// dailyVisitor, chosen from the calendar date of now in now's own location.
// The candidates take turns in id order, one per day, so the visitor stays
// the same all day and changes the next day whenever there are at least two.
// ok is false when no species is flagged.
func DailyVisitor(species []Species, now time.Time) (id string, ok bool) {
	var ids []string
	for i := range species {
		if species[i].DailyVisitor {
			ids = append(ids, species[i].ID)
		}
	}
	if len(ids) == 0 {
		return "", false
	}
	slices.Sort(ids)
	n := int64(len(ids))
	i := (dayNumber(now)%n + n) % n
	return ids[i], true
}

// dayNumber is the number of days from 1970-01-01 to now's calendar date in
// now's own location, negative before then.
func dayNumber(now time.Time) int64 {
	y, m, d := now.Date()
	midnight := time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
	return midnight.Unix() / (24 * 60 * 60)
}
