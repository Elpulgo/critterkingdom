package assets_test

import (
	"reflect"
	"testing"

	"github.com/Elpulgo/critterkingdom/internal/assets"
	"github.com/Elpulgo/critterkingdom/internal/critter"
)

func TestLoadCritters_LoadsTheShippedSpeciesFromTheEmbeddedFile(t *testing.T) {
	species, err := assets.LoadCritters()
	if err != nil {
		t.Fatalf("LoadCritters: %v", err)
	}

	want := []critter.Species{
		{
			ID:       "firefly",
			Name:     critter.Text{"en": "Firefly", "sv": "Eldfluga"},
			Habitat:  "pond",
			ActiveAt: []critter.TimeOfDay{critter.Night},
			Likes:    "berry",
			Rarity:   critter.Uncommon,
			Sprite:   "critters/firefly",
			Ability:  "light",
			Fact:     critter.Text{"en": "Fireflies glow to say hello.", "sv": "Eldflugor lyser för att säga hej."},
		},
		{
			ID:      "frog",
			Name:    critter.Text{"en": "Frog", "sv": "Groda"},
			Habitat: "pond",
			Likes:   "cookie",
			Rarity:  critter.Common,
			Sprite:  "critters/frog",
			Ability: "fetch",
			Fact:    critter.Text{"en": "Frogs can jump many times their own length.", "sv": "Grodor kan hoppa många gånger sin egen längd."},
		},
		{
			ID:      "snail",
			Name:    critter.Text{"en": "Snail", "sv": "Snigel"},
			Habitat: "meadow",
			Likes:   "berry",
			Rarity:  critter.Common,
			Sprite:  "critters/snail",
			Fact:    critter.Text{"en": "Snails carry their house on their back.", "sv": "Sniglar bär sitt hus på ryggen."},
		},
		{
			ID:      "bird",
			Name:    critter.Text{"en": "Bird", "sv": "Fågel"},
			Habitat: "meadow",
			Likes:   "berry",
			Rarity:  critter.Uncommon,
			Sprite:  "critters/bird",
			Ability: "point",
			Fact:    critter.Text{"en": "Birds sing to talk to each other.", "sv": "Fåglar sjunger för att prata med varandra."},
		},
		{
			ID:      "squirrel",
			Name:    critter.Text{"en": "Squirrel", "sv": "Ekorre"},
			Habitat: "meadow",
			Likes:   "cookie",
			Rarity:  critter.Rare,
			Sprite:  "critters/squirrel",
			Fact:    critter.Text{"en": "Squirrels hide nuts to eat in winter.", "sv": "Ekorrar gömmer nötter att äta på vintern."},
		},
	}
	if !reflect.DeepEqual(species, want) {
		t.Errorf("LoadCritters() =\n%+v\nwant\n%+v", species, want)
	}
}
