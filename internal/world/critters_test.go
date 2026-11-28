package world

import (
	"errors"
	"math/rand/v2"
	"testing"

	"github.com/Elpulgo/critterkingdom/internal/critter"
	"github.com/Elpulgo/critterkingdom/internal/geom"
)

// The wilds are 10x6 tiles, 160x96 world pixels, with the start spawn at
// (120,88) and a warp over the tile (144,80)-(160,96) east of it. The pond
// zone "shore" covers (16,16)-(80,80) around solid water at (32,32)-(64,64)
// and holds the default 3 critters; the meadow zone "grass" covers
// (96,16)-(144,48) and holds 2.
//
// The garden is 6x4 tiles with its spawn from_wilds at (24,40) and one meadow
// zone "lawn" over (48,16)-(96,48) that holds 1 critter. The yard is the same
// size with the same spawn and no zones.
const (
	wildsObjects = `
		{"id": 1, "name": "start", "type": "player_spawn", "x": 960, "y": 704, "point": true},
		{"id": 2, "name": "shore", "type": "spawn_zone", "x": 128, "y": 128, "width": 512, "height": 512,
			"properties": [{"name": "habitat", "type": "string", "value": "pond"}]},
		{"id": 3, "name": "grass", "type": "spawn_zone", "x": 768, "y": 128, "width": 384, "height": 256,
			"properties": [{"name": "habitat", "type": "string", "value": "meadow"},
				{"name": "max", "type": "int", "value": 2}]}`
	wildsWater   = `{"id": 5, "x": 256, "y": 256, "width": 256, "height": 256}`
	gardenSpawn  = `{"id": 1, "name": "from_wilds", "type": "player_spawn", "x": 192, "y": 320, "point": true}`
	gardenObject = gardenSpawn + `,
		{"id": 2, "name": "lawn", "type": "spawn_zone", "x": 384, "y": 128, "width": 384, "height": 256,
			"properties": [{"name": "habitat", "type": "string", "value": "meadow"},
				{"name": "max", "type": "int", "value": 1}]}`
)

var (
	shoreRect = geom.Rect{Min: geom.Vec{X: 16, Y: 16}, Max: geom.Vec{X: 80, Y: 80}}
	grassRect = geom.Rect{Min: geom.Vec{X: 96, Y: 16}, Max: geom.Vec{X: 144, Y: 48}}
	lawnRect  = geom.Rect{Min: geom.Vec{X: 48, Y: 16}, Max: geom.Vec{X: 96, Y: 48}}
)

// wildsLoader loads the inline wilds, garden and yard maps through the real
// LoadMap; the wilds' warp leads to the from_wilds spawn of warpTo.
func wildsLoader(warpTo string) MapLoader {
	maps := map[string][]byte{
		"wilds":  roomMap(10, 6, "wilds", wildsObjects+`,`+warpObject(4, 9, 5, warpTo, "from_wilds"), wildsWater),
		"garden": roomMap(6, 4, "garden", gardenObject, ``),
		"yard":   roomMap(6, 4, "yard", gardenSpawn, ``),
	}
	return func(name string) (*Map, error) {
		data, ok := maps[name]
		if !ok {
			return nil, errors.New("not found")
		}
		return LoadMap(data)
	}
}

// species returns a valid species of habitat with the given rarity, about at
// the given times of day (any time when none are given).
func species(id, habitat string, rarity critter.Rarity, at ...critter.TimeOfDay) critter.Species {
	return critter.Species{ID: id, Habitat: habitat, Likes: "berry", Rarity: rarity, ActiveAt: at}
}

// newWildsScene starts a scene on the wilds with wild critters of sp, seeded
// with seed, and updates it once standing still.
func newWildsScene(t *testing.T, warpTo string, seed uint64, sp ...critter.Species) *Scene {
	t.Helper()
	s, err := NewScene(nil, wildsLoader(warpTo), "wilds")
	if err != nil {
		t.Fatalf("NewScene: %v", err)
	}
	s.AddWildCritters(sp, rand.New(rand.NewPCG(seed, 0)))
	standStill(t, s, 1)
	return s
}

// standStill updates s ticks times with the player standing still.
func standStill(t *testing.T, s *Scene, ticks int) {
	t.Helper()
	for range ticks {
		if err := s.Update(walkFrame(geom.Vec{})); err != nil {
			t.Fatalf("Update: %v", err)
		}
	}
}

// inWildsWater reports whether box overlaps the wilds' water at
// (32,32)-(64,64) with positive area.
func inWildsWater(box geom.Rect) bool {
	return box.Max.X > 32 && box.Min.X < 64 && box.Max.Y > 32 && box.Min.Y < 64
}

