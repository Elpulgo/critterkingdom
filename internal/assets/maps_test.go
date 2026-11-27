package assets_test

import (
	"strings"
	"testing"

	"github.com/Elpulgo/critterkingdom/internal/assets"
	"github.com/Elpulgo/critterkingdom/internal/world"
)

func TestLoadMap_LoadsMeadowWithPlayerSpawnOnOpenGroundInsideMap(t *testing.T) {
	if _, err := assets.LoadMap("no_such_map"); err == nil {
		t.Fatal("LoadMap(no_such_map) succeeded; want error for unknown map name")
	}

	m, err := assets.LoadMap("meadow")
	if err != nil {
		t.Fatalf("LoadMap(meadow): %v", err)
	}

	if m.Cols != 40 || m.Rows != 30 || m.TileSize != 16 {
		t.Fatalf("meadow dimensions = %dx%d tileSize %v; want 40x30 tileSize 16", m.Cols, m.Rows, m.TileSize)
	}

	for _, name := range []string{"ground", "decor", "above", "collision"} {
		if m.TileLayer(name) == nil {
			t.Errorf("TileLayer(%q) = nil; want layer present", name)
		}
	}
	for _, name := range []string{"props", "objects"} {
		if m.ObjectLayer(name) == nil {
			t.Errorf("ObjectLayer(%q) = nil; want layer present", name)
		}
	}

	objects := m.ObjectLayer("objects")
	if objects == nil {
		t.Fatal("ObjectLayer(objects) = nil")
	}
	var spawns []world.Object
	for _, obj := range objects.Objects {
		if obj.Type == "player_spawn" {
			spawns = append(spawns, obj)
		}
	}
	if len(spawns) != 1 {
		t.Fatalf("player_spawn count = %d; want 1", len(spawns))
	}
	spawn := spawns[0]
	if !spawn.Point {
		t.Errorf("player_spawn Point = false; want true")
	}
	size := m.Size()
	if spawn.X <= 0 || spawn.X >= size.X || spawn.Y <= 0 || spawn.Y >= size.Y {
		t.Fatalf("player_spawn at (%v, %v) is outside map bounds (0,0)..(%v,%v)", spawn.X, spawn.Y, size.X, size.Y)
	}
	col := int(spawn.X / m.TileSize)
	row := int(spawn.Y / m.TileSize)
	if m.Collision.Solid(col, row) {
		t.Errorf("player_spawn cell (%d, %d) is solid; want open ground", col, row)
	}

	// Every border cell around the meadow is solid.
	for c := 0; c < m.Cols; c++ {
		if !m.Collision.Solid(c, 0) {
			t.Errorf("top border cell (%d, 0) is not solid", c)
		}
		if !m.Collision.Solid(c, m.Rows-1) {
			t.Errorf("bottom border cell (%d, %d) is not solid", c, m.Rows-1)
		}
	}
	for r := 0; r < m.Rows; r++ {
		if !m.Collision.Solid(0, r) {
			t.Errorf("left border cell (0, %d) is not solid", r)
		}
		if !m.Collision.Solid(m.Cols-1, r) {
			t.Errorf("right border cell (%d, %d) is not solid", m.Cols-1, r)
		}
	}

	// Every non-zero GID used in tile layers and object layers resolves to a tile with a non-empty type.
	for _, name := range []string{"ground", "decor", "above", "collision"} {
		tl := m.TileLayer(name)
		if tl == nil {
			continue
		}
		for i, gid := range tl.GIDs {
			if gid == 0 {
				continue
			}
			tile, ok := m.Tile(gid)
			if !ok || tile.Type == "" {
				t.Errorf("layer %q cell %d gid %d has tile %+v (ok=%v); want non-empty tile type", name, i, gid, tile, ok)
			}
		}
	}
	for _, name := range []string{"props", "objects"} {
		ol := m.ObjectLayer(name)
		if ol == nil {
			continue
		}
		for _, obj := range ol.Objects {
			if obj.GID == 0 {
				continue
			}
			tile, ok := m.Tile(obj.GID)
			if !ok || tile.Type == "" {
				t.Errorf("layer %q object %d gid %d has tile %+v (ok=%v); want non-empty tile type", name, obj.ID, obj.GID, tile, ok)
			}
		}
	}
}

func TestLoadMap_MeadowWaterTilesAreAnimatedAndSolid(t *testing.T) {
	m, err := assets.LoadMap("meadow")
	if err != nil {
		t.Fatalf("LoadMap(meadow): %v", err)
	}

	ground := m.TileLayer("ground")
	if ground == nil {
		t.Fatal("TileLayer(ground) = nil")
	}

	waterCount := 0
	for row := 0; row < m.Rows; row++ {
		for col := 0; col < m.Cols; col++ {
			gid := ground.At(col, row)
			tile, ok := m.Tile(gid)
			if !ok {
				t.Fatalf("ground(%d, %d) gid %d not in tileset", col, row, gid)
			}
			if !strings.HasPrefix(tile.Type, "water") {
				continue
			}
			waterCount++
			if len(tile.Animation) == 0 {
				t.Errorf("water tile at (%d, %d) gid %d has no animation", col, row, gid)
			}
			if !m.Collision.Solid(col, row) {
				t.Errorf("water tile at (%d, %d) is not solid in Collision grid", col, row)
			}
		}
	}
	if waterCount == 0 {
		t.Fatal("meadow ground layer has no water tiles; want a pond of animated water")
	}
}
