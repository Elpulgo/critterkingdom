package assets_test

import (
	"math"
	"reflect"
	"sort"
	"testing"

	"github.com/Elpulgo/critterkingdom/internal/assets"
	"github.com/Elpulgo/critterkingdom/internal/combat"
	"github.com/Elpulgo/critterkingdom/internal/critter"
	"github.com/Elpulgo/critterkingdom/internal/geom"
	"github.com/Elpulgo/critterkingdom/internal/world"
)

// The content lint: rules every embedded map must keep, so a broken or
// unreachable warp fails the tests instead of a game.

// loadEmbeddedMaps loads every map MapNames lists, by name.
func loadEmbeddedMaps(t *testing.T) map[string]*world.Map {
	t.Helper()
	names, err := assets.MapNames()
	if err != nil {
		t.Fatalf("MapNames: %v", err)
	}
	if len(names) == 0 {
		t.Fatal("MapNames returned no maps")
	}
	maps := make(map[string]*world.Map, len(names))
	for _, name := range names {
		m, err := assets.LoadMap(name)
		if err != nil {
			t.Fatalf("LoadMap(%s): %v", name, err)
		}
		maps[name] = m
	}
	return maps
}

// namedSpawn is a player_spawn object of a map.
type namedSpawn struct {
	name string
	pos  geom.Vec
}

// spawnsOf lists the player_spawn objects of m's objects layer.
func spawnsOf(m *world.Map) []namedSpawn {
	var spawns []namedSpawn
	if l := m.ObjectLayer("objects"); l != nil {
		for _, o := range l.Objects {
			if o.Type == "player_spawn" {
				spawns = append(spawns, namedSpawn{o.Name, geom.Vec{X: o.X, Y: o.Y}})
			}
		}
	}
	return spawns
}

func TestEmbeddedMaps_EveryWarpLeadsToAnExistingMapAndSpawn(t *testing.T) {
	for name, m := range loadEmbeddedMaps(t) {
		for _, w := range m.Warps() {
			if _, _, err := world.ResolveWarp(assets.LoadMap, w); err != nil {
				t.Errorf("map %s: warp at %v: %v", name, w.Rect, err)
			}
		}
	}
}

func TestEmbeddedMaps_SpawnsStandOnOpenGroundOutsideEveryWarp(t *testing.T) {
	for name, m := range loadEmbeddedMaps(t) {
		if _, ok := m.DefaultSpawn(); !ok {
			t.Errorf("map %s has no player_spawn to start on", name)
		}
		size := m.Size()
		for _, s := range spawnsOf(m) {
			if s.pos.X <= 0 || s.pos.X >= size.X || s.pos.Y <= 0 || s.pos.Y >= size.Y {
				t.Errorf("map %s: spawn %q at %v is outside the map (0,0)..%v", name, s.name, s.pos, size)
				continue
			}
			col, row := int(s.pos.X/m.TileSize), int(s.pos.Y/m.TileSize)
			if m.Collision.Solid(col, row) {
				t.Errorf("map %s: spawn %q stands on solid cell (%d, %d)", name, s.name, col, row)
			}
			if w, ok := m.WarpAt(s.pos); ok {
				t.Errorf("map %s: spawn %q at %v is inside the warp to %s/%s, so arriving warps straight away", name, s.name, s.pos, w.TargetMap, w.TargetSpawn)
			}
		}
	}
}

func TestEmbeddedMaps_EveryWarpCanBeWalkedToFromEverySpawn(t *testing.T) {
	for name, m := range loadEmbeddedMaps(t) {
		for _, s := range spawnsOf(m) {
			reached := openCellsReachableFrom(m, int(s.pos.X/m.TileSize), int(s.pos.Y/m.TileSize))
			for _, w := range m.Warps() {
				if !anyCellReached(m, reached, w.Rect) {
					t.Errorf("map %s: the warp to %s/%s at %v cannot be walked to from spawn %q", name, w.TargetMap, w.TargetSpawn, w.Rect, s.name)
				}
			}
		}
	}
}

