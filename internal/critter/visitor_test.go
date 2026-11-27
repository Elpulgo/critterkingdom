package critter_test

import (
	"slices"
	"testing"
	"time"

	"github.com/Elpulgo/critterkingdom/internal/critter"
)

func TestDailyVisitor_IsTheSameAllDayAndChangesTheNextDay(t *testing.T) {
	// Three flagged visitors take turns in id order: badger, hedgehog, owl.
	// 2026-11-27 is day 20784 since 1970-01-01, and 20784 % 3 == 0, so that
	// day is the badger's.
	visitors := []critter.Species{
		{ID: "owl", DailyVisitor: true},
		{ID: "firefly"},
		{ID: "hedgehog", DailyVisitor: true},
		{ID: "badger", DailyVisitor: true},
	}
	lonely := []critter.Species{{ID: "firefly"}, {ID: "owl", DailyVisitor: true}}
	utcMinus5 := time.FixedZone("UTC-5", -5*60*60)
	utcPlus9 := time.FixedZone("UTC+9", 9*60*60)

	tests := []struct {
		name    string
		species []critter.Species
		now     time.Time
		wantID  string
		wantOK  bool
	}{
		{
			name:    "first minute of the day",
			species: visitors,
			now:     time.Date(2026, 11, 27, 0, 0, 0, 0, time.UTC),
			wantID:  "badger", wantOK: true,
		},
		{
			name:    "last minute of the same day",
			species: visitors,
			now:     time.Date(2026, 11, 27, 23, 59, 59, 0, time.UTC),
			wantID:  "badger", wantOK: true,
		},
		{
			name:    "next day brings the next visitor",
			species: visitors,
			now:     time.Date(2026, 11, 28, 0, 0, 0, 0, time.UTC),
			wantID:  "hedgehog", wantOK: true,
		},
		{
			name:    "day after that",
			species: visitors,
			now:     time.Date(2026, 11, 29, 12, 0, 0, 0, time.UTC),
			wantID:  "owl", wantOK: true,
		},
		{
			name:    "turns wrap around to the first visitor",
			species: visitors,
			now:     time.Date(2026, 11, 30, 12, 0, 0, 0, time.UTC),
			wantID:  "badger", wantOK: true,
		},
		{
			// 2026-11-28 03:00 UTC is still 2026-11-27 22:00 at UTC-5.
			name:    "late evening west of UTC counts as the local date",
			species: visitors,
			now:     time.Date(2026, 11, 28, 3, 0, 0, 0, time.UTC).In(utcMinus5),
			wantID:  "badger", wantOK: true,
		},
		{
			// 2026-11-27 23:30 UTC is already 2026-11-28 08:30 at UTC+9.
			name:    "morning east of UTC counts as the local date",
			species: visitors,
			now:     time.Date(2026, 11, 27, 23, 30, 0, 0, time.UTC).In(utcPlus9),
			wantID:  "hedgehog", wantOK: true,
		},
		{
			// 1969-12-31 is day -1, which is turn 2 of 3.
			name:    "a clock set before 1970 still picks a visitor",
			species: visitors,
			now:     time.Date(1969, 12, 31, 12, 0, 0, 0, time.UTC),
			wantID:  "owl", wantOK: true,
		},
		{
			name:    "a single flagged species visits today",
			species: lonely,
			now:     time.Date(2026, 11, 27, 12, 0, 0, 0, time.UTC),
			wantID:  "owl", wantOK: true,
		},
		{
			name:    "a single flagged species visits tomorrow too",
			species: lonely,
			now:     time.Date(2026, 11, 28, 12, 0, 0, 0, time.UTC),
			wantID:  "owl", wantOK: true,
		},
		{
			name:    "no flagged species means no visitor",
			species: []critter.Species{{ID: "firefly"}, {ID: "frog"}},
			now:     time.Date(2026, 11, 27, 12, 0, 0, 0, time.UTC),
			wantID:  "", wantOK: false,
		},
		{
			name:   "no species at all means no visitor",
			now:    time.Date(2026, 11, 27, 12, 0, 0, 0, time.UTC),
			wantID: "", wantOK: false,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			reversed := slices.Clone(tc.species)
			slices.Reverse(reversed)
			// The order of the species list must not change the visitor.
			for _, species := range [][]critter.Species{tc.species, reversed} {
				id, ok := critter.DailyVisitor(species, tc.now)
				if id != tc.wantID || ok != tc.wantOK {
					t.Errorf("DailyVisitor(%v, %v) = (%q, %v), want (%q, %v)",
						ids(species), tc.now, id, ok, tc.wantID, tc.wantOK)
				}
			}
		})
	}
}

func ids(species []critter.Species) []string {
	out := make([]string, len(species))
	for i := range species {
		out[i] = species[i].ID
	}
	return out
}
