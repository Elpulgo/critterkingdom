package combat_test

import (
	"math"
	"strings"
	"testing"
	"time"

	"github.com/Elpulgo/critterkingdom/internal/combat"
	"github.com/Elpulgo/critterkingdom/internal/geom"
)

// fakeClock stands in for game time, the only boundary a behaviour has.
type fakeClock struct{ now time.Duration }

func (c *fakeClock) read() time.Duration { return c.now }

func slimeDef() combat.EnemyDef {
	return combat.EnemyDef{
		ID:               "sulky_slime",
		Behaviour:        combat.BehaviourHopToward,
		HP:               1,
		Damage:           1,
		Speed:            30,
		TelegraphSeconds: 0.6,
	}
}

func newBehaviour(t *testing.T, def combat.EnemyDef, clock *fakeClock) combat.Behaviour {
	t.Helper()
	b, err := combat.NewBehaviour(def, clock.read)
	if err != nil {
		t.Fatalf("NewBehaviour(%q) error: %v", def.Behaviour, err)
	}
	return b
}

func near(a, b geom.Vec) bool {
	return math.Abs(a.X-b.X) < 1e-9 && math.Abs(a.Y-b.Y) < 1e-9
}

func TestNewBehaviour_RejectsAnUnknownBehaviourAndATooShortTelegraph(t *testing.T) {
	tests := []struct {
		name    string
		edit    func(d *combat.EnemyDef)
		wantErr string
	}{
		{"unknown behaviour name", func(d *combat.EnemyDef) { d.Behaviour = "wobble" }, `unknown behaviour "wobble"`},
		{"telegraph under 0.6 s", func(d *combat.EnemyDef) { d.TelegraphSeconds = 0.59 }, "telegraph"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			def := slimeDef()
			tt.edit(&def)
			_, err := combat.NewBehaviour(def, (&fakeClock{}).read)
			if err == nil {
				t.Fatalf("NewBehaviour error = nil, want one containing %q", tt.wantErr)
			}
			if !strings.Contains(err.Error(), tt.wantErr) || !strings.Contains(err.Error(), "sulky_slime") {
				t.Errorf("NewBehaviour error = %q, want it to name sulky_slime and contain %q", err, tt.wantErr)
			}
		})
	}
}

func TestBehaviour_NoHazardBeforeTheTelegraphHasRun(t *testing.T) {
	defs := []combat.EnemyDef{slimeDef()}
	self := geom.Vec{X: 100, Y: 100}
	target := geom.Vec{X: 120, Y: 100} // inside every notice range
	noticed := 5 * time.Second

	for _, def := range defs {
		telegraph := def.Telegraph()

		t.Run(def.Behaviour+" stepped every millisecond", func(t *testing.T) {
			clock := &fakeClock{now: noticed}
			b := newBehaviour(t, def, clock)
			for ; clock.now < noticed+telegraph; clock.now += time.Millisecond {
				if a := b.Step(self, target); len(a.Hazards) != 0 || a.Phase != combat.Telegraph {
					t.Fatalf("at +%v: phase %v with %d hazards, want Telegraph with none", clock.now-noticed, a.Phase, len(a.Hazards))
				}
			}
			if a := b.Step(self, target); a.Phase != combat.Attack || len(a.Hazards) == 0 {
				t.Errorf("at +%v: phase %v with %d hazards, want Attack with hazards", telegraph, a.Phase, len(a.Hazards))
			}
		})

		t.Run(def.Behaviour+" one step a nanosecond short", func(t *testing.T) {
			clock := &fakeClock{now: noticed}
			b := newBehaviour(t, def, clock)
			b.Step(self, target)
			clock.now = noticed + telegraph - time.Nanosecond
			if a := b.Step(self, target); len(a.Hazards) != 0 || a.Phase != combat.Telegraph {
				t.Errorf("at +%v: phase %v with %d hazards, want Telegraph with none", clock.now-noticed, a.Phase, len(a.Hazards))
			}
		})
	}
}
