package combat_test

import (
	"testing"
	"time"

	"github.com/Elpulgo/critterkingdom/internal/combat"
	"github.com/Elpulgo/critterkingdom/internal/geom"
)

func TestPuff_StandsStillPuffsFourSporesThatVanishAtTheirRangeAndRecovers(t *testing.T) {
	clock := &fakeClock{}
	mushroom := newBehaviour(t, mushroomDef(), clock) // spore speed 25, telegraph 0.9 s
	self := geom.Vec{X: 100, Y: 100}                  // sprite box centre (100, 92)

	// A 10x10 spore box centred on (x, y).
	spore := func(x, y float64) geom.Rect {
		return geom.Rect{Min: geom.Vec{X: x - 5, Y: y - 5}, Max: geom.Vec{X: x + 5, Y: y + 5}}
	}
	// The four spores in the order up, down, left, right, d px from (100, 92).
	four := func(d float64) []geom.Rect {
		return []geom.Rect{spore(100, 92-d), spore(100, 92+d), spore(100-d, 92), spore(100+d, 92)}
	}

	steps := []struct {
		at        time.Duration
		target    geom.Vec
		wantPhase combat.Phase
		wantHurt  []geom.Rect
	}{
		// Critter 200 px away: out of the 56 px notice range.
		{0, geom.Vec{X: 300, Y: 100}, combat.Idle, nil},
		// Critter 30 px to the right: it puffs up for 0.9 s, until 1.9 s.
		{time.Second, geom.Vec{X: 130, Y: 100}, combat.Telegraph, nil},
		// The critter presses against it during the puff-up: nothing hurts.
		{1899 * time.Millisecond, geom.Vec{X: 100, Y: 100}, combat.Telegraph, nil},
		// The spores leave its centre at 1.9 s.
		{1900 * time.Millisecond, geom.Vec{X: 100, Y: 100}, combat.Attack, four(0)},
		// 0.8 s later at 25 px/s they are 20 px out, clear of its body.
		{2700 * time.Millisecond, geom.Vec{X: 100, Y: 100}, combat.Attack, four(20)},
		// 48 px takes 1.92 s: at 3.82 s they are at their range.
		{3820 * time.Millisecond, geom.Vec{X: 100, Y: 100}, combat.Attack, four(48)},
		// Then they are gone and it rests 1.2 s, until 5.02 s, critter or not.
		{3830 * time.Millisecond, geom.Vec{X: 100, Y: 100}, combat.Recover, nil},
		{5019 * time.Millisecond, geom.Vec{X: 130, Y: 100}, combat.Recover, nil},
		// Back to waiting once the critter has walked away.
		{5020 * time.Millisecond, geom.Vec{X: 300, Y: 100}, combat.Idle, nil},
	}

	for _, s := range steps {
		clock.now = s.at
		a := mushroom.Step(self, s.target)
		if a.Phase != s.wantPhase {
			t.Fatalf("at %v: phase %v, want %v", s.at, a.Phase, s.wantPhase)
		}
		if !a.Move.IsZero() {
			t.Errorf("at %v: move %v, want none: the mushroom never moves", s.at, a.Move)
		}
		// Only the spores hurt: the mushroom's body is never a fifth hazard.
		if len(a.Hazards) != len(s.wantHurt) {
			t.Fatalf("at %v: hazards %v, want %v", s.at, a.Hazards, s.wantHurt)
		}
		for i := range a.Hazards {
			if !near(a.Hazards[i].Min, s.wantHurt[i].Min) || !near(a.Hazards[i].Max, s.wantHurt[i].Max) {
				t.Errorf("at %v: hazard %d = %v, want %v", s.at, i, a.Hazards[i], s.wantHurt[i])
			}
		}
	}
}
