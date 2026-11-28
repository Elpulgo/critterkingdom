package entity_test

import (
	"math"
	"math/rand/v2"
	"testing"

	"github.com/Elpulgo/critterkingdom/internal/entity"
	"github.com/Elpulgo/critterkingdom/internal/geom"
	"github.com/Elpulgo/critterkingdom/internal/world"
)

// farAway is a player position nowhere near any test zone.
var farAway = geom.Vec{X: 1000, Y: 1000}

func seeded(seed uint64) *rand.Rand {
	return rand.New(rand.NewPCG(seed, 7))
}

func TestWild_LeftAloneRestsAndHopsAndEveryHopEndsInsideItsZone(t *testing.T) {
	// A zone small enough that hops of up to 12 px often aim past its edge.
	zone := geom.Rect{Min: geom.Vec{X: 40, Y: 40}, Max: geom.Vec{X: 64, Y: 56}}
	w := entity.NewWild(geom.Vec{X: 52, Y: 48}, zone, seeded(1))

	var stillTicks, hops, movingHops int
	hopStart := w.Pos
	sawLift := false
	prevState := w.State()
	// One minute of game time.
	for range 60 * 60 {
		before := w.Pos
		w.Update(tick, farAway)

		if !zone.Contains(w.Pos) {
			t.Fatalf("critter at %v left its zone %v", w.Pos, zone)
		}
		switch w.State() {
		case entity.WildIdle:
			if w.Pos != before {
				t.Fatalf("critter moved from %v to %v while idle", before, w.Pos)
			}
			if w.HopHeight() != 0 {
				t.Fatalf("idle critter is %v px off the ground", w.HopHeight())
			}
			stillTicks++
			if prevState == entity.WildHopping {
				hops++
				if w.Pos != hopStart {
					movingHops++
				}
			}
		case entity.WildHopping:
			if prevState == entity.WildIdle {
				hopStart = before
			}
			if w.HopHeight() > 0 {
				sawLift = true
			}
		}
		prevState = w.State()
	}

	// Rests last 0.8 to 2.5 s and a hop 0.35 s, so a minute holds roughly
	// 20 to 50 hops.
	if hops < 20 || hops > 50 {
		t.Errorf("critter finished %d hops in a minute, want 20 to 50", hops)
	}
	if movingHops < hops/2 {
		t.Errorf("only %d of %d hops moved the critter", movingHops, hops)
	}
	if stillTicks < 60*60/2 {
		t.Errorf("critter stood still for %d of 3600 ticks, want most of them", stillTicks)
	}
	if !sawLift {
		t.Error("critter never left the ground while hopping")
	}
}

func TestWild_TheSameSeedAndTicksEndAtTheSamePosition(t *testing.T) {
	zone := geom.Rect{Min: geom.Vec{X: 0, Y: 0}, Max: geom.Vec{X: 200, Y: 200}}
	a := entity.NewWild(geom.Vec{X: 100, Y: 100}, zone, seeded(42))
	b := entity.NewWild(geom.Vec{X: 100, Y: 100}, zone, seeded(42))
	other := entity.NewWild(geom.Vec{X: 100, Y: 100}, zone, seeded(43))

	for range 30 * 60 {
		a.Update(tick, farAway)
		b.Update(tick, farAway)
		other.Update(tick, farAway)
	}

	if a.Pos != b.Pos || a.Facing != b.Facing {
		t.Errorf("same seed ended at %v facing %v and at %v facing %v", a.Pos, a.Facing, b.Pos, b.Facing)
	}
	if a.Pos == other.Pos {
		t.Errorf("a different seed ended at the same %v; the rng is not driving the hops", a.Pos)
	}
}

func TestWild_InAPartlySolidZoneNeverStandsInAWall(t *testing.T) {
	// The 48x48 test map has its middle cell, x and y 16..32, solid. The zone
	// lets the feet go anywhere a 12x10 box fits on the map, so it covers the
	// solid cell and hops keep aiming into it.
	grid := loadTestGrid(t)
	zone := geom.Rect{Min: geom.Vec{X: 6, Y: 10}, Max: geom.Vec{X: 42, Y: 48}}

	for seed := range uint64(20) {
		w := entity.NewWild(geom.Vec{X: 8, Y: 46}, zone, seeded(seed))
		w.Walls = grid
		start := w.Pos
		for i := range 60 * 60 {
			w.Update(tick, farAway)
			if cell, ok := solidCellUnder(grid, w.Box()); ok {
				t.Fatalf("seed %d, tick %d: feet box %v overlaps solid cell %v", seed, i, w.Box(), cell)
			}
			if !zone.Contains(w.Pos) {
				t.Fatalf("seed %d, tick %d: critter at %v left its zone", seed, i, w.Pos)
			}
		}
		if w.Pos == start {
			t.Errorf("seed %d: critter never got away from %v", seed, start)
		}
	}
}

// solidCellUnder returns a solid cell that box overlaps with positive area.
func solidCellUnder(g *world.Grid, box geom.Rect) ([2]int, bool) {
	// A hair of slack so a box resting flush on a cell edge does not count.
	const eps = 1e-9
	c0 := int(math.Floor((box.Min.X + eps) / g.TileSize))
	c1 := int(math.Ceil((box.Max.X-eps)/g.TileSize)) - 1
	r0 := int(math.Floor((box.Min.Y + eps) / g.TileSize))
	r1 := int(math.Ceil((box.Max.Y-eps)/g.TileSize)) - 1
	for row := r0; row <= r1; row++ {
		for col := c0; col <= c1; col++ {
			if g.Solid(col, row) {
				return [2]int{col, row}, true
			}
		}
	}
	return [2]int{}, false
}
