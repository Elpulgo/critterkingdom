package entity

import (
	"math"
	"testing"

	"github.com/Elpulgo/critterkingdom/internal/geom"
)

// tick is one tick at 60 ticks per second. The critter walks 1.2 pixels in
// one tick.
const tick = 1.0 / 60

var field = geom.Rect{Max: geom.Vec{X: 100, Y: 100}}

func TestPlayer_WalksToTheTargetAndStopsExactlyOnIt(t *testing.T) {
	p := NewPlayer(geom.Vec{X: 10, Y: 50})
	p.WalkTo(geom.Vec{X: 20, Y: 50})

	for range 8 {
		p.Update(tick, geom.Vec{}, field)
	}
	if _, walking := p.Target(); !walking || math.Abs(p.Pos.X-19.6) > 1e-9 {
		t.Fatalf("after 8 ticks: at %v, walking %v; want x 19.6 and still walking", p.Pos, walking)
	}

	p.Update(tick, geom.Vec{}, field)
	if _, walking := p.Target(); walking || p.Pos != (geom.Vec{X: 20, Y: 50}) {
		t.Fatalf("after 9 ticks: at %v, walking %v; want exactly 20,50 and stopped", p.Pos, walking)
	}
}

func TestPlayer_DirectionInputCancelsTheWalkTarget(t *testing.T) {
	p := NewPlayer(geom.Vec{X: 10, Y: 50})
	p.WalkTo(geom.Vec{X: 90, Y: 50})

	p.Update(tick, geom.Vec{Y: -1}, field)
	p.Update(tick, geom.Vec{}, field)

	if want := (geom.Vec{X: 10, Y: 48.8}); !near(p.Pos, want) {
		t.Errorf("at %v, want %v: one step up, then standing still", p.Pos, want)
	}
}

func TestPlayer_NeverLeavesTheWalkableArea(t *testing.T) {
	t.Run("walking into the edge", func(t *testing.T) {
		p := NewPlayer(geom.Vec{X: 99.5, Y: 50})
		p.Update(tick, geom.Vec{X: 1}, field)
		if p.Pos != (geom.Vec{X: 100, Y: 50}) {
			t.Errorf("at %v, want 100,50", p.Pos)
		}
	})
	t.Run("walking to a target outside the area", func(t *testing.T) {
		p := NewPlayer(geom.Vec{X: 95, Y: 50})
		p.WalkTo(geom.Vec{X: 150, Y: 50})
		for range 10 {
			p.Update(tick, geom.Vec{}, field)
		}
		if _, walking := p.Target(); walking || p.Pos != (geom.Vec{X: 100, Y: 50}) {
			t.Errorf("at %v, walking %v; want stopped at the edge, 100,50", p.Pos, walking)
		}
	})
}

func near(a, b geom.Vec) bool {
	return math.Abs(a.X-b.X) < 1e-9 && math.Abs(a.Y-b.Y) < 1e-9
}
