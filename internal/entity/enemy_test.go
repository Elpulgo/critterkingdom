package entity_test

import (
	"testing"
	"time"

	"github.com/Elpulgo/critterkingdom/internal/combat"
	"github.com/Elpulgo/critterkingdom/internal/entity"
	"github.com/Elpulgo/critterkingdom/internal/geom"
)

// fakeClock stands in for game time, the only boundary an enemy has.
type fakeClock struct{ now time.Duration }

func (c *fakeClock) read() time.Duration { return c.now }

func slimeDef() combat.EnemyDef {
	return combat.EnemyDef{ID: "sulky_slime", Behaviour: combat.BehaviourHopToward, HP: 1, Damage: 1, Speed: 30, TelegraphSeconds: 0.6}
}

func thistleDef() combat.EnemyDef {
	return combat.EnemyDef{ID: "thistle_beast", Behaviour: combat.BehaviourCharge, HP: 2, Damage: 1, Speed: 60, TelegraphSeconds: 0.8}
}

var openArea = geom.Rect{Max: geom.Vec{X: 320, Y: 240}}

// newEnemy builds an enemy whose defeats are collected in the returned slice.
func newEnemy(t *testing.T, def combat.EnemyDef, pos geom.Vec, clock *fakeClock) (*entity.Enemy, *[]combat.EnemyDefeated) {
	t.Helper()
	var events []combat.EnemyDefeated
	e, err := entity.NewEnemy(def, pos, clock.read, func(ev combat.EnemyDefeated) { events = append(events, ev) })
	if err != nil {
		t.Fatalf("NewEnemy(%q): %v", def.ID, err)
	}
	return e, &events
}

func TestEnemy_IsDefeatedByTheHitThatTakesItsHitPointsToZeroAndReportedOnce(t *testing.T) {
	tests := []struct {
		name  string
		def   combat.EnemyDef
		hpLog []int // hit points left after each hit that lands
	}{
		{"thistle beast survives one hit and falls to the second", thistleDef(), []int{1, 0}},
		{"sulky slime falls to the first hit", slimeDef(), []int{0}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e, events := newEnemy(t, tt.def, geom.Vec{X: 100, Y: 100}, &fakeClock{})
			for i, wantHP := range tt.hpLog {
				if e.Defeated() {
					t.Fatalf("defeated before hit %d", i+1)
				}
				if !e.Hit(1) {
					t.Fatalf("hit %d did not land", i+1)
				}
				if e.HP() != wantHP {
					t.Errorf("after hit %d: HP %d, want %d", i+1, e.HP(), wantHP)
				}
			}
			if !e.Defeated() {
				t.Fatalf("not defeated after %d hits", len(tt.hpLog))
			}
			if e.Hit(1) {
				t.Errorf("a hit on a defeated enemy landed")
			}
			want := []combat.EnemyDefeated{{EnemyID: tt.def.ID}}
			if len(*events) != 1 || (*events)[0] != want[0] {
				t.Errorf("defeats reported %v, want %v", *events, want)
			}
		})
	}
}

func TestEnemy_DefeatedEnemyStopsHurtingRunsAwayAndIsGoneAfterTheRunOff(t *testing.T) {
	clock := &fakeClock{}
	critter := geom.Vec{X: 70, Y: 100} // 30 px to the left, inside the slime's notice range
	e, _ := newEnemy(t, slimeDef(), geom.Vec{X: 100, Y: 100}, clock)

	// Squash for 0.6 s, then hop left at 30 px/s: 3 px by 0.7 s.
	for _, at := range []time.Duration{0, 600 * time.Millisecond, 700 * time.Millisecond} {
		clock.now = at
		e.Update(critter, openArea)
	}
	if e.Phase() != combat.Attack || len(e.Hazards()) == 0 || !near(e.Pos, geom.Vec{X: 97, Y: 100}) {
		t.Fatalf("before the hit: phase %v hazards %v Pos %v, want a hop at (97, 100) with its body as a hazard", e.Phase(), e.Hazards(), e.Pos)
	}

	e.Hit(1)
	if hz := e.Hazards(); len(hz) != 0 {
		t.Errorf("right after defeat: hazards %v, want none", hz)
	}

	// Runs right, away from the critter, at 60 px/s for 1 s from 0.7 s.
	steps := []struct {
		at       time.Duration
		wantX    float64
		wantGone bool
	}{
		{1200 * time.Millisecond, 127, false},
		{1699 * time.Millisecond, 156.94, false},
		{1700 * time.Millisecond, 157, true},
		{2500 * time.Millisecond, 157, true},
	}
	for _, s := range steps {
		clock.now = s.at
		e.Update(critter, openArea)
		if hz := e.Hazards(); len(hz) != 0 {
			t.Errorf("at %v: hazards %v, want none", s.at, hz)
		}
		if !near(e.Pos, geom.Vec{X: s.wantX, Y: 100}) {
			t.Errorf("at %v: Pos %v, want (%v, 100)", s.at, e.Pos, s.wantX)
		}
		if e.Gone() != s.wantGone {
			t.Errorf("at %v: Gone %v, want %v", s.at, e.Gone(), s.wantGone)
		}
	}
}

func TestEnemy_ChargeIntoAWallStopsAtTheWallAndRecovers(t *testing.T) {
	clock := &fakeClock{}
	// Feet box at (8, 28) is x 2..14, y 18..28, beside the solid cell at
	// x 16..32, y 16..32. The critter stands 32 px to the right, past it.
	critter := geom.Vec{X: 40, Y: 28}
	e, _ := newEnemy(t, thistleDef(), geom.Vec{X: 8, Y: 28}, clock)
	e.Walls = loadTestGrid(t)

	// Shake 0.8 s, charge 72 px at 60 px/s for 1.2 s until 2.0 s, then rest.
	for i := 0; i <= 25; i++ {
		clock.now = time.Duration(i) * 100 * time.Millisecond
		e.Update(critter, mapArea)
		if e.Pos.X > 10 {
			t.Fatalf("at %v: Pos %v, feet box past the wall at x 16", clock.now, e.Pos)
		}
		if clock.now == 1900*time.Millisecond && e.Phase() != combat.Attack {
			t.Fatalf("at 1.9 s: phase %v, want still pushing against the wall in Attack", e.Phase())
		}
	}
	if e.Phase() != combat.Recover {
		t.Errorf("at 2.5 s: phase %v, want Recover", e.Phase())
	}
	if !near(e.Pos, geom.Vec{X: 10, Y: 28}) {
		t.Errorf("Pos %v, want (10, 28), flush against the wall", e.Pos)
	}
}
