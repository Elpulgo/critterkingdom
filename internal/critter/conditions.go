package critter

import "slices"

// World is the outside state that decides which critters can appear. The
// caller builds it from the clock, quest log and companion each time it asks.
type World struct {
	// TimeOfDay is the current part of the day.
	TimeOfDay TimeOfDay
	// Flags holds the quest flags that are set.
	Flags map[string]bool
	// Companion is the species id of the critter following the player, or
	// empty when none is.
	Companion string
	// Visitor is the species id of today's daily visitor (see DailyVisitor),
	// or empty when there is none.
	Visitor string
}

// Active reports whether the species can appear in w: every condition it
// states must hold. A species with no activeAt is about at any time.
func (sp *Species) Active(w World) bool {
	if len(sp.ActiveAt) > 0 && !slices.Contains(sp.ActiveAt, w.TimeOfDay) {
		return false
	}
	if sp.NeedsFlag != "" && !w.Flags[sp.NeedsFlag] {
		return false
	}
	if sp.NeedsCompanion != "" && w.Companion != sp.NeedsCompanion {
		return false
	}
	if sp.DailyVisitor && w.Visitor != sp.ID {
		return false
	}
	return true
}
