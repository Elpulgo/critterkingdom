package world

import (
	"testing"

	"github.com/Elpulgo/critterkingdom/internal/geom"
)

func TestLoadMap_TileCollisionLayer_MarksNonZeroTilesSolid(t *testing.T) {
	const tmj = `{
		"orientation": "orthogonal",
		"infinite": false,
		"width": 3,
		"height": 2,
		"tilewidth": 128,
		"tileheight": 128,
		"tilesets": [
			{"firstgid": 1, "name": "solid", "tilecount": 2}
		],
		"layers": [
			{
				"name": "collision",
				"type": "tilelayer",
				"width": 3,
				"height": 2,
				"data": [0, 1, 0, 2, 0, 2147483649]
			}
		]
	}`

	m, err := LoadMap([]byte(tmj))
	if err != nil {
		t.Fatalf("LoadMap: %v", err)
	}
	if m.Collision == nil {
		t.Fatal("Map.Collision = nil")
	}

	wantSolid := [2][3]bool{
		{false, true, false},
		{true, false, true},
	}
	for row := 0; row < 2; row++ {
		for col := 0; col < 3; col++ {
			if got := m.Collision.Solid(col, row); got != wantSolid[row][col] {
				t.Errorf("Collision.Solid(%d, %d) = %v; want %v", col, row, got, wantSolid[row][col])
			}
		}
	}

	// Cells outside the grid are never solid.
	for _, cell := range [][2]int{{-1, 0}, {0, -1}, {3, 0}, {0, 2}, {10, 10}} {
		if m.Collision.Solid(cell[0], cell[1]) {
			t.Errorf("Collision.Solid(%d, %d) = true outside grid; want false", cell[0], cell[1])
		}
	}
}

func TestLoadMap_ObjectCollisionLayer_MarksEveryOverlappedCellSolid(t *testing.T) {
	// Map is 4x3 tiles (each tile is 128 Tiled px = 16 world px).
	// Object 1: x=64..256 (world 8..32), y=64..192 (world 8..24) -> overlaps
	// cols 0,1 (x 8..16 and 16..32, touching col 2 at x=32 with zero area) and
	// rows 0,1 (y 8..16 and 16..24).
	// Object 2: x=384..512 (world 48..64, exactly col 3), y=256..384 (world 32..48, exactly row 2).
	const tmj = `{
		"orientation": "orthogonal",
		"infinite": false,
		"width": 4,
		"height": 3,
		"tilewidth": 128,
		"tileheight": 128,
		"tilesets": [],
		"layers": [
			{
				"name": "collision",
				"type": "objectgroup",
				"objects": [
					{"id": 1, "x": 64, "y": 64, "width": 192, "height": 128},
					{"id": 2, "x": 384, "y": 256, "width": 128, "height": 128}
				]
			}
		]
	}`

	m, err := LoadMap([]byte(tmj))
	if err != nil {
		t.Fatalf("LoadMap: %v", err)
	}
	if m.Collision == nil {
		t.Fatal("Map.Collision = nil")
	}

	wantSolid := [3][4]bool{
		{true, true, false, false},
		{true, true, false, false},
		{false, false, false, true},
	}
	for row := 0; row < 3; row++ {
		for col := 0; col < 4; col++ {
			if got := m.Collision.Solid(col, row); got != wantSolid[row][col] {
				t.Errorf("Collision.Solid(%d, %d) = %v; want %v", col, row, got, wantSolid[row][col])
			}
		}
	}
}

