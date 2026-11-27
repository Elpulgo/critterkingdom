package critter_test

import (
	"testing"

	"github.com/Elpulgo/critterkingdom/internal/critter"
)

func TestSpeciesActive_OnlyWhenEveryConditionHolds(t *testing.T) {
	tests := []struct {
		name    string
		species critter.Species
		world   critter.World
		want    bool
	}{
		{
			name:    "night-only critter is about at night",
			species: critter.Species{ActiveAt: []critter.TimeOfDay{critter.Night}},
			world:   critter.World{TimeOfDay: critter.Night},
			want:    true,
		},
		{
			name:    "night-only critter hides by day",
			species: critter.Species{ActiveAt: []critter.TimeOfDay{critter.Night}},
			world:   critter.World{TimeOfDay: critter.Day},
			want:    false,
		},
		{
			name:    "day-only critter is about by day",
			species: critter.Species{ActiveAt: []critter.TimeOfDay{critter.Day}},
			world:   critter.World{TimeOfDay: critter.Day},
			want:    true,
		},
		{
			name:    "day-only critter hides at night",
			species: critter.Species{ActiveAt: []critter.TimeOfDay{critter.Day}},
			world:   critter.World{TimeOfDay: critter.Night},
			want:    false,
		},
		{
			name:    "critter without activeAt is about by day",
			species: critter.Species{},
			world:   critter.World{TimeOfDay: critter.Day},
			want:    true,
		},
		{
			name:    "critter without activeAt is about at night",
			species: critter.Species{},
			world:   critter.World{TimeOfDay: critter.Night},
			want:    true,
		},
		{
			name:    "quest critter appears once its flag is set",
			species: critter.Species{NeedsFlag: "bridge_fixed"},
			world:   critter.World{TimeOfDay: critter.Day, Flags: map[string]bool{"bridge_fixed": true}},
			want:    true,
		},
		{
			name:    "quest critter stays away while its flag is unset",
			species: critter.Species{NeedsFlag: "bridge_fixed"},
			world:   critter.World{TimeOfDay: critter.Day, Flags: map[string]bool{"met_owl": true}},
			want:    false,
		},
		{
			name:    "quest critter stays away when no flags are set",
			species: critter.Species{NeedsFlag: "bridge_fixed"},
			world:   critter.World{TimeOfDay: critter.Day},
			want:    false,
		},
		{
			name:    "companion critter appears with the right companion",
			species: critter.Species{NeedsCompanion: "fox"},
			world:   critter.World{TimeOfDay: critter.Day, Companion: "fox"},
			want:    true,
		},
		{
			name:    "companion critter stays away with the wrong companion",
			species: critter.Species{NeedsCompanion: "fox"},
			world:   critter.World{TimeOfDay: critter.Day, Companion: "firefly"},
			want:    false,
		},
		{
			name:    "companion critter stays away with no companion",
			species: critter.Species{NeedsCompanion: "fox"},
			world:   critter.World{TimeOfDay: critter.Day},
			want:    false,
		},
		{
			name: "critter with every condition met appears",
			species: critter.Species{
				ActiveAt:       []critter.TimeOfDay{critter.Night},
				NeedsFlag:      "bridge_fixed",
				NeedsCompanion: "fox",
			},
			world: critter.World{
				TimeOfDay: critter.Night,
				Flags:     map[string]bool{"bridge_fixed": true},
				Companion: "fox",
			},
			want: true,
		},
		{
			name: "critter with conditions stays away when only the time is wrong",
			species: critter.Species{
				ActiveAt:       []critter.TimeOfDay{critter.Night},
				NeedsFlag:      "bridge_fixed",
				NeedsCompanion: "fox",
			},
			world: critter.World{
				TimeOfDay: critter.Day,
				Flags:     map[string]bool{"bridge_fixed": true},
				Companion: "fox",
			},
			want: false,
		},
		{
			name: "critter with conditions stays away when only the flag is missing",
			species: critter.Species{
				ActiveAt:       []critter.TimeOfDay{critter.Night},
				NeedsFlag:      "bridge_fixed",
				NeedsCompanion: "fox",
			},
			world: critter.World{
				TimeOfDay: critter.Night,
				Companion: "fox",
			},
			want: false,
		},
		{
			name: "critter with conditions stays away when only the companion is wrong",
			species: critter.Species{
				ActiveAt:       []critter.TimeOfDay{critter.Night},
				NeedsFlag:      "bridge_fixed",
				NeedsCompanion: "fox",
			},
			world: critter.World{
				TimeOfDay: critter.Night,
				Flags:     map[string]bool{"bridge_fixed": true},
				Companion: "firefly",
			},
			want: false,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.species.Active(tc.world); got != tc.want {
				t.Errorf("Active(%+v) = %v, want %v", tc.world, got, tc.want)
			}
		})
	}
}
