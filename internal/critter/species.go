// Package critter holds the rules for the friendly critters: species data,
// when they appear, and befriending them with a snack. It is pure Go so every
// rule can be tested without a window.
package critter

import (
	"encoding/json"
	"errors"
	"fmt"
)

// Text is a string in each language, keyed by language code such as "en".
type Text map[string]string

// TimeOfDay is the part of the day a critter can be active in.
type TimeOfDay string

// The times of day a species can list in activeAt.
const (
	Day   TimeOfDay = "day"
	Night TimeOfDay = "night"
)

// Rarity says how often a species is drawn when a critter spawns.
type Rarity string

// The rarities a species can have.
const (
	Common   Rarity = "common"
	Uncommon Rarity = "uncommon"
	Rare     Rarity = "rare"
)

// Species is one kind of critter as described in critters.json.
type Species struct {
	ID      string `json:"id"`
	Name    Text   `json:"name"`
	Habitat string `json:"habitat"`
	// ActiveAt lists the times of day the species is about; empty means any
	// time.
	ActiveAt []TimeOfDay `json:"activeAt"`
	// Likes is the item id of the snack that befriends it.
	Likes   string `json:"likes"`
	Rarity  Rarity `json:"rarity"`
	Sprite  string `json:"sprite"`
	Ability string `json:"ability"`
	Fact    Text   `json:"fact"`

	// NeedsFlag, when set, is a quest flag that must be set for the species
	// to appear.
	NeedsFlag string `json:"needsFlag,omitempty"`
	// NeedsCompanion, when set, is the species id of the companion the player
	// must have with them for the species to appear.
	NeedsCompanion string `json:"needsCompanion,omitempty"`
	// DailyVisitor marks a species that only appears on the days it is the
	// daily visitor.
	DailyVisitor bool `json:"dailyVisitor,omitempty"`
}

// LoadSpecies parses critters.json, a JSON array of species, and checks each
// species for the mistakes it can see on its own: a missing id, habitat or
// snack, a duplicate id, and an unknown rarity or time of day.
func LoadSpecies(data []byte) ([]Species, error) {
	var species []Species
	if err := json.Unmarshal(data, &species); err != nil {
		return nil, fmt.Errorf("parse critters: %w", err)
	}
	seen := make(map[string]bool, len(species))
	for i := range species {
		sp := &species[i]
		if sp.ID == "" {
			return nil, fmt.Errorf("critter %d: missing id", i)
		}
		if seen[sp.ID] {
			return nil, fmt.Errorf("critter %q: duplicate id", sp.ID)
		}
		seen[sp.ID] = true
		if err := sp.validate(); err != nil {
			return nil, fmt.Errorf("critter %q: %w", sp.ID, err)
		}
	}
	return species, nil
}

func (sp *Species) validate() error {
	if sp.Habitat == "" {
		return errors.New("missing habitat")
	}
	if sp.Likes == "" {
		return errors.New("missing likes")
	}
	switch sp.Rarity {
	case Common, Uncommon, Rare:
	default:
		return fmt.Errorf("unknown rarity %q", sp.Rarity)
	}
	for _, t := range sp.ActiveAt {
		switch t {
		case Day, Night:
		default:
			return fmt.Errorf("unknown activeAt %q", t)
		}
	}
	return nil
}
