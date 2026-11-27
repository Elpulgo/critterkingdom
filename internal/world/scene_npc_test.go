package world

import (
	"errors"
	"fmt"
	"slices"
	"testing"

	"github.com/Elpulgo/critterkingdom/internal/engine"
	"github.com/Elpulgo/critterkingdom/internal/entity"
	"github.com/Elpulgo/critterkingdom/internal/geom"
)

// pointNPC returns a point npc object called id at the Tiled pixel x, y, a
// villager that stands still there.
func pointNPC(objectID int, id string, x, y int) string {
	return fmt.Sprintf(`{"id": %d, "type": "npc", "x": %d, "y": %d, "point": true,
		"properties": [{"name": "id", "type": "string", "value": %q}]}`, objectID, x, y, id)
}

// mapLoader loads the named inline maps through the real LoadMap.
func mapLoader(maps map[string][]byte) MapLoader {
	return func(name string) (*Map, error) {
		data, ok := maps[name]
		if !ok {
			return nil, errors.New("not found")
		}
		return LoadMap(data)
	}
}

// villager is what the tests check of a villager in the scene.
type villager struct {
	ID  string
	Pos geom.Vec
}

// villagersOf returns the id and feet of every villager in s, in order.
func villagersOf(s *Scene) []villager {
	var got []villager
	for _, n := range s.npcs {
		got = append(got, villager{n.ID, n.Pos})
	}
	return got
}

func TestScene_VillagersStartAtTheirSpawnsAndAWarpSwapsThemForTheTargetMapsVillagers(t *testing.T) {
	// The village of sceneLoader, with the baker wandering the rectangle
	// world (0,0)..(32,16) from its centre (16,8) and grandma standing at
	// Tiled (640,128), world (80,16). The house has the owl standing at
	// Tiled (64,448), world (8,56).
	baker := `{"id": 3, "type": "npc", "x": 0, "y": 0, "width": 256, "height": 128,
		"properties": [{"name": "id", "type": "string", "value": "baker_badger"}]}`
	load := mapLoader(map[string][]byte{
		"village": roomMap(6, 4, "village",
			villageStart+`,`+warpObject(2, 4, 2, "house", "from_village")+`,`+baker+`,`+pointNPC(4, "grandma_hedgehog", 640, 128), ``),
		"house": roomMap(5, 4, "house", houseObjects+`,`+pointNPC(4, "mayor_owl", 64, 448), houseWall),
	})
	s := newVillageScene(t, load)

	want := []villager{
		{"baker_badger", geom.Vec{X: 16, Y: 8}},
		{"grandma_hedgehog", geom.Vec{X: 80, Y: 16}},
	}
	if got := villagersOf(s); !slices.Equal(got, want) {
		t.Fatalf("villagers in the village = %+v; want %+v", got, want)
	}

	updateUntilMap(t, s, "house", walkFrame(geom.Vec{X: 1}))

	want = []villager{{"mayor_owl", geom.Vec{X: 8, Y: 56}}}
	if got := villagersOf(s); !slices.Equal(got, want) {
		t.Errorf("villagers after warping into the house = %+v; want only %+v", got, want)
	}
}

func TestScene_AVillagerNextToTheCritterTurnsToFaceIt(t *testing.T) {
	// The critter starts at (24,40). The baker stands 24 px east of it at
	// Tiled (384,320), world (48,40); grandma stands 20 px south of it at
	// Tiled (192,480), world (24,60). Both start facing down.
	load := mapLoader(map[string][]byte{
		"village": roomMap(6, 4, "village",
			villageStart+`,`+pointNPC(2, "baker_badger", 384, 320)+`,`+pointNPC(3, "grandma_hedgehog", 192, 480), ``),
	})
	s := newVillageScene(t, load)

	if err := s.Update(engine.Frame{View: villageView, DT: 1.0 / 60}); err != nil {
		t.Fatalf("Update: %v", err)
	}

	want := map[string]entity.Facing{
		"baker_badger":     entity.FacingLeft,
		"grandma_hedgehog": entity.FacingUp,
	}
	if len(s.npcs) != len(want) {
		t.Fatalf("scene has %d villagers; want %d", len(s.npcs), len(want))
	}
	for _, n := range s.npcs {
		if n.Facing != want[n.ID] {
			t.Errorf("%s faces %s; want %s, towards the critter", n.ID, n.Facing, want[n.ID])
		}
	}
}

// pawFrame is one tick at 60 ticks per second with the paw button pressed.
var pawFrame = engine.Frame{Input: engine.Input{Paw: true}, View: villageView, DT: 1.0 / 60}

func TestScene_PawInFrontOfAVillagerStartsAnInteractionInsteadOfAHop(t *testing.T) {
	// The critter starts at (24,40) facing down. Grandma stands 20 px in
	// front of it at Tiled (192,480), world (24,60); the baker stands nearer,
	// 16 px behind it, at Tiled (192,192), world (24,24).
	load := mapLoader(map[string][]byte{
		"village": roomMap(6, 4, "village",
			villageStart+`,`+pointNPC(2, "baker_badger", 192, 192)+`,`+pointNPC(3, "grandma_hedgehog", 192, 480), ``),
	})
	s := newVillageScene(t, load)

	if err := s.Update(pawFrame); err != nil {
		t.Fatalf("Update: %v", err)
	}

	got, ok := s.TakeInteraction()
	if want := (Interaction{NPC: "grandma_hedgehog"}); !ok || got != want {
		t.Fatalf("TakeInteraction() = %+v, %v; want %+v, true", got, ok, want)
	}
	if got, ok := s.TakeInteraction(); ok {
		t.Errorf("second TakeInteraction() = %+v, true; want the interaction taken only once", got)
	}
	if h := s.player.HopHeight(); h != 0 || s.greet != 0 {
		t.Errorf("hop height %v, greeting %v s left; want no hop and no greeting when paw talks to a villager", h, s.greet)
	}
}

func TestScene_PawWithNoVillagerInFrontHopsAndStartsNoInteraction(t *testing.T) {
	// The critter starts at (24,40) facing down, on a village 80 px high.
	tests := []struct {
		name     string
		villager string
	}{
		{"no villager on the map", ``},
		// At Tiled (192,192), world (24,24): 16 px behind the critter.
		{"villager behind", `,` + pointNPC(2, "baker_badger", 192, 192)},
		// At Tiled (192,520), world (24,65): 25 px in front, beyond reach.
		{"villager out of reach", `,` + pointNPC(2, "baker_badger", 192, 520)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			load := mapLoader(map[string][]byte{
				"village": roomMap(6, 5, "village", villageStart+tt.villager, ``),
			})
			s := newVillageScene(t, load)

			if err := s.Update(pawFrame); err != nil {
				t.Fatalf("Update: %v", err)
			}

			if got, ok := s.TakeInteraction(); ok {
				t.Errorf("TakeInteraction() = %+v, true; want no interaction", got)
			}
			if h := s.player.HopHeight(); h <= 0 || s.greet <= 0 {
				t.Errorf("hop height %v, greeting %v s left; want the critter to hop and greet", h, s.greet)
			}
		})
	}
}
