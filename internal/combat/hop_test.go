package combat_test

import (
	"testing"
	"time"

	"github.com/Elpulgo/critterkingdom/internal/combat"
	"github.com/Elpulgo/critterkingdom/internal/geom"
)

func TestHopToward_IdlesUntilNearThenTelegraphsHopsAtTheCritterAndRecovers(t *testing.T) {
	clock := &fakeClock{}
	slime := newBehaviour(t, slimeDef(), clock) // speed 30, telegraph 0.6 s
	self := geom.Vec{X: 100, Y: 100}

	// Body hurt-box of the 16x16 sprite box standing at (100, 100), shrunk by
	// 3.2 on each side, then the same box after each hop step.
	body := func(x float64) geom.Rect {
		return geom.Rect{Min: geom.Vec{X: x - 4.8, Y: 87.2}, Max: geom.Vec{X: x + 4.8, Y: 96.8}}
	}

	steps := []struct {
		at        time.Duration
		target    geom.Vec
		wantPhase combat.Phase
		wantMove  geom.Vec
		wantHurt  []geom.Rect
	}{
		// Critter 200 px away: out of the 48 px notice range.
		{0, geom.Vec{X: 300, Y: 100}, combat.Idle, geom.Vec{}, nil},
		{time.Second, geom.Vec{X: 300, Y: 100}, combat.Idle, geom.Vec{}, nil},
		// Critter 30 px to the right: the squash starts and lasts 0.6 s.
		{time.Second, geom.Vec{X: 130, Y: 100}, combat.Telegraph, geom.Vec{}, nil},
		{1300 * time.Millisecond, geom.Vec{X: 130, Y: 100}, combat.Telegraph, geom.Vec{}, nil},
		{1599 * time.Millisecond, geom.Vec{X: 130, Y: 100}, combat.Telegraph, geom.Vec{}, nil},
		// The hop starts at 1.6 s toward (130, 100); 30 px/s for 0.5 s is 15 px.
		{1600 * time.Millisecond, geom.Vec{X: 130, Y: 100}, combat.Attack, geom.Vec{}, []geom.Rect{body(100)}},
		{1700 * time.Millisecond, geom.Vec{X: 130, Y: 100}, combat.Attack, geom.Vec{X: 3}, []geom.Rect{body(100)}},
		// The critter steps down; the hop keeps its line.
		{1900 * time.Millisecond, geom.Vec{X: 103, Y: 140}, combat.Attack, geom.Vec{X: 6}, []geom.Rect{body(103)}},
		// Lands at 2.1 s after 15 px in total.
		{2100 * time.Millisecond, geom.Vec{X: 103, Y: 140}, combat.Attack, geom.Vec{X: 6}, []geom.Rect{body(109)}},
		// Rests 0.6 s, until 2.7 s, even with the critter beside it.
		{2200 * time.Millisecond, geom.Vec{X: 120, Y: 100}, combat.Recover, geom.Vec{}, nil},
		{2699 * time.Millisecond, geom.Vec{X: 120, Y: 100}, combat.Recover, geom.Vec{}, nil},
		// Back to waiting once the critter has walked away.
		{2700 * time.Millisecond, geom.Vec{X: 300, Y: 100}, combat.Idle, geom.Vec{}, nil},
	}

	for _, s := range steps {
		clock.now = s.at
		a := slime.Step(self, s.target)
		if a.Phase != s.wantPhase {
			t.Fatalf("at %v: phase %v, want %v", s.at, a.Phase, s.wantPhase)
		}
		if !near(a.Move, s.wantMove) {
			t.Errorf("at %v: move %v, want %v", s.at, a.Move, s.wantMove)
		}
		if len(a.Hazards) != len(s.wantHurt) {
			t.Fatalf("at %v: hazards %v, want %v", s.at, a.Hazards, s.wantHurt)
		}
		for i := range a.Hazards {
			if !near(a.Hazards[i].Min, s.wantHurt[i].Min) || !near(a.Hazards[i].Max, s.wantHurt[i].Max) {
				t.Errorf("at %v: hazard %d = %v, want %v", s.at, i, a.Hazards[i], s.wantHurt[i])
			}
		}
		self = self.Add(a.Move)
	}
	if !near(self, geom.Vec{X: 115, Y: 100}) {
		t.Errorf("slime ends at %v, want (115, 100) after one 15 px hop", self)
	}
}
