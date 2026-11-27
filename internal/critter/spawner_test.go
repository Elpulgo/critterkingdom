package critter_test

import (
	"math/rand/v2"
	"slices"
	"testing"
	"time"

	"github.com/Elpulgo/critterkingdom/internal/critter"
)

// gameClock is a game clock the test moves by hand.
type gameClock struct{ now time.Duration }

func (c *gameClock) read() time.Duration { return c.now }

func newSpawner(species []critter.Species, zones []critter.Zone, clock *gameClock) *critter.Spawner {
	return critter.NewSpawner(species, zones, rand.New(rand.NewPCG(1, 2)), clock.read)
}

func presentIDs(present []critter.Spawned) []string {
	var out []string
	for _, c := range present {
		out = append(out, c.Species.ID)
	}
	return out
}

func TestSpawner_DrawsSpeciesInProportionToRarity(t *testing.T) {
	pond := []critter.Species{
		{ID: "frog", Habitat: "pond", Rarity: critter.Common},
		{ID: "newt", Habitat: "pond", Rarity: critter.Uncommon},
		{ID: "axolotl", Habitat: "pond", Rarity: critter.Rare},
	}
	zones := []critter.Zone{{ID: "pond-1", Habitat: "pond", Max: 10000}}

	run := func() []string {
		s := newSpawner(pond, zones, &gameClock{})
		s.Update(critter.World{TimeOfDay: critter.Day})
		return presentIDs(s.Present("pond-1"))
	}
	first := run()

	counts := map[string]int{}
	for _, id := range first {
		counts[id]++
	}
	// Weights 6:3:1 over 10000 draws expect 6000, 3000 and 1000; allow 200
	// either way (more than four standard deviations for each).
	want := map[string]int{"frog": 6000, "newt": 3000, "axolotl": 1000}
	const tolerance = 200
	for id, n := range want {
		if counts[id] < n-tolerance || counts[id] > n+tolerance {
			t.Errorf("%s drawn %d times, want %d ± %d", id, counts[id], n, tolerance)
		}
	}

	if again := run(); !slices.Equal(first, again) {
		t.Errorf("a second run with the same seed drew a different sequence")
	}
}

func TestSpawner_SpawnsOnlyActiveCrittersOfTheZoneHabitat(t *testing.T) {
	species := []critter.Species{
		{ID: "firefly", Habitat: "pond", Rarity: critter.Uncommon, ActiveAt: []critter.TimeOfDay{critter.Night}},
		{ID: "frog", Habitat: "pond", Rarity: critter.Common},
		{ID: "heron", Habitat: "pond", Rarity: critter.Common, NeedsFlag: "bridge_fixed"},
		{ID: "squirrel", Habitat: "forest", Rarity: critter.Common},
	}
	zones := []critter.Zone{
		{ID: "pond-1", Habitat: "pond", Max: 4},
		{ID: "wood-1", Habitat: "forest", Max: 2},
	}
	s := newSpawner(species, zones, &gameClock{})

	s.Update(critter.World{TimeOfDay: critter.Day})

	if got, want := presentIDs(s.Present("pond-1")), []string{"frog", "frog", "frog", "frog"}; !slices.Equal(got, want) {
		t.Errorf("pond by day = %v, want %v", got, want)
	}
	if got, want := presentIDs(s.Present("wood-1")), []string{"squirrel", "squirrel"}; !slices.Equal(got, want) {
		t.Errorf("forest = %v, want %v", got, want)
	}
}

func TestSpawner_ZoneWithNoActiveSpeciesStaysShortUntilOneBecomesActive(t *testing.T) {
	species := []critter.Species{
		{ID: "firefly", Habitat: "pond", Rarity: critter.Uncommon, ActiveAt: []critter.TimeOfDay{critter.Night}},
	}
	zones := []critter.Zone{{ID: "pond-1", Habitat: "pond", Max: 3, RespawnAfter: time.Hour}}
	s := newSpawner(species, zones, &gameClock{})

	s.Update(critter.World{TimeOfDay: critter.Day})
	if got := len(s.Present("pond-1")); got != 0 {
		t.Fatalf("pond by day holds %d critters, want 0", got)
	}

	s.Update(critter.World{TimeOfDay: critter.Night})
	if got, want := presentIDs(s.Present("pond-1")), []string{"firefly", "firefly", "firefly"}; !slices.Equal(got, want) {
		t.Errorf("pond at night = %v, want %v", got, want)
	}
}

