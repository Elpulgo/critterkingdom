package assets_test

import (
	"reflect"
	"testing"

	"github.com/Elpulgo/critterkingdom/internal/assets"
	"github.com/Elpulgo/critterkingdom/internal/critter"
)

func TestLoadCritters_LoadsTheFireflyFromTheEmbeddedFile(t *testing.T) {
	species, err := assets.LoadCritters()
	if err != nil {
		t.Fatalf("LoadCritters: %v", err)
	}

	want := []critter.Species{{
		ID:       "firefly",
		Name:     critter.Text{"en": "Firefly", "sv": "Eldfluga"},
		Habitat:  "pond",
		ActiveAt: []critter.TimeOfDay{critter.Night},
		Likes:    "berry",
		Rarity:   critter.Uncommon,
		Sprite:   "critters/firefly",
		Ability:  "light",
		Fact:     critter.Text{"en": "Fireflies glow to say hello.", "sv": "Eldflugor lyser för att säga hej."},
	}}
	if !reflect.DeepEqual(species, want) {
		t.Errorf("LoadCritters() =\n%+v\nwant\n%+v", species, want)
	}
}
