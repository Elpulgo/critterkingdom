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

func TestWild_StopsAndFacesAPlayerWhoComesWithinTheNoticeRadius(t *testing.T) {
	zone := geom.Rect{Min: geom.Vec{X: 0, Y: 0}, Max: geom.Vec{X: 200, Y: 200}}
	home := geom.Vec{X: 100, Y: 100}
	tests := []struct {
		name   string
		player geom.Vec
		start  entity.Facing
		want   entity.Facing
	}{
		{"player to the left", geom.Vec{X: 75, Y: 100}, entity.FacingRight, entity.FacingLeft},
		{"player to the right", geom.Vec{X: 125, Y: 100}, entity.FacingLeft, entity.FacingRight},
		{"player above", geom.Vec{X: 100, Y: 70}, entity.FacingDown, entity.FacingUp},
		{"player below", geom.Vec{X: 100, Y: 130}, entity.FacingUp, entity.FacingDown},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			w := entity.NewWild(home, zone, seeded(3))
			w.Facing = tc.start

			// Five seconds is longer than any rest, so a critter that had
			// not noticed would have hopped by the end.
			for i := range 5 * 60 {
				w.Update(tick, tc.player)
				if !w.Noticing() {
					t.Fatalf("tick %d: critter is in state %v, want noticing", i, w.State())
				}
				if w.Pos != home || w.HopHeight() != 0 {
					t.Fatalf("tick %d: noticing critter moved to %v, %v px up", i, w.Pos, w.HopHeight())
				}
			}
			if w.Facing != tc.want {
				t.Errorf("critter faces %v, want %v", w.Facing, tc.want)
			}
		})
	}
}

func TestWild_MidHopFinishesItsArcBeforeNoticingThePlayer(t *testing.T) {
	zone := geom.Rect{Min: geom.Vec{X: 0, Y: 0}, Max: geom.Vec{X: 200, Y: 200}}
	w := entity.NewWild(geom.Vec{X: 100, Y: 100}, zone, seeded(5))
	for w.State() != entity.WildHopping {
		w.Update(tick, farAway)
	}
	w.Update(tick, farAway)
	takeOff := w.Pos

	// The player steps right next to the critter while it is in the air.
	player := geom.Vec{X: 100, Y: 110}
	airborne := 0
	for w.State() == entity.WildHopping {
		w.Update(tick, player)
		airborne++
	}

	if airborne < 10 {
		t.Errorf("critter came down after %d ticks of a 21-tick hop", airborne)
	}
	if w.Pos == takeOff {
		t.Errorf("critter stopped at %v instead of finishing its hop", w.Pos)
	}
	if !w.Noticing() {
		t.Fatalf("critter landed next to the player in state %v, want noticing", w.State())
	}
	if w.HopHeight() != 0 {
		t.Errorf("noticing critter is %v px off the ground", w.HopHeight())
	}
}

func TestWild_ResumesHoppingOnlyOnceThePlayerIsBeyondTheMargin(t *testing.T) {
	zone := geom.Rect{Min: geom.Vec{X: 0, Y: 0}, Max: geom.Vec{X: 200, Y: 200}}
	home := geom.Vec{X: 100, Y: 100}
	w := entity.NewWild(home, zone, seeded(9))

	w.Update(tick, geom.Vec{X: 100 + entity.NoticeRadius - 2, Y: 100})
	if !w.Noticing() {
		t.Fatalf("critter is in state %v with the player inside the radius, want noticing", w.State())
	}

	// Between the radius and radius plus margin it keeps watching.
	between := geom.Vec{X: 100 + entity.NoticeRadius + entity.NoticeMargin/2, Y: 100}
	for i := range 10 * 60 {
		w.Update(tick, between)
		if !w.Noticing() || w.Pos != home {
			t.Fatalf("tick %d: critter is in state %v at %v with the player in the margin, want still noticing at %v",
				i, w.State(), w.Pos, home)
		}
	}

	// Well beyond the margin it goes back to resting and hopping.
	gone := geom.Vec{X: 100 + entity.NoticeRadius + entity.NoticeMargin + 1, Y: 100}
	w.Update(tick, gone)
	if w.State() != entity.WildIdle {
		t.Fatalf("critter is in state %v after the player left, want idle", w.State())
	}
	hopped := false
	for range 10 * 60 {
		w.Update(tick, farAway)
		if w.State() == entity.WildHopping {
			hopped = true
		}
	}
	if !hopped || w.Pos == home {
		t.Errorf("critter never hopped away from %v after the player left (now at %v)", home, w.Pos)
	}
}
