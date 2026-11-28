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

// darkRoomMap returns a cols by rows Tiled map like roomMap, without walls,
// whose dark property is dark.
func darkRoomMap(cols, rows int, biome string, dark bool, objects string) []byte {
	return fmt.Appendf(nil, `{
		"orientation": "orthogonal", "infinite": false,
		"width": %d, "height": %d, "tilewidth": 128, "tileheight": 128,
		"properties": [
			{"name": "biome", "type": "string", "value": %q},
			{"name": "dark", "type": "bool", "value": %t}],
		"layers": [
			{"name": "collision", "type": "objectgroup", "objects": []},
			{"name": "objects", "type": "objectgroup", "objects": [%s]}
		]
	}`, cols, rows, biome, dark, objects)
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

// The cave is a dark 6x4-tile room, 96x64 world pixels, with its first spawn
// from_village at (24,40) and a warp over the tile (64,32)-(80,48) back to the
// village's start, so walking east from the spawn warps back. The meadow is
// the same room with dark set to false and no warp.
const caveObjects = `{"id": 1, "name": "from_village", "type": "player_spawn", "x": 192, "y": 320, "point": true}`

// sceneLoader loads the inline village, house, dark_cave and lit_meadow maps
// through the real LoadMap. The village's warp leads to targetMap's spawn targetSpawn. Any
// other name fails with an error that does not mention the name.
func sceneLoader(targetMap, targetSpawn string) MapLoader {
	maps := map[string][]byte{
		"village": roomMap(6, 4, "village", villageStart+`,`+warpObject(2, 4, 2, targetMap, targetSpawn), ``),
		"house":   roomMap(5, 4, "house", houseObjects, houseWall),
		"empty":   roomMap(5, 4, "empty", ``, ``),
		"dark_cave": darkRoomMap(6, 4, "cave", true,
			caveObjects+`,`+warpObject(2, 4, 2, "village", "start")),
		"lit_meadow": darkRoomMap(6, 4, "meadow", false, caveObjects),
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

// The garden is 6x4 tiles, 96x64 world pixels, like the village, so
// villageView fits it. A wall fills column 3 from the top down to row 2,
// (48,0)-(64,48), leaving a one-tile gap in the bottom row. The critter starts
// at (24,29), centred in cell (1,1) west of the wall.
const (
	gardenStart = `{"id": 1, "name": "start", "type": "player_spawn", "x": 192, "y": 232, "point": true}`
	gardenWall  = `{"id": 2, "x": 384, "y": 0, "width": 128, "height": 384}`
)

// newGardenScene starts a scene on the garden, loaded through LoadMap.
func newGardenScene(t *testing.T) *Scene {
	t.Helper()
	data := roomMap(6, 4, "garden", gardenStart, gardenWall)
	s, err := NewScene(nil, func(string) (*Map, error) { return LoadMap(data) }, "garden")
	if err != nil {
		t.Fatalf("NewScene: %v", err)
	}
	return s
}

// update runs s for ticks ticks of f.
func update(t *testing.T, s *Scene, f engine.Frame, ticks int) {
	t.Helper()
	for range ticks {
		if err := s.Update(f); err != nil {
			t.Fatalf("Update: %v", err)
		}
	}
}

// checkStoppedAt fails the test unless the critter's feet are within 0.5 px
// of want and it is no longer walking.
func checkStoppedAt(t *testing.T, s *Scene, want geom.Vec) {
	t.Helper()
	if d := s.player.Pos.Sub(want).Len(); d > 0.5 {
		t.Errorf("critter at %v, %.2f px from %v; want within 0.5 px", s.player.Pos, d, want)
	}
	if target, walking := s.player.Target(); walking {
		t.Errorf("critter still walking to %v; want it stopped", target)
	}
}

func TestScene_TappingAcrossAWallWalksAroundItToTheTappedSpot(t *testing.T) {
	s := newGardenScene(t)
	spot := geom.Vec{X: 80, Y: 24} // east of the wall, in cell (5,1)

	// The way round, down through the gap and back up, is about 100 px:
	// under 1.5 s at the walking speed. Give it 4 s.
	update(t, s, pointFrame(true, spot), 240)

	checkStoppedAt(t, s, spot)
}

func TestScene_TappingASolidTileWalksToTheNearestReachableTile(t *testing.T) {
	s := newGardenScene(t)
	// Start east of the wall in cell (5,1). The tapped wall cell (3,1) has
	// two open neighbours one cell away, (2,1) and (4,1); the tie goes to
	// the lower column, (2,1), behind the wall, so the critter must go
	// round through the gap. The box is centred in (2,1) with feet at 40,29.
	s.player.Place(geom.Vec{X: 88, Y: 29})
	onTheWall := geom.Vec{X: 56, Y: 24}

	update(t, s, pointFrame(true, onTheWall), 240)

	checkStoppedAt(t, s, geom.Vec{X: 40, Y: 29})
}

func TestScene_MovingAHeldPointerToAnotherTileReplansTheRoute(t *testing.T) {
	s := newGardenScene(t)
	// Head for the far side of the wall, then, while the critter is on its
	// way to the gap, lead it back to a spot on the start side. The old
	// route would take it on through the gap and round the wall again, well
	// over 100 px; the spot is under 40 px straight back, so one second of
	// walking (72 px) is enough only when the route is replanned.
	update(t, s, pointFrame(true, geom.Vec{X: 80, Y: 24}), 20)
	back := geom.Vec{X: 24, Y: 14}

	update(t, s, pointFrame(true, back), 60)

	checkStoppedAt(t, s, back)
}

func TestScene_ADirectionInputTakesOverFromARouteForGood(t *testing.T) {
	s := newGardenScene(t)
	update(t, s, pointFrame(true, geom.Vec{X: 80, Y: 24}), 10)
	before := s.player.Pos

	// Ten ticks walking west at 72 px/s is 12 px, from wherever the route
	// had got to; the critter is still west of the wall and clear of x=6.
	update(t, s, walkFrame(geom.Vec{X: -1}), 10)
	want := geom.Vec{X: before.X - 12, Y: before.Y}
	if d := s.player.Pos.Sub(want).Len(); d > 1e-9 {
		t.Fatalf("after walking west: at %v; want 12 px west of %v, at %v", s.player.Pos, before, want)
	}

	update(t, s, pointFrame(false, geom.Vec{}), 60)
	checkStoppedAt(t, s, want)
}

func TestScene_HoldingThePointerWhereTheCritterDoesNotFitStopsItAgainstTheWall(t *testing.T) {
	tests := []struct {
		name  string
		point geom.Vec
		want  geom.Vec
	}{
		// In the gap cell (3,3), under the wall's bottom at y=48: the 10 px
		// tall box fits only with its feet at y=58 or lower.
		{"under a wall", geom.Vec{X: 56, Y: 52}, geom.Vec{X: 56, Y: 58}},
		// In cell (2,1), beside the wall's west face at x=48: the 12 px wide
		// box fits only with its feet at x=42 or further west.
		{"beside a wall", geom.Vec{X: 46, Y: 29}, geom.Vec{X: 42, Y: 29}},
		// In cell (2,3), below-left of the wall's bottom corner at (48,48),
		// with open cells above and to the right: the box clears the wall by
		// moving 5 px west rather than 8 px down.
		{"by a wall's corner", geom.Vec{X: 47, Y: 50}, geom.Vec{X: 42, Y: 50}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := newGardenScene(t)
			held := pointFrame(true, tt.point)

			update(t, s, held, 240)
			checkStoppedAt(t, s, tt.want)

			// Still holding, the critter stays put rather than stepping
			// back and forth against the wall.
			stopped := s.player.Pos
			for i := range 30 {
				update(t, s, held, 1)
				if s.player.Pos != stopped {
					t.Fatalf("tick %d after stopping: critter at %v; want it to stay at %v", i+1, s.player.Pos, stopped)
				}
			}
		})
	}
}

// lights returns the lights of s, failing the test unless there are want of
// them.
func lights(t *testing.T, s *Scene, want int) []Light {
	t.Helper()
	got := s.Lights()
	if len(got) != want {
		t.Fatalf("on %q with the critter at %v: Lights() = %v; want %d lights", s.mapName, s.player.Pos, got, want)
	}
	return got
}

func TestScene_OnADarkMapTheCritterCarriesALanternThatFollowsIt(t *testing.T) {
	s, err := NewScene(nil, sceneLoader("house", "from_village"), "dark_cave")
	if err != nil {
		t.Fatalf("NewScene: %v", err)
	}
	// The spawn is at feet 24,40; the lantern sits 8 above the feet.
	if got := lights(t, s, 1)[0]; got != (Light{Pos: geom.Vec{X: 24, Y: 32}, Radius: 56}) {
		t.Errorf("at the spawn: lantern = %+v; want radius 56 at 24,32", got)
	}

	// A second of walking west stops the feet at the cave's west edge, 6,40.
	for range 60 {
		if err := s.Update(walkFrame(geom.Vec{X: -1})); err != nil {
			t.Fatalf("Update: %v", err)
		}
	}
	if got := lights(t, s, 1)[0]; got != (Light{Pos: geom.Vec{X: 6, Y: 32}, Radius: 56}) {
		t.Errorf("after walking west: lantern = %+v; want radius 56 at 6,32", got)
	}
}

func TestScene_ALitMapHasNoLights(t *testing.T) {
	tests := []struct {
		name     string
		startMap string
	}{
		{"no dark property", "village"},
		{"dark set to false", "lit_meadow"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, err := NewScene(nil, sceneLoader("house", "from_village"), tt.startMap)
			if err != nil {
				t.Fatalf("NewScene: %v", err)
			}
			lights(t, s, 0)
		})
	}
}

func TestScene_WarpingIntoADarkMapLightsTheLanternAndWarpingOutPutsItOut(t *testing.T) {
	s := newVillageScene(t, sceneLoader("dark_cave", "from_village"))
	east := geom.Vec{X: 1}

	updateUntilMap(t, s, "dark_cave", walkFrame(east))
	// Arriving on the cave's from_village spawn, feet at 24,40.
	if got := lights(t, s, 1)[0]; got != (Light{Pos: geom.Vec{X: 24, Y: 32}, Radius: 56}) {
		t.Errorf("in the cave: lantern = %+v; want radius 56 at 24,32", got)
	}

	updateUntilMap(t, s, "village", walkFrame(east))
	lights(t, s, 0)
}
