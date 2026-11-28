package combat

import (
	"time"

	"github.com/Elpulgo/critterkingdom/internal/geom"
)

const (
	// PuffNoticeRange is how near, in world pixels, the critter must come
	// before a puff enemy starts puffing up.
	PuffNoticeRange = 56.0
	// PuffSporeRange is how far, in world pixels, a spore's centre travels
	// from the centre of the enemy's sprite box before it vanishes.
	PuffSporeRange = 48.0
	// PuffSporeSize is the width and height, in world pixels, of the box a
	// spore hurts with.
	PuffSporeSize = 10.0
	// PuffRecover is how long a puff enemy rests after its spores vanish:
	// the moment to hit it.
	PuffRecover = 1200 * time.Millisecond
)

// puffDirs are the four directions the spores travel, in the order their
// hazards are reported: up, down, left, right.
var puffDirs = [4]geom.Vec{{X: 0, Y: -1}, {X: 0, Y: 1}, {X: -1, Y: 0}, {X: 1, Y: 0}}

// puff is the puff behaviour: it never moves. When the critter is near it
// puffs up for the telegraph, then sends four spores up, down, left and right
// from its centre at the enemy's speed until they reach PuffSporeRange and
// vanish, then rests. Only the spores hurt, never the enemy's body.
type puff struct {
	clock     Clock
	telegraph time.Duration
	speed     float64
	travel    time.Duration // how long a spore takes to reach PuffSporeRange

	phase Phase
	start time.Duration // clock time the current phase began
}

// newPuff returns a puff behaviour for def; def.Speed must be positive.
func newPuff(def EnemyDef, clock Clock) *puff {
	travel := time.Duration(PuffSporeRange / def.Speed * float64(time.Second))
	return &puff{clock: clock, telegraph: def.Telegraph(), speed: def.Speed, travel: travel}
}

// Step implements Behaviour. Phase ends are exact clock times, so the
// telegraph lasts exactly the definition's time, and the spores are placed
// from the time since they left, so they travel the same way however the
// steps fall. The step in which the spores reach their range still reports
// Attack with them at the range, so every puff has at least one step with
// hazards.
func (p *puff) Step(self, target geom.Vec) Action {
	now := p.clock()

	for {
		switch p.phase {
		case Idle:
			if target.Sub(self).Len() > PuffNoticeRange {
				return Action{Phase: Idle}
			}
			p.phase, p.start = Telegraph, now
		case Telegraph:
			end := p.start + p.telegraph
			if now < end {
				return Action{Phase: Telegraph}
			}
			p.phase, p.start = Attack, end
		case Attack:
			end := p.start + p.travel
			dist := PuffSporeRange
			if now < end {
				dist = p.speed * (now - p.start).Seconds()
			}
			if now >= end {
				p.phase, p.start = Recover, end
			}
			return Action{Phase: Attack, Hazards: spores(self, dist)}
		case Recover:
			if now < p.start+PuffRecover {
				return Action{Phase: Recover}
			}
			p.phase = Idle
		}
	}
}

// spores returns the four spore boxes of an enemy standing at pos whose
// spores have travelled dist from the centre of its sprite box, in puffDirs
// order.
func spores(pos geom.Vec, dist float64) []geom.Rect {
	centre := geom.Vec{X: pos.X, Y: pos.Y - EnemySize/2}
	const h = PuffSporeSize / 2
	boxes := make([]geom.Rect, 0, len(puffDirs))
	for _, d := range puffDirs {
		c := centre.Add(d.Scale(dist))
		boxes = append(boxes, geom.Rect{
			Min: geom.Vec{X: c.X - h, Y: c.Y - h},
			Max: geom.Vec{X: c.X + h, Y: c.Y + h},
		})
	}
	return boxes
}
