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

func TestPlayer_PlacingTheCritterStopsItsWalk(t *testing.T) {
	p := NewPlayer(geom.Vec{X: 10, Y: 50})
	p.WalkTo(geom.Vec{X: 90, Y: 50})

	p.Place(geom.Vec{X: 40, Y: 20})
	p.Update(tick, geom.Vec{}, field)

	if _, walking := p.Target(); walking || p.Pos != (geom.Vec{X: 40, Y: 20}) {
		t.Errorf("at %v, walking %v; want standing still on 40,20", p.Pos, walking)
	}
}

func TestPlayer_FollowsARouteThroughEachPointInOrderAndStopsOnTheLast(t *testing.T) {
	a, b, c := geom.Vec{X: 12, Y: 50}, geom.Vec{X: 12, Y: 47}, geom.Vec{X: 14.2, Y: 47}
	p := NewPlayer(geom.Vec{X: 10, Y: 50})
	p.WalkAlong([]geom.Vec{a, b, c})

	// 1.2 px a tick; a tick that would overshoot a point stops on it, and the
	// next tick heads for the following point.
	want := []geom.Vec{
		{X: 11.2, Y: 50},
		a,
		{X: 12, Y: 48.8},
		{X: 12, Y: 47.6},
		b,
		{X: 13.2, Y: 47},
		c,
	}
	for i, w := range want {
		p.Update(tick, geom.Vec{}, field)
		if !near(p.Pos, w) {
			t.Fatalf("tick %d: at %v, want %v", i+1, p.Pos, w)
		}
		target, walking := p.Target()
		if last := i == len(want)-1; walking == last {
			t.Fatalf("tick %d: walking %v, want %v", i+1, walking, !last)
		}
		if walking && target != c {
			t.Fatalf("tick %d: target %v, want the route's last point %v", i+1, target, c)
		}
	}
	if p.Pos != c {
		t.Fatalf("stopped at %v, want exactly %v", p.Pos, c)
	}
	p.Update(tick, geom.Vec{}, field)
	if p.Pos != c || p.Moved() {
		t.Errorf("after arriving: at %v, moved %v; want standing still on %v", p.Pos, p.Moved(), c)
	}
}

func TestPlayer_InterruptingARouteDropsAllItsRemainingPoints(t *testing.T) {
	route := []geom.Vec{{X: 20, Y: 50}, {X: 20, Y: 80}, {X: 60, Y: 80}}
	tests := []struct {
		name      string
		interrupt func(p *Player)
		want      geom.Vec
	}{
		{
			name:      "a direction input walks that way",
			interrupt: func(p *Player) { p.Update(tick, geom.Vec{Y: -1}, field) },
			want:      geom.Vec{X: 12.4, Y: 48.8},
		},
		{
			name:      "an empty route stops the walk",
			interrupt: func(p *Player) { p.WalkAlong(nil) },
			want:      geom.Vec{X: 12.4, Y: 50},
		},
		{
			name:      "placing the critter stops the walk",
			interrupt: func(p *Player) { p.Place(geom.Vec{X: 40, Y: 20}) },
			want:      geom.Vec{X: 40, Y: 20},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			p := NewPlayer(geom.Vec{X: 10, Y: 50})
			p.WalkAlong(route)
			p.Update(tick, geom.Vec{}, field)
			p.Update(tick, geom.Vec{}, field)

			tc.interrupt(p)
			for range 60 {
				p.Update(tick, geom.Vec{}, field)
			}

			if _, walking := p.Target(); walking || !near(p.Pos, tc.want) {
				t.Errorf("at %v, walking %v; want standing still on %v, not resuming the route", p.Pos, walking, tc.want)
			}
		})
	}
}
