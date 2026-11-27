package world

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/Elpulgo/critterkingdom/internal/engine"
	"github.com/Elpulgo/critterkingdom/internal/geom"
)

// roomMap returns a cols by rows Tiled map whose biome property is biome, with
// the given objects and collision object layers, both as JSON array bodies in
// Tiled pixels (128 to a tile, 8 to a world pixel).
func roomMap(cols, rows int, biome, objects, walls string) []byte {
	return fmt.Appendf(nil, `{
		"orientation": "orthogonal", "infinite": false,
		"width": %d, "height": %d, "tilewidth": 128, "tileheight": 128,
		"properties": [{"name": "biome", "type": "string", "value": %q}],
		"layers": [
			{"name": "collision", "type": "objectgroup", "objects": [%s]},
			{"name": "objects", "type": "objectgroup", "objects": [%s]}
		]
	}`, cols, rows, biome, walls, objects)
}

// warpObject returns a warp over the tile at col, row to targetMap's spawn
// targetSpawn.
func warpObject(id, col, row int, targetMap, targetSpawn string) string {
	return fmt.Sprintf(`{"id": %d, "type": "warp",
		"x": %d, "y": %d, "width": 128, "height": 128,
		"properties": [
			{"name": "target_map", "type": "string", "value": %q},
			{"name": "target_spawn", "type": "string", "value": %q}]}`,
		id, col*128, row*128, targetMap, targetSpawn)
}

// The village is 6x4 tiles, 96x64 world pixels, with its start spawn at
// (24,40) and a warp over the tile (64,32)-(80,48) east of it, so walking east
// from the start enters the warp.
//
// The house is 5x4 tiles, 80x64 world pixels, with spawns by_the_bed (first)
// at (56,24) and from_village at (24,28), and one solid tile at (64,32)-(80,48).
const (
	villageStart = `{"id": 1, "name": "start", "type": "player_spawn", "x": 192, "y": 320, "point": true}`
	houseObjects = `
		{"id": 1, "name": "by_the_bed", "type": "player_spawn", "x": 448, "y": 192, "point": true},
		{"id": 2, "name": "from_village", "type": "player_spawn", "x": 192, "y": 224, "point": true}`
	houseWall = `{"id": 3, "x": 512, "y": 256, "width": 128, "height": 128}`
)

// sceneLoader loads the inline village and house maps through the real
// LoadMap. The village's warp leads to targetMap's spawn targetSpawn. Any
// other name fails with an error that does not mention the name.
func sceneLoader(targetMap, targetSpawn string) MapLoader {
	maps := map[string][]byte{
		"village": roomMap(6, 4, "village", villageStart+`,`+warpObject(2, 4, 2, targetMap, targetSpawn), ``),
		"house":   roomMap(5, 4, "house", houseObjects, houseWall),
		"empty":   roomMap(5, 4, "empty", ``, ``),
	}
	return func(name string) (*Map, error) {
		data, ok := maps[name]
		if !ok {
			return nil, errors.New("not found")
		}
		return LoadMap(data)
	}
}

// villageView fits the whole village, so the camera starts and stays at 0,0
// there and a pointer's view position is its world position.
var villageView = engine.View{Scale: 1, W: 96, H: 64}

// walkFrame is one tick at 60 ticks per second, walking dir.
func walkFrame(dir geom.Vec) engine.Frame {
	return engine.Frame{Input: engine.Input{Move: dir}, View: villageView, DT: 1.0 / 60}
}

// pointFrame is one tick at 60 ticks per second with the pointer held at p
// in view pixels, or let go when held is false.
func pointFrame(held bool, p geom.Vec) engine.Frame {
	return engine.Frame{Input: engine.Input{Pointing: held, Point: p}, View: villageView, DT: 1.0 / 60}
}

// newVillageScene starts a scene on the village of load.
func newVillageScene(t *testing.T, load MapLoader) *Scene {
	t.Helper()
	s, err := NewScene(nil, load, "village")
	if err != nil {
		t.Fatalf("NewScene: %v", err)
	}
	return s
}

// updateUntilMap updates s with f until its map is called name, failing after
// a hundred ticks.
func updateUntilMap(t *testing.T, s *Scene, name string, f engine.Frame) {
	t.Helper()
	for range 100 {
		if err := s.Update(f); err != nil {
			t.Fatalf("Update: %v", err)
		}
		if s.mapName == name {
			return
		}
	}
	t.Fatalf("still on map %q at %v after 100 ticks; want to have warped to %q", s.mapName, s.player.Pos, name)
}