func TestEmbeddedMaps_SpawnZonesLieInsideTheMapClearOfWarpsAndSpawnsOnReachableGround(t *testing.T) {
	for name, m := range loadEmbeddedMaps(t) {
		bounds := geom.Rect{Max: m.Size()}
		spawns := spawnsOf(m)
		for _, z := range m.SpawnZones() {
			if z.Rect.Min.X < bounds.Min.X || z.Rect.Min.Y < bounds.Min.Y || z.Rect.Max.X > bounds.Max.X || z.Rect.Max.Y > bounds.Max.Y {
				t.Errorf("map %s: spawn zone %s at %v reaches outside the map %v", name, z.ID, z.Rect, bounds)
			}
			for _, w := range m.Warps() {
				if overlaps(z.Rect, w.Rect) {
					t.Errorf("map %s: spawn zone %s at %v overlaps the warp to %s/%s at %v", name, z.ID, z.Rect, w.TargetMap, w.TargetSpawn, w.Rect)
				}
			}
			for _, s := range spawns {
				if z.Rect.Contains(s.pos) {
					t.Errorf("map %s: spawn zone %s at %v covers spawn %q at %v", name, z.ID, z.Rect, s.name, s.pos)
				}
				reached := openCellsReachableFrom(m, int(s.pos.X/m.TileSize), int(s.pos.Y/m.TileSize))
				if !anyCellReached(m, reached, z.Rect) {
					t.Errorf("map %s: spawn zone %s at %v has no open cell that can be walked to from spawn %q", name, z.ID, z.Rect, s.name)
				}
			}
		}
	}
}