func TestGrid_Slide_StopsFlushAgainstSolidTile(t *testing.T) {
	// 3x3 map of 16 px tiles, with center cell (1,1) [16..32, 16..32] solid.
	const tmj = `{
		"orientation": "orthogonal",
		"infinite": false,
		"width": 3,
		"height": 3,
		"tilewidth": 128,
		"tileheight": 128,
		"tilesets": [
			{"firstgid": 1, "name": "solid", "tilecount": 1}
		],
		"layers": [
			{
				"name": "collision",
				"type": "tilelayer",
				"width": 3,
				"height": 3,
				"data": [
					0, 0, 0,
					0, 1, 0,
					0, 0, 0
				]
			}
		]
	}`

	m, err := LoadMap([]byte(tmj))
	if err != nil {
		t.Fatalf("LoadMap: %v", err)
	}
	g := m.Collision

	tests := []struct {
		name  string
		box   geom.Rect
		delta geom.Vec
		want  geom.Vec
	}{
		{
			name:  "moving right into solid cell stops at left edge x=16",
			box:   geom.Rect{Min: geom.Vec{X: 2, Y: 18}, Max: geom.Vec{X: 14, Y: 28}},
			delta: geom.Vec{X: 5, Y: 0},
			want:  geom.Vec{X: 2, Y: 0},
		},
		{
			name:  "moving left into solid cell stops at right edge x=32",
			box:   geom.Rect{Min: geom.Vec{X: 35, Y: 18}, Max: geom.Vec{X: 47, Y: 28}},
			delta: geom.Vec{X: -6, Y: 0},
			want:  geom.Vec{X: -3, Y: 0},
		},
		{
			name:  "moving down into solid cell stops at top edge y=16",
			box:   geom.Rect{Min: geom.Vec{X: 18, Y: 3}, Max: geom.Vec{X: 30, Y: 13}},
			delta: geom.Vec{X: 0, Y: 7},
			want:  geom.Vec{X: 0, Y: 3},
		},
		{
			name:  "moving up into solid cell stops at bottom edge y=32",
			box:   geom.Rect{Min: geom.Vec{X: 18, Y: 34}, Max: geom.Vec{X: 30, Y: 44}},
			delta: geom.Vec{X: 0, Y: -5},
			want:  geom.Vec{X: 0, Y: -2},
		},
		{
			name:  "free move in open space returns delta unchanged",
			box:   geom.Rect{Min: geom.Vec{X: 2, Y: 2}, Max: geom.Vec{X: 14, Y: 12}},
			delta: geom.Vec{X: 4, Y: 3},
			want:  geom.Vec{X: 4, Y: 3},
		},
		{
			name:  "box touching solid cell edge is not blocked moving along the edge",
			box:   geom.Rect{Min: geom.Vec{X: 4, Y: 6}, Max: geom.Vec{X: 16, Y: 16}},
			delta: geom.Vec{X: 0, Y: 5},
			want:  geom.Vec{X: 0, Y: 5},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := g.Slide(tt.box, tt.delta); got != tt.want {
				t.Errorf("Slide(%v, %v) = %v; want %v", tt.box, tt.delta, got, tt.want)
			}
		})
	}
}

func TestGrid_Slide_SlidesAlongWallOnDiagonalMove(t *testing.T) {
	// 3x3 map of 16 px tiles: right wall at col 2 (x = 32..48) and bottom wall at row 2 (y = 32..48).
	const tmj = `{
		"orientation": "orthogonal",
		"infinite": false,
		"width": 3,
		"height": 3,
		"tilewidth": 128,
		"tileheight": 128,
		"tilesets": [
			{"firstgid": 1, "name": "solid", "tilecount": 1}
		],
		"layers": [
			{
				"name": "collision",
				"type": "tilelayer",
				"width": 3,
				"height": 3,
				"data": [
					0, 0, 1,
					0, 0, 1,
					1, 1, 1
				]
			}
		]
	}`

	m, err := LoadMap([]byte(tmj))
	if err != nil {
		t.Fatalf("LoadMap: %v", err)
	}
	g := m.Collision

	// Box [18..30, 4..14] moves diagonally (5, 4) toward the vertical wall at x=32:
	// X stops after 2 px (at x=32), while Y moves the full 4 px along the wall.
	boxNearVerticalWall := geom.Rect{Min: geom.Vec{X: 18, Y: 4}, Max: geom.Vec{X: 30, Y: 14}}
	if got, want := g.Slide(boxNearVerticalWall, geom.Vec{X: 5, Y: 4}), (geom.Vec{X: 2, Y: 4}); got != want {
		t.Errorf("Slide against vertical wall = %v; want %v", got, want)
	}

	// Box [4..16, 21..31] moves diagonally (3, 4) toward the horizontal wall at y=32:
	// X moves the full 3 px along the wall, while Y stops after 1 px (at y=32).
	boxNearHorizontalWall := geom.Rect{Min: geom.Vec{X: 4, Y: 21}, Max: geom.Vec{X: 16, Y: 31}}
	if got, want := g.Slide(boxNearHorizontalWall, geom.Vec{X: 3, Y: 4}), (geom.Vec{X: 3, Y: 1}); got != want {
		t.Errorf("Slide against horizontal wall = %v; want %v", got, want)
	}
}