func TestScene_StartsOnTheFirstSpawnOfTheStartMap(t *testing.T) {
	s, err := NewScene(nil, sceneLoader("house", "from_village"), "house")
	if err != nil {
		t.Fatalf("NewScene: %v", err)
	}
	if s.player.Pos != (geom.Vec{X: 56, Y: 24}) {
		t.Errorf("critter at %v; want on by_the_bed, the house's first spawn, at 56,24", s.player.Pos)
	}
}

func TestScene_StartingOnAMissingMapOrOneWithoutSpawnIsAnError(t *testing.T) {
	for _, name := range []string{"cave", "empty"} {
		t.Run(name, func(t *testing.T) {
			s, err := NewScene(nil, sceneLoader("house", "from_village"), name)
			if err == nil {
				t.Fatalf("NewScene(%q) = %v; want an error", name, s)
			}
			if !strings.Contains(err.Error(), `"`+name+`"`) {
				t.Errorf("NewScene(%q) error = %q; want it to name the map", name, err)
			}
		})
	}
}

func TestScene_WalkingIntoAWarpLoadsTheTargetMapAtTheTargetSpawn(t *testing.T) {
	s := newVillageScene(t, sceneLoader("house", "from_village"))
	east := geom.Vec{X: 1}

	updateUntilMap(t, s, "house", walkFrame(east))

	if s.m.Biome != "house" {
		t.Errorf("map biome = %q; want the house map", s.m.Biome)
	}
	if s.player.Pos != (geom.Vec{X: 24, Y: 28}) {
		t.Errorf("critter at %v; want exactly on the house's from_village spawn, 24,28", s.player.Pos)
	}
	if target, walking := s.player.Target(); walking {
		t.Errorf("critter walking to %v; want no walk target after the warp", target)
	}

	// The house is 80 wide, so walking east stops 6 short of its edge, where
	// the village would have let the critter go on to 90.
	for range 60 {
		if err := s.Update(walkFrame(east)); err != nil {
			t.Fatalf("Update: %v", err)
		}
	}
	if s.player.Pos != (geom.Vec{X: 74, Y: 28}) {
		t.Fatalf("after walking east in the house: at %v; want stopped at the house's edge, 74,28", s.player.Pos)
	}
	// Walking south runs into the house's solid tile, whose top is at 32; the
	// village has no wall there.
	for range 60 {
		if err := s.Update(walkFrame(geom.Vec{Y: 1})); err != nil {
			t.Fatalf("Update: %v", err)
		}
	}
	if s.player.Pos != (geom.Vec{X: 74, Y: 32}) {
		t.Errorf("after walking south in the house: at %v; want stopped on the house's wall, 74,32", s.player.Pos)
	}
}

func TestScene_WalkingIntoAWarpToAMissingMapOrSpawnFailsNamingIt(t *testing.T) {
	tests := []struct {
		name        string
		targetMap   string
		targetSpawn string
		wantErrs    []string
	}{
		{"missing map", "cave", "from_village", []string{`"cave"`}},
		{"missing spawn", "house", "from_cellar", []string{`"house"`, `"from_cellar"`}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := newVillageScene(t, sceneLoader(tt.targetMap, tt.targetSpawn))
			var err error
			for range 100 {
				if err = s.Update(walkFrame(geom.Vec{X: 1})); err != nil {
					break
				}
			}
			if err == nil {
				t.Fatalf("Update never failed; critter at %v on %q", s.player.Pos, s.mapName)
			}
			for _, want := range tt.wantErrs {
				if !strings.Contains(err.Error(), want) {
					t.Errorf("Update error = %q; want it to contain %s", err, want)
				}
			}
		})
	}
}

func TestScene_APointerHeldAcrossAWarpDoesNotWalkUntilReleased(t *testing.T) {
	s := newVillageScene(t, sceneLoader("house", "from_village"))
	inWarp := geom.Vec{X: 72, Y: 40}
	spawn := geom.Vec{X: 24, Y: 28}

	updateUntilMap(t, s, "house", pointFrame(true, inWarp))
	for range 30 {
		if err := s.Update(pointFrame(true, inWarp)); err != nil {
			t.Fatalf("Update: %v", err)
		}
	}
	if s.player.Pos != spawn {
		t.Fatalf("with the pointer still held: at %v; want still on the spawn %v", s.player.Pos, spawn)
	}

	if err := s.Update(pointFrame(false, geom.Vec{})); err != nil {
		t.Fatalf("Update: %v", err)
	}
	for range 10 {
		if err := s.Update(pointFrame(true, inWarp)); err != nil {
			t.Fatalf("Update: %v", err)
		}
	}
	if s.player.Pos == spawn {
		t.Errorf("after letting go and pressing again: still on the spawn %v; want the critter to walk", spawn)
	}
}