func TestEmbeddedMaps_EverySpawnZoneHasASpeciesThatComesOutByDay(t *testing.T) {
	species, err := assets.LoadCritters()
	if err != nil {
		t.Fatalf("LoadCritters: %v", err)
	}
	// The world the game hands the spawner until night (#50), quest flags,
	// companions and the daily visitor exist.
	day := critter.World{TimeOfDay: critter.Day}
	for name, m := range loadEmbeddedMaps(t) {
		for _, z := range m.SpawnZones() {
			found := false
			for i := range species {
				if species[i].Habitat == z.Habitat && species[i].Active(day) {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("map %s: spawn zone %s has habitat %q, which no species of critters.json lives in by day", name, z.ID, z.Habitat)
			}
		}
	}
}

func TestEmbeddedMaps_VillagersStartOnOpenGroundAndWanderInsideTheMapClearOfWarps(t *testing.T) {
	for name, m := range loadEmbeddedMaps(t) {
		size := m.Size()
		for _, n := range m.NPCs() {
			if n.Pos.X <= 0 || n.Pos.X >= size.X || n.Pos.Y <= 0 || n.Pos.Y >= size.Y {
				t.Errorf("map %s: npc %q (object %d) starts at %v, outside the map (0,0)..%v", name, n.ID, n.ObjectID, n.Pos, size)
			} else if col, row := int(n.Pos.X/m.TileSize), int(n.Pos.Y/m.TileSize); m.Collision.Solid(col, row) {
				t.Errorf("map %s: npc %q (object %d) starts on solid cell (%d, %d)", name, n.ID, n.ObjectID, col, row)
			}
			if w, ok := m.WarpAt(n.Pos); ok {
				t.Errorf("map %s: npc %q (object %d) starts at %v inside the warp to %s/%s", name, n.ID, n.ObjectID, n.Pos, w.TargetMap, w.TargetSpawn)
			}
			if n.Area.Min.X < 0 || n.Area.Min.Y < 0 || n.Area.Max.X > size.X || n.Area.Max.Y > size.Y {
				t.Errorf("map %s: npc %q (object %d) wanders in %v, which reaches outside the map (0,0)..%v", name, n.ID, n.ObjectID, n.Area, size)
			}
			for _, w := range m.Warps() {
				if overlaps(n.Area, w.Rect) {
					t.Errorf("map %s: npc %q (object %d) wanders in %v, which overlaps the warp to %s/%s at %v", name, n.ID, n.ObjectID, n.Area, w.TargetMap, w.TargetSpawn, w.Rect)
				}
			}
		}
	}
}

func TestEmbeddedMaps_VillageHasGrandmaHedgehogBakerBadgerAndMayorOwl(t *testing.T) {
	m, err := assets.LoadMap("village")
	if err != nil {
		t.Fatalf("LoadMap(village): %v", err)
	}
	got := []string{}
	for _, n := range m.NPCs() {
		got = append(got, n.ID)
	}
	sort.Strings(got)
	want := []string{"baker_badger", "grandma_hedgehog", "mayor_owl"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("village villagers = %v; want %v", got, want)
	}
}

func TestEmbeddedMaps_VillageJoinsTheHouseAndTheMeadowBothWays(t *testing.T) {
	maps := loadEmbeddedMaps(t)
	got := make(map[string][]string, len(maps))
	for name, m := range maps {
		dests := []string{}
		for _, w := range m.Warps() {
			dests = append(dests, w.TargetMap+"/"+w.TargetSpawn)
		}
		sort.Strings(dests)
		got[name] = dests
	}
	want := map[string][]string{
		"dark_test": {},
		"house":     {"village/from_house"},
		"meadow":    {"village/from_meadow"},
		"testroom":  {},
		"village":   {"house/from_village", "meadow/from_village"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("warps by map = %v; want %v", got, want)
	}
}

func TestEmbeddedMaps_EnemiesAreKnownStandOnOpenGroundAndNeverInTheVillage(t *testing.T) {
	defs := enemyDefsByID(t)
	for name, m := range loadEmbeddedMaps(t) {
		enemies := m.Enemies()
		if (m.Biome == "village" || m.Biome == "house") && len(enemies) > 0 {
			t.Errorf("map %s has biome %s, part of the village, but %d enemy objects; enemies never appear in the village", name, m.Biome, len(enemies))
		}
		size := m.Size()
		for _, e := range enemies {
			if _, ok := defs[e.EnemyID]; !ok {
				t.Errorf("map %s: enemy object %d names %q, which enemies.json does not define", name, e.ObjectID, e.EnemyID)
			}
			if e.Pos.X <= 0 || e.Pos.X >= size.X || e.Pos.Y <= 0 || e.Pos.Y >= size.Y {
				t.Errorf("map %s: enemy object %d at %v is outside the map (0,0)..%v", name, e.ObjectID, e.Pos, size)
				continue
			}
			col, row := int(e.Pos.X/m.TileSize), int(e.Pos.Y/m.TileSize)
			if m.Collision.Solid(col, row) {
				t.Errorf("map %s: enemy object %d (%s) stands on solid cell (%d, %d)", name, e.ObjectID, e.EnemyID, col, row)
			}
		}
	}
}

func TestEmbeddedMaps_MeadowAndTestRoomHoldEveryEnemyAndStartAScene(t *testing.T) {
	defs, err := assets.LoadEnemies()
	if err != nil {
		t.Fatalf("LoadEnemies: %v", err)
	}
	for _, name := range []string{"meadow", "testroom"} {
		t.Run(name, func(t *testing.T) {
			m, err := assets.LoadMap(name)
			if err != nil {
				t.Fatalf("LoadMap(%s): %v", name, err)
			}
			var ids []string
			for _, e := range m.Enemies() {
				ids = append(ids, e.EnemyID)
			}
			sort.Strings(ids)
			want := []string{"grumpy_mushroom", "sulky_slime", "thistle_beast"}
			if !reflect.DeepEqual(ids, want) {
				t.Errorf("enemies = %v; want one of each, %v", ids, want)
			}
			if _, err := world.NewScene(nil, assets.LoadMap, name, world.WithEnemies(defs)); err != nil {
				t.Errorf("NewScene(%s) with the embedded enemies: %v", name, err)
			}
		})
	}
}

func TestEmbeddedMaps_NoEnemyOnTheMeadowOrInTheTestRoomNoticesACritterOnASpawn(t *testing.T) {
	notice := map[string]float64{
		combat.BehaviourHopToward: combat.HopNoticeRange,
		combat.BehaviourCharge:    combat.ChargeNoticeRange,
		combat.BehaviourPuff:      combat.PuffNoticeRange,
	}
	defs := enemyDefsByID(t)
	for _, name := range []string{"meadow", "testroom"} {
		m, err := assets.LoadMap(name)
		if err != nil {
			t.Fatalf("LoadMap(%s): %v", name, err)
		}
		for _, s := range spawnsOf(m) {
			for _, e := range m.Enemies() {
				r, ok := notice[defs[e.EnemyID].Behaviour]
				if !ok {
					t.Fatalf("map %s: enemy object %d (%s) has no known behaviour", name, e.ObjectID, e.EnemyID)
				}
				if d := e.Pos.Sub(s.pos).Len(); d <= r {
					t.Errorf("map %s: %s at %v is %.1f from spawn %q, within its notice range %v, so it attacks a critter arriving there", name, e.EnemyID, e.Pos, d, s.name, r)
				}
			}
		}
	}
}

// enemyDefsByID loads the embedded enemies.json, by enemy id.
func enemyDefsByID(t *testing.T) map[string]combat.EnemyDef {
	t.Helper()
	defs, err := assets.LoadEnemies()
	if err != nil {
		t.Fatalf("LoadEnemies: %v", err)
	}
	byID := make(map[string]combat.EnemyDef, len(defs))
	for _, d := range defs {
		byID[d.ID] = d
	}
	return byID
}

// openCellsReachableFrom flood-fills m's non-solid cells from (col, row),
// stepping up, down, left and right, and returns the cells it reaches. A
// solid start cell reaches nothing.
func openCellsReachableFrom(m *world.Map, col, row int) map[[2]int]bool {
	reached := make(map[[2]int]bool)
	open := func(c, r int) bool {
		return c >= 0 && c < m.Cols && r >= 0 && r < m.Rows && !m.Collision.Solid(c, r)
	}
	if !open(col, row) {
		return reached
	}
	queue := [][2]int{{col, row}}
	reached[queue[0]] = true
	for len(queue) > 0 {
		cell := queue[0]
		queue = queue[1:]
		for _, step := range [][2]int{{1, 0}, {-1, 0}, {0, 1}, {0, -1}} {
			next := [2]int{cell[0] + step[0], cell[1] + step[1]}
			if !reached[next] && open(next[0], next[1]) {
				reached[next] = true
				queue = append(queue, next)
			}
		}
	}
	return reached
}

// anyCellReached reports whether a cell that rect covers is in reached.
func anyCellReached(m *world.Map, reached map[[2]int]bool, rect geom.Rect) bool {
	c0, r0 := int(math.Floor(rect.Min.X/m.TileSize)), int(math.Floor(rect.Min.Y/m.TileSize))
	c1, r1 := int(math.Ceil(rect.Max.X/m.TileSize)), int(math.Ceil(rect.Max.Y/m.TileSize))
	for r := r0; r < r1; r++ {
		for c := c0; c < c1; c++ {
			if reached[[2]int{c, r}] {
				return true
			}
		}
	}
	return false
}

// overlaps reports whether a and b share some area; rectangles that only
// touch along an edge do not overlap.
func overlaps(a, b geom.Rect) bool {
	return a.Min.X < b.Max.X && b.Min.X < a.Max.X && a.Min.Y < b.Max.Y && b.Min.Y < a.Max.Y
}
