package combat_test

import (
	"testing"
	"time"

	"github.com/Elpulgo/critterkingdom/internal/combat"
	"github.com/Elpulgo/critterkingdom/internal/geom"
)

func TestCharge_ShakesThenChargesTheLinePickedAtTheShakeAndRecovers(t *testing.T) {
	clock := &fakeClock{}
	beast := newBehaviour(t, thistleDef(), clock) // speed 60, telegraph 0.8 s
	self := geom.Vec{X: 100, Y: 100}

	// Body hurt-box of the 16x16 sprite box standing at (x, 100), shrunk by
	// 3.2 on each side.
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
		// Critter 200 px away: out of the 64 px notice range.
		{0, geom.Vec{X: 300, Y: 100}, combat.Idle, geom.Vec{}, nil},
		// Critter 50 px to the right: the line is picked and the shake lasts 0.8 s.
		{time.Second, geom.Vec{X: 150, Y: 100}, combat.Telegraph, geom.Vec{}, nil},
		// The critter steps down during the shake; the line stays.
		{1400 * time.Millisecond, geom.Vec{X: 150, Y: 140}, combat.Telegraph, geom.Vec{}, nil},
		{1799 * time.Millisecond, geom.Vec{X: 150, Y: 140}, combat.Telegraph, geom.Vec{}, nil},
		// The charge starts at 1.8 s; 72 px at 60 px/s takes 1.2 s, until 3.0 s.
		{1800 * time.Millisecond, geom.Vec{X: 150, Y: 140}, combat.Attack, geom.Vec{}, []geom.Rect{body(100)}},
		{2100 * time.Millisecond, geom.Vec{X: 150, Y: 140}, combat.Attack, geom.Vec{X: 18}, []geom.Rect{body(100)}},
		// The critter moves above the beast during the charge; the line stays.
		{2400 * time.Millisecond, geom.Vec{X: 118, Y: 60}, combat.Attack, geom.Vec{X: 18}, []geom.Rect{body(118)}},
		// A late step ends the charge at exactly 72 px.
		{3000 * time.Millisecond, geom.Vec{X: 118, Y: 60}, combat.Attack, geom.Vec{X: 36}, []geom.Rect{body(136)}},
		// Rests 1 s, until 4.0 s, even with the critter beside it.
		{3100 * time.Millisecond, geom.Vec{X: 180, Y: 100}, combat.Recover, geom.Vec{}, nil},
		{3999 * time.Millisecond, geom.Vec{X: 180, Y: 100}, combat.Recover, geom.Vec{}, nil},
		// Back to waiting once the critter has walked away.
		{4000 * time.Millisecond, geom.Vec{X: 300, Y: 100}, combat.Idle, geom.Vec{}, nil},
	}

	for _, s := range steps {
		clock.now = s.at
		a := beast.Step(self, s.target)
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
	if !near(self, geom.Vec{X: 172, Y: 100}) {
		t.Errorf("beast ends at %v, want (172, 100) after one 72 px charge", self)
	}
}

func TestCharge_ACritterThatStepsSidewaysDuringTheShakeIsNeverHit(t *testing.T) {
	// The critter's hurt-box: the middle of a 16x16 box standing on its feet.
	critterHurt := func(feet geom.Vec) geom.Rect {
		return combat.HurtBox(geom.Rect{
			Min: geom.Vec{X: feet.X - 8, Y: feet.Y - 16},
			Max: geom.Vec{X: feet.X + 8, Y: feet.Y},
		})
	}

	tests := []struct {
		name    string
		stepTo  geom.Vec // where the critter stands from 0.4 s into the shake
		wantHit bool
	}{
		{"stays in the line and is run into", geom.Vec{X: 148, Y: 100}, true},
		{"steps 16 px down and is safe", geom.Vec{X: 148, Y: 116}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clock := &fakeClock{}
			beast := newBehaviour(t, thistleDef(), clock)
			self := geom.Vec{X: 100, Y: 100}
			critter := geom.Vec{X: 148, Y: 100} // 48 px right: noticed at once

			hit, charged := false, false
			// Shake 0 to 0.8 s, charge to 2.0 s, rest to 3.0 s.
			for ; clock.now <= 3*time.Second; clock.now += 10 * time.Millisecond {
				if clock.now >= 400*time.Millisecond {
					critter = tt.stepTo
				}
				a := beast.Step(self, critter)
				charged = charged || a.Phase == combat.Attack
				for _, h := range a.Hazards {
					hit = hit || combat.Overlap(h, critterHurt(critter))
				}
				self = self.Add(a.Move)
			}
			if !charged {
				t.Fatal("the beast never charged")
			}
			if hit != tt.wantHit {
				t.Errorf("critter hit = %v, want %v (beast ended at %v)", hit, tt.wantHit, self)
			}
		})
	}
}
