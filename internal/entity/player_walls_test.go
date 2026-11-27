package entity_test

import (
	"math"
	"testing"

	"github.com/Elpulgo/critterkingdom/internal/entity"
	"github.com/Elpulgo/critterkingdom/internal/geom"
	"github.com/Elpulgo/critterkingdom/internal/world"
)

// tick is one tick at 60 ticks per second (WalkSpeed 72 px/s -> 1.2 px/tick).
const tick = 1.0 / 60

// wallTestTMJ is a 3x3 map of 16 px world tiles where the middle cell (1,1)
// covering x=16..32, y=16..32 is solid and the surrounding cells are open.
const wallTestTMJ = `{
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

var mapArea = geom.Rect{Max: geom.Vec{X: 48, Y: 48}}

func loadTestGrid(t *testing.T) *world.Grid {
	t.Helper()
	m, err := world.LoadMap([]byte(wallTestTMJ))
	if err != nil {
		t.Fatalf("LoadMap: %v", err)
	}
	return m.Collision
}

func TestPlayer_CannotWalkIntoSolidTile(t *testing.T) {
	p := entity.NewPlayer(geom.Vec{X: 9, Y: 24})
	p.Walls = loadTestGrid(t)

	// At Pos (9, 24), Box() is [3..15, 14..24] (12 wide by 10 tall, bottom
	// centre at Pos), 1 px left of the solid tile at x=16..32.
	if got, want := p.Box(), (geom.Rect{Min: geom.Vec{X: 3, Y: 14}, Max: geom.Vec{X: 15, Y: 24}}); got != want {
		t.Fatalf("initial Box() = %v; want %v", got, want)
	}

	// One tick walking right tries to move 1.2 px, but stops flush at x=16
	// (Pos.X = 10, Box.Max.X = 16).
	p.Update(tick, geom.Vec{X: 1}, mapArea)
	if got, want := p.Pos, (geom.Vec{X: 10, Y: 24}); !near(got, want) {
		t.Fatalf("after 1st tick into wall: Pos = %v; want %v", got, want)
	}
	if got, want := p.Box(), (geom.Rect{Min: geom.Vec{X: 4, Y: 14}, Max: geom.Vec{X: 16, Y: 24}}); !nearRect(got, want) {
		t.Fatalf("after 1st tick into wall: Box() = %v; want %v", got, want)
	}

	// Pushing into the wall on the next tick moves nowhere.
	p.Update(tick, geom.Vec{X: 1}, mapArea)
	if got, want := p.Pos, (geom.Vec{X: 10, Y: 24}); !near(got, want) {
		t.Errorf("after 2nd tick into wall: Pos = %v; want %v", got, want)
	}
	if p.Moved() {
		t.Errorf("Moved() = true when blocked flush against wall; want false")
	}
}

func TestPlayer_SlidesAlongWallWhenPushedDiagonally(t *testing.T) {
	p := entity.NewPlayer(geom.Vec{X: 9.5, Y: 24})
	p.Walls = loadTestGrid(t)

	// Pushing diagonally with dir=(0.8, 0.6) at step=1.2 px tries delta
	// (0.96, 0.72). X is clamped to 0.5 px (so Box.Max.X = 15.5 + 0.5 = 16,
	// flush with the wall at x=16), while Y slides the full 0.72 px to 24.72.
	p.Update(tick, geom.Vec{X: 0.8, Y: 0.6}, mapArea)

	if got, want := p.Pos, (geom.Vec{X: 10, Y: 24.72}); !near(got, want) {
		t.Fatalf("after diagonal push: Pos = %v; want %v", got, want)
	}
	if !p.Moved() {
		t.Errorf("Moved() = false while sliding along wall; want true")
	}
}

func TestPlayer_GivesUpWalkTargetBlockedByWall(t *testing.T) {
	p := entity.NewPlayer(geom.Vec{X: 9, Y: 24})
	p.Walls = loadTestGrid(t)
	p.WalkTo(geom.Vec{X: 40, Y: 24})

	// First tick moves 1.0 px of the 1.2 px step (to Pos.X = 10, flush with
	// the wall at x=16), which is >= 10% of the step so the critter is still
	// walking.
	p.Update(tick, geom.Vec{}, mapArea)
	if _, walking := p.Target(); !walking {
		t.Fatalf("after 1st tick: walking = false; want true after moving 1.0 of 1.2 px")
	}

	// Second tick moves 0 px (< 0.12 px, a tenth of the 1.2 px intended step),
	// so the critter drops its walk target instead of walking on the spot.
	p.Update(tick, geom.Vec{}, mapArea)
	if _, walking := p.Target(); walking {
		t.Fatalf("after 2nd tick against wall: walking = true; want false")
	}
	if got, want := p.Pos, (geom.Vec{X: 10, Y: 24}); !near(got, want) {
		t.Errorf("Pos = %v; want %v", got, want)
	}
}

func near(a, b geom.Vec) bool {
	return math.Abs(a.X-b.X) < 1e-9 && math.Abs(a.Y-b.Y) < 1e-9
}

func nearRect(a, b geom.Rect) bool {
	return near(a.Min, b.Min) && near(a.Max, b.Max)
}