func TestScene_WildCrittersFillOnlyZonesWhoseHabitatHasASpecies_InsideTheZoneAndOffWater(t *testing.T) {
	for seed := range uint64(10) {
		s := newWildsScene(t, "garden", seed, species("frog", "pond", critter.Common))

		if len(s.wild) != 3 {
			t.Fatalf("seed %d: %d wild critters; want the shore's max of 3 and none on the grass, which no species lives in", seed, len(s.wild))
		}
		for _, w := range s.wild {
			if w.zone != "shore" || w.species.ID != "frog" {
				t.Errorf("seed %d: critter %s in zone %q; want frogs in the shore only", seed, w.species.ID, w.zone)
			}
			if !shoreRect.Contains(w.Pos) {
				t.Errorf("seed %d: critter at %v; want inside the shore %v", seed, w.Pos, shoreRect)
			}
			if inWildsWater(w.Box()) {
				t.Errorf("seed %d: critter feet box %v overlaps the water (32,32)-(64,64)", seed, w.Box())
			}
		}
	}
}

func TestScene_WildCrittersHopAboutAndNoticeAPlayerWhoComesClose(t *testing.T) {
	s := newWildsScene(t, "garden", 3, species("frog", "pond", critter.Common))
	start := make([]geom.Vec, len(s.wild))
	for i, w := range s.wild {
		start[i] = w.Pos
	}

	standStill(t, s, 180) // 3 s, longer than the longest rest

	moved := false
	for i, w := range s.wild {
		moved = moved || w.Pos != start[i]
		if w.Noticing() {
			t.Errorf("critter at %v noticing the player at %v, over 32 px away", w.Pos, s.player.Pos)
		}
	}
	if !moved {
		t.Fatalf("no wild critter moved in 3 s; still at %v", start)
	}

	near := s.wild[0]
	s.player.Place(near.Pos.Add(geom.Vec{X: 8}))
	standStill(t, s, 30) // 0.5 s, long enough to land a hop
	if !near.Noticing() {
		t.Errorf("critter at %v with the player at %v is not noticing it", near.Pos, s.player.Pos)
	}
}

func TestScene_NightOnlySpeciesStaysAwayInTheDay(t *testing.T) {
	s := newWildsScene(t, "garden", 1,
		species("frog", "pond", critter.Common, critter.Day),
		species("firefly", "pond", critter.Common, critter.Night))

	if len(s.wild) != 3 {
		t.Fatalf("%d wild critters in the day; want the shore's 3", len(s.wild))
	}
	for _, w := range s.wild {
		if w.species.ID != "frog" {
			t.Errorf("a %s is out in the day; want only frogs", w.species.ID)
		}
	}

	s.timeOfDay = critter.Night
	standStill(t, s, 1)

	if len(s.wild) != 3 {
		t.Fatalf("%d wild critters at night; want the shore's 3", len(s.wild))
	}
	for _, w := range s.wild {
		if w.species.ID != "firefly" {
			t.Errorf("a %s is out at night; want only fireflies", w.species.ID)
		}
	}
}

func TestScene_WarpingReplacesTheWildCrittersWithTheTargetMapsOwn(t *testing.T) {
	meadowOnly := species("snail", "meadow", critter.Common)

	t.Run("map with a zone", func(t *testing.T) {
		s := newWildsScene(t, "garden", 1, meadowOnly)
		if len(s.wild) != 2 {
			t.Fatalf("%d wild critters in the wilds; want the grass's 2", len(s.wild))
		}
		for _, w := range s.wild {
			if !grassRect.Contains(w.Pos) {
				t.Fatalf("critter at %v in the wilds; want in the grass %v", w.Pos, grassRect)
			}
		}

		updateUntilMap(t, s, "garden", walkFrame(geom.Vec{X: 1}))

		if len(s.wild) != 1 {
			t.Fatalf("%d wild critters in the garden; want the lawn's 1", len(s.wild))
		}
		if w := s.wild[0]; w.zone != "lawn" || !lawnRect.Contains(w.Pos) {
			t.Errorf("critter in zone %q at %v; want in the lawn %v", w.zone, w.Pos, lawnRect)
		}
	})

	t.Run("map without zones", func(t *testing.T) {
		s := newWildsScene(t, "yard", 1, meadowOnly)

		updateUntilMap(t, s, "yard", walkFrame(geom.Vec{X: 1}))

		if len(s.wild) != 0 {
			t.Errorf("%d wild critters in the yard, which has no zones; want none", len(s.wild))
		}
	})
}

func TestScene_RareSpeciesCrittersCarryTheRareFlag(t *testing.T) {
	s := newWildsScene(t, "garden", 1,
		species("lilyfin", "pond", critter.Rare),
		species("snail", "meadow", critter.Common))

	if len(s.wild) != 5 {
		t.Fatalf("%d wild critters; want 3 on the shore and 2 on the grass", len(s.wild))
	}
	for _, w := range s.wild {
		wantRare := w.species.ID == "lilyfin"
		if w.Rare != wantRare {
			t.Errorf("%s critter has Rare %v; want %v", w.species.ID, w.Rare, wantRare)
		}
	}
}
