package entity

import (
	"time"

	"github.com/Elpulgo/critterkingdom/internal/combat"
	"github.com/Elpulgo/critterkingdom/internal/geom"
)

const (
	// RunOffSpeed is how fast, in world pixels per second, a defeated enemy
	// runs away from the critter.
	RunOffSpeed = 60.0
	// RunOffTime is how long a defeated enemy runs off, smiling, before it is
	// gone.
	RunOffTime = time.Second
	// blockSlack is how much shorter, in world pixels, a move may come out
	// than wanted from float rounding alone before it counts as blocked.
	blockSlack = 1e-6
)

// Enemy is one enemy in the world: where it stands, its hit points, what its
// behaviour does each step, and the defeat and run-off once its hit points
// are gone.
type Enemy struct {
	// Pos is where the enemy's feet touch the ground, in world pixels.
	Pos geom.Vec
	// Walls, when non-nil, constrains movement against solid scenery.
	Walls Walls

	def       combat.EnemyDef
	behaviour combat.Behaviour
	clock     combat.Clock
	report    func(combat.EnemyDefeated)

	hp       int
	action   combat.Action // what the enemy did in the last Update
	defeated bool
	asleep   bool
	runFrom  time.Duration // clock time up to which the run-off has been moved
	goneAt   time.Duration // clock time the run-off ends
}

// NewEnemy returns an enemy of kind def standing at pos with full hit points,
// reading game time from clock. report, when non-nil, is called once with
// the defeat. It returns an error when def's behaviour cannot be built.
func NewEnemy(def combat.EnemyDef, pos geom.Vec, clock combat.Clock, report func(combat.EnemyDefeated)) (*Enemy, error) {
	b, err := combat.NewBehaviour(def, clock)
	if err != nil {
		return nil, err
	}
	return &Enemy{Pos: pos, def: def, behaviour: b, clock: clock, report: report, hp: def.HP}, nil
}

// Def returns the enemy's definition.
func (e *Enemy) Def() combat.EnemyDef {
	return e.def
}

// HP returns the enemy's hit points left.
func (e *Enemy) HP() int {
	return e.hp
}

// Sleep puts the enemy to sleep for good: from now on it stands still, Idle,
// with no hazards, however close the critter comes. It can still be hit and
// defeated, and then runs off like an awake enemy.
func (e *Enemy) Sleep() {
	e.asleep = true
	e.action = combat.Action{}
}

// Asleep reports whether the enemy has been put to sleep and not yet
// defeated: a defeated enemy runs off awake.
func (e *Enemy) Asleep() bool {
	return e.asleep && !e.defeated
}

// Phase returns the phase of the enemy's behaviour in the last Update. A
// defeated or sleeping enemy is Idle.
func (e *Enemy) Phase() combat.Phase {
	return e.action.Phase
}

// Hazards returns the boxes that hurt the critter after the last Update. A
// defeated or sleeping enemy has none.
func (e *Enemy) Hazards() []geom.Rect {
	return e.action.Hazards
}

// Box returns the enemy's 16 by 16 sprite box, with its bottom centre at Pos.
func (e *Enemy) Box() geom.Rect {
	return combat.EnemyBox(e.Pos)
}

// HurtBox returns the part of the enemy that a swing must touch to hit it.
func (e *Enemy) HurtBox() geom.Rect {
	return combat.HurtBox(e.Box())
}

// feetBox is the enemy's 12 by 10 collision box against walls, with its
// bottom centre at Pos, the same size as the critter's.
func (e *Enemy) feetBox() geom.Rect {
	return geom.Rect{
		Min: geom.Vec{X: e.Pos.X - boxHalfWidth, Y: e.Pos.Y - boxHeight},
		Max: geom.Vec{X: e.Pos.X + boxHalfWidth, Y: e.Pos.Y},
	}
}

// Defeated reports whether the enemy's hit points have run out.
func (e *Enemy) Defeated() bool {
	return e.defeated
}

// Gone reports whether a defeated enemy has finished running off and should
// leave the world.
func (e *Enemy) Gone() bool {
	return e.defeated && e.clock() >= e.goneAt
}

// Hit takes damage off the enemy's hit points and returns true. The hit that
// takes them to zero defeats the enemy and reports the defeat. It does
// nothing and returns false when damage is not positive or the enemy is
// already defeated.
func (e *Enemy) Hit(damage int) bool {
	if damage <= 0 || e.defeated {
		return false
	}
	e.hp -= damage
	if e.hp > 0 {
		return true
	}
	now := e.clock()
	e.hp = 0
	e.defeated = true
	e.action = combat.Action{}
	e.runFrom, e.goneAt = now, now+RunOffTime
	if e.report != nil {
		e.report(combat.EnemyDefeated{EnemyID: e.def.ID})
	}
	return true
}

// Update moves the enemy one step toward or away from the critter at target.
// A live enemy steps its behaviour; a defeated one runs straight away from
// target at RunOffSpeed until RunOffTime has passed since the defeat. A
// sleeping enemy that is not defeated does not step its behaviour and stays
// where it is. Otherwise the move slides along Walls and the enemy's feet never leave area. When
// a wall or area cuts a live enemy's move short and its behaviour is
// combat.Blockable, the behaviour is told, so a charge stops at the wall and
// recovers.
func (e *Enemy) Update(target geom.Vec, area geom.Rect) {
	if e.Asleep() {
		return
	}
	var move geom.Vec
	if e.defeated {
		move = e.runOff(target)
	} else {
		e.action = e.behaviour.Step(e.Pos, target)
		move = e.action.Move
	}
	next := area.Clamp(e.Pos.Add(move))
	if e.Walls != nil {
		next = e.Pos.Add(e.Walls.Slide(e.feetBox(), next.Sub(e.Pos)))
	}
	if b, ok := e.behaviour.(combat.Blockable); ok && !e.defeated && cutShort(move, next.Sub(e.Pos)) {
		e.action = b.Blocked()
	}
	e.Pos = next
}

// cutShort reports whether the move made is shorter than the move wanted, by
// more than float rounding.
func cutShort(wanted, made geom.Vec) bool {
	return made.Len() < wanted.Len()-blockSlack
}

// runOff returns the run-off move for the part of the run-off time since the
// previous Update, straight away from target. A critter standing exactly on
// the enemy sends it to the right.
func (e *Enemy) runOff(target geom.Vec) geom.Vec {
	to := min(e.clock(), e.goneAt)
	if to <= e.runFrom {
		return geom.Vec{}
	}
	dist := RunOffSpeed * (to - e.runFrom).Seconds()
	e.runFrom = to
	away := e.Pos.Sub(target).Normalized()
	if away.IsZero() {
		away = geom.Vec{X: 1}
	}
	return away.Scale(dist)
}
