package combat

import (
	"fmt"
	"time"

	"github.com/Elpulgo/critterkingdom/internal/geom"
)

// EnemySize is the width and height, in world pixels, of an enemy's sprite
// box. The box's bottom centre is the enemy's position.
const EnemySize = 16.0

// Phase is where an enemy is in its attack cycle.
type Phase int

// The phases every behaviour moves through, in order.
const (
	// Idle waits in place until the critter comes near.
	Idle Phase = iota
	// Telegraph is the wind-up the child can see; nothing hurts during it.
	Telegraph
	// Attack is the only phase in which anything can hurt the critter.
	Attack
	// Recover is the pause after an attack, the moment to hit back.
	Recover
)

// String returns the phase's name.
func (p Phase) String() string {
	switch p {
	case Idle:
		return "Idle"
	case Telegraph:
		return "Telegraph"
	case Attack:
		return "Attack"
	case Recover:
		return "Recover"
	}
	return fmt.Sprintf("Phase(%d)", int(p))
}

// Action is what an enemy does during one step.
type Action struct {
	// Phase is the phase the enemy is in for this step.
	Phase Phase
	// Move is the displacement, in world pixels, the enemy wants this step.
	Move geom.Vec
	// Hazards are the boxes that hurt the critter this step. They are only
	// ever non-empty in the Attack phase.
	Hazards []geom.Rect
}

// Behaviour is an enemy's state machine. Step reads the behaviour's clock,
// advances by the time since the previous Step and returns what the enemy at
// self does this step toward the critter at target. Both positions are
// bottom-centre points in world pixels.
type Behaviour interface {
	Step(self, target geom.Vec) Action
}

// Blockable is implemented by behaviours whose attack ends when scenery or the
// map's edge cuts its move short, such as a charge that runs into a wall.
// Whoever moves the enemy calls Blocked right after a Step whose Move it could
// not make in full.
type Blockable interface {
	// Blocked tells the behaviour that the Move of the last Step was cut
	// short, and returns what the enemy does for the rest of that step.
	Blocked() Action
}

// NewBehaviour returns the state machine def.Behaviour names, reading game
// time from clock. It returns an error for a behaviour name it does not know,
// for a telegraph shorter than MinTelegraph, and for a charge or puff without
// a positive speed (for puff, the speed of its spores).
func NewBehaviour(def EnemyDef, clock Clock) (Behaviour, error) {
	if def.Telegraph() < MinTelegraph {
		return nil, fmt.Errorf("enemy %q: telegraph %v is below the minimum %v", def.ID, def.Telegraph(), MinTelegraph)
	}
	switch def.Behaviour {
	case BehaviourHopToward:
		return newHop(def, clock), nil
	case BehaviourCharge:
		if def.Speed <= 0 {
			return nil, fmt.Errorf("enemy %q: charge needs a positive speed, got %v", def.ID, def.Speed)
		}
		return newCharge(def, clock), nil
	case BehaviourPuff:
		if def.Speed <= 0 {
			return nil, fmt.Errorf("enemy %q: puff needs a positive spore speed, got %v", def.ID, def.Speed)
		}
		return newPuff(def, clock), nil
	}
	return nil, fmt.Errorf("enemy %q: unknown behaviour %q", def.ID, def.Behaviour)
}

// EnemyBox returns the sprite box of an enemy standing at pos: EnemySize
// square with pos at its bottom centre.
func EnemyBox(pos geom.Vec) geom.Rect {
	return geom.Rect{
		Min: geom.Vec{X: pos.X - EnemySize/2, Y: pos.Y - EnemySize},
		Max: geom.Vec{X: pos.X + EnemySize/2, Y: pos.Y},
	}
}

// bodyHazard is the box an enemy's body hurts with: the hurt-box of its
// sprite box at pos.
func bodyHazard(pos geom.Vec) geom.Rect {
	return HurtBox(EnemyBox(pos))
}

// span returns how much of the step (prev, now] falls inside the phase
// [start, end), as seconds.
func span(prev, now, start, end time.Duration) float64 {
	from, to := max(prev, start), min(now, end)
	if to <= from {
		return 0
	}
	return (to - from).Seconds()
}

// EnemyDefeated is reported once when an enemy's hit points reach zero.
type EnemyDefeated struct {
	// EnemyID is the id of the defeated enemy's definition, such as
	// "sulky_slime".
	EnemyID string
}