func TestSpawner_RemovedCritterIsReplacedOnlyAfterTheRespawnTime(t *testing.T) {
	species := []critter.Species{{ID: "frog", Habitat: "pond", Rarity: critter.Common}}
	zones := []critter.Zone{{ID: "pond-1", Habitat: "pond", Max: 3, RespawnAfter: 30 * time.Second}}
	clock := &gameClock{now: 10 * time.Second}
	s := newSpawner(species, zones, clock)
	day := critter.World{TimeOfDay: critter.Day}

	s.Update(day)
	s.Update(day)
	if got := len(s.Present("pond-1")); got != 3 {
		t.Fatalf("after two updates the zone holds %d, want Max 3", got)
	}

	s.Remove(s.Present("pond-1")[0].ID)
	for _, at := range []time.Duration{10 * time.Second, 39 * time.Second} {
		clock.now = at
		s.Update(day)
		if got := len(s.Present("pond-1")); got != 2 {
			t.Fatalf("at %v the zone holds %d, want 2 while the slot waits", at, got)
		}
	}

	clock.now = 40 * time.Second
	s.Update(day)
	if got := len(s.Present("pond-1")); got != 3 {
		t.Errorf("at 40s, 30s after the removal, the zone holds %d, want 3", got)
	}
}

func TestSpawner_NightCritterLeavesWhenDayComesAndItsSlotRefillsAtOnce(t *testing.T) {
	species := []critter.Species{
		{ID: "firefly", Habitat: "pond", Rarity: critter.Uncommon, ActiveAt: []critter.TimeOfDay{critter.Night}},
		{ID: "dragonfly", Habitat: "pond", Rarity: critter.Common, ActiveAt: []critter.TimeOfDay{critter.Day}},
	}
	zones := []critter.Zone{{ID: "pond-1", Habitat: "pond", Max: 2, RespawnAfter: time.Hour}}
	s := newSpawner(species, zones, &gameClock{})

	s.Update(critter.World{TimeOfDay: critter.Night})
	if got, want := presentIDs(s.Present("pond-1")), []string{"firefly", "firefly"}; !slices.Equal(got, want) {
		t.Fatalf("pond at night = %v, want %v", got, want)
	}

	s.Update(critter.World{TimeOfDay: critter.Day})
	if got, want := presentIDs(s.Present("pond-1")), []string{"dragonfly", "dragonfly"}; !slices.Equal(got, want) {
		t.Errorf("pond when day comes = %v, want %v", got, want)
	}
}

func TestSpawner_FlagsExactlyTheRareCritters(t *testing.T) {
	species := []critter.Species{
		{ID: "frog", Habitat: "pond", Rarity: critter.Common},
		{ID: "newt", Habitat: "pond", Rarity: critter.Uncommon},
		{ID: "axolotl", Habitat: "pond", Rarity: critter.Rare},
	}
	zones := []critter.Zone{{ID: "pond-1", Habitat: "pond", Max: 100}}
	s := newSpawner(species, zones, &gameClock{})

	s.Update(critter.World{TimeOfDay: critter.Day})

	rare := 0
	for _, c := range s.Present("pond-1") {
		if want := c.Species.ID == "axolotl"; c.Rare != want {
			t.Errorf("%s spawned with Rare = %v, want %v", c.Species.ID, c.Rare, want)
		}
		if c.Rare {
			rare++
		}
	}
	if rare == 0 {
		t.Errorf("no rare critter among 100 draws, so the flag was never checked")
	}
}
