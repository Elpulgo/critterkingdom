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

var enemyArea = geom.Rect{Max: geom.Vec{X: 320, Y: 240}}

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
		e.Update(critter, enemyArea)
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
		e.Update(critter, enemyArea)
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

	// Shake 0.8 s, then charge right at 60 px/s. By 0.9 s it wants 6 px but
	// the wall lets it go only 2, to x 10: the charge ends there and the
	// beast rests, harmless, for 1 s from 0.9 s instead of pushing on until
	// the full 72 px would have taken (2.0 s).
	for i := 0; i <= 18; i++ {
		clock.now = time.Duration(i) * 100 * time.Millisecond
		e.Update(critter, mapArea)
		if e.Pos.X > 10 {
			t.Fatalf("at %v: Pos %v, feet box past the wall at x 16", clock.now, e.Pos)
		}
		if clock.now >= 900*time.Millisecond && (e.Phase() != combat.Recover || len(e.Hazards()) != 0) {
			t.Fatalf("at %v: phase %v hazards %v, want Recover with no hazards once the wall stopped the charge", clock.now, e.Phase(), e.Hazards())
		}
	}
	if !near(e.Pos, geom.Vec{X: 10, Y: 28}) {
		t.Errorf("Pos %v, want (10, 28), flush against the wall", e.Pos)
	}
}

func TestEnemy_ChargeOnOpenGroundRunsItsFullDistance(t *testing.T) {
	clock := &fakeClock{}
	critter := geom.Vec{X: 150, Y: 100} // 50 px to the right, inside notice range
	e, _ := newEnemy(t, thistleDef(), geom.Vec{X: 100, Y: 100}, clock)
	e.Walls = loadTestGrid(t) // its one solid cell is far from the charge

	// 60 ticks a second: shake until 0.8 s, charge 72 px right until 2.0 s.
	for i := 0; i <= 120; i++ {
		clock.now = time.Duration(i) * time.Second / 60
		e.Update(critter, enemyArea)
		if clock.now > 800*time.Millisecond && e.Phase() != combat.Attack {
			t.Fatalf("at %v: phase %v at Pos %v, want the charge still running", clock.now, e.Phase(), e.Pos)
		}
	}
	if !near(e.Pos, geom.Vec{X: 172, Y: 100}) {
		t.Errorf("Pos %v, want (172, 100), 72 px right of the start", e.Pos)
	}
}

func mushroomDef() combat.EnemyDef {
	return combat.EnemyDef{ID: "grumpy_mushroom", Behaviour: combat.BehaviourPuff, HP: 1, Damage: 1, Speed: 25, TelegraphSeconds: 0.9}
}

func TestEnemy_ASleepingEnemyNeverMovesTelegraphsOrHurts(t *testing.T) {
	tests := []struct {
		name string
		def  combat.EnemyDef
	}{
		{"hop toward", slimeDef()},
		{"charge", thistleDef()},
		{"puff", mushroomDef()},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clock := &fakeClock{}
			start := geom.Vec{X: 100, Y: 100}
			critter := geom.Vec{X: 96, Y: 100} // on the enemy, a little to its left
			e, _ := newEnemy(t, tt.def, start, clock)
			if e.Asleep() {
				t.Fatal("a new enemy is asleep")
			}
			e.Sleep()
			if !e.Asleep() {
				t.Fatal("not asleep after Sleep")
			}

			// 10 s at 60 ticks a second, far past every telegraph (0.9 s at
			// most), attack and recovery.
			for i := 0; i <= 600; i++ {
				clock.now = time.Duration(i) * time.Second / 60
				e.Update(critter, enemyArea)
				if e.Pos != start || e.Phase() != combat.Idle || len(e.Hazards()) != 0 {
					t.Fatalf("at %v: Pos %v phase %v hazards %v; want still at %v, Idle, no hazards",
						clock.now, e.Pos, e.Phase(), e.Hazards(), start)
				}
			}
		})
	}
}

func TestEnemy_ASleepingEnemyCanStillBeDefeatedAndRunsOff(t *testing.T) {
	clock := &fakeClock{}
	critter := geom.Vec{X: 70, Y: 100}
	e, events := newEnemy(t, thistleDef(), geom.Vec{X: 100, Y: 100}, clock)
	e.Sleep()
	e.Update(critter, enemyArea)

	if !e.Hit(1) || e.HP() != 1 || e.Defeated() {
		t.Fatalf("first hit on the sleeping beast: HP %d, defeated %v; want the hit to land, HP 1, not defeated", e.HP(), e.Defeated())
	}
	if !e.Hit(1) || !e.Defeated() {
		t.Fatalf("second hit on the sleeping beast: defeated %v; want defeated", e.Defeated())
	}
	want := combat.EnemyDefeated{EnemyID: "thistle_beast"}
	if len(*events) != 1 || (*events)[0] != want {
		t.Errorf("defeats reported %v, want exactly [%v]", *events, want)
	}

	// Runs right, away from the critter, at 60 px/s for 1 s.
	clock.now = 500 * time.Millisecond
	e.Update(critter, enemyArea)
	if !near(e.Pos, geom.Vec{X: 130, Y: 100}) {
		t.Errorf("half a second after the defeat: Pos %v, want (130, 100)", e.Pos)
	}
	clock.now = time.Second
	e.Update(critter, enemyArea)
	if !e.Gone() {
		t.Errorf("a second after the defeat: not gone")
	}
}
