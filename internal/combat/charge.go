package combat

import (
	"time"

	"github.com/Elpulgo/critterkingdom/internal/geom"
)

const (
	// ChargeNoticeRange is how near, in world pixels, the critter must come
	// before a charge enemy starts shaking.
	ChargeNoticeRange = 64.0
	// ChargeDistance is how far, in world pixels, one charge runs in its
	// straight line, at the enemy's speed.
	ChargeDistance = 72.0
	// ChargeRecover is how long a charge enemy stands, puffed out, after a
	// charge.
	ChargeRecover = time.Second
)

// charge is the charge behaviour: when the critter is near it picks the line
// to the critter and shakes for the telegraph, then runs ChargeDistance along
// that line, hurting with its body only while running, then rests. The line
// never changes once picked, so a child who steps sideways during the shake
// is safe.
type charge struct {
	clock     Clock
	telegraph time.Duration
	speed     float64
	run       time.Duration // how long a charge of ChargeDistance takes

	phase   Phase
	start   time.Duration // clock time the current phase began
	dir     geom.Vec      // unit direction of the current charge
	covered float64       // distance run so far in the current charge
}

// newCharge returns a charge behaviour for def; def.Speed must be positive.
func newCharge(def EnemyDef, clock Clock) *charge {
	run := time.Duration(ChargeDistance / def.Speed * float64(time.Second))
	return &charge{clock: clock, telegraph: def.Telegraph(), speed: def.Speed, run: run}
}

// Step implements Behaviour. Phase ends are exact clock times, so the
// telegraph lasts exactly the definition's time, and a charge always covers
// exactly ChargeDistance however the steps fall: each step moves the beast
// from the distance it had run to the distance it should have run by now.
// The step in which the charge ends still reports Attack, so every charge
// has at least one step with a hazard.
func (c *charge) Step(self, target geom.Vec) Action {
	now := c.clock()

	for {
		switch c.phase {
		case Idle:
			if target.Sub(self).Len() > ChargeNoticeRange {
				return Action{Phase: Idle}
			}
			c.phase, c.start = Telegraph, now
			c.dir = target.Sub(self).Normalized()
		case Telegraph:
			end := c.start + c.telegraph
			if now < end {
				return Action{Phase: Telegraph}
			}
			c.phase, c.start, c.covered = Attack, end, 0
		case Attack:
			end := c.start + c.run
			total := ChargeDistance
			if now < end {
				total = c.speed * (now - c.start).Seconds()
			}
			move := c.dir.Scale(total - c.covered)
			c.covered = total
			if now >= end {
				c.phase, c.start = Recover, end
			}
			return Action{Phase: Attack, Move: move, Hazards: []geom.Rect{bodyHazard(self)}}
		case Recover:
			if now < c.start+ChargeRecover {
				return Action{Phase: Recover}
			}
			c.phase = Idle
		}
	}
}
