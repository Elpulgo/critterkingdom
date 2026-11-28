package combat

import (
	"time"

	"github.com/Elpulgo/critterkingdom/internal/geom"
)

const (
	// HopNoticeRange is how near, in world pixels, the critter must come
	// before a hop_toward enemy starts its telegraph.
	HopNoticeRange = 48.0
	// HopDuration is how long one hop lasts; the hop covers the enemy's speed
	// times this.
	HopDuration = 500 * time.Millisecond
	// HopRecover is how long a hop_toward enemy rests after landing.
	HopRecover = 600 * time.Millisecond
)

// hop is the hop_toward behaviour: when the critter is near it squashes for
// the telegraph, hops once toward where the critter was when the hop began,
// hurting with its body only while in the air, then rests.
type hop struct {
	clock     Clock
	telegraph time.Duration
	speed     float64

	started bool
	last    time.Duration // clock time of the previous Step
	phase   Phase
	start   time.Duration // clock time the current phase began
	dir     geom.Vec      // unit direction of the current hop
}

func newHop(def EnemyDef, clock Clock) *hop {
	return &hop{clock: clock, telegraph: def.Telegraph(), speed: def.Speed}
}

// Step implements Behaviour. Phase ends are exact clock times, so the
// telegraph lasts exactly the definition's time and a hop covers the same
// distance however the steps fall. The step in which the hop lands still
// reports Attack, so every hop has at least one step with a hazard.
func (h *hop) Step(self, target geom.Vec) Action {
	now := h.clock()
	if !h.started {
		h.started = true
		h.last = now
	}
	prev := h.last
	h.last = now

	for {
		switch h.phase {
		case Idle:
			if target.Sub(self).Len() > HopNoticeRange {
				return Action{Phase: Idle}
			}
			h.phase, h.start = Telegraph, now
		case Telegraph:
			end := h.start + h.telegraph
			if now < end {
				return Action{Phase: Telegraph}
			}
			h.phase, h.start = Attack, end
			h.dir = target.Sub(self).Normalized()
		case Attack:
			end := h.start + HopDuration
			move := h.dir.Scale(h.speed * span(prev, now, h.start, end))
			if now >= end {
				h.phase, h.start = Recover, end
			}
			return Action{Phase: Attack, Move: move, Hazards: []geom.Rect{bodyHazard(self)}}
		case Recover:
			if now < h.start+HopRecover {
				return Action{Phase: Recover}
			}
			h.phase = Idle
		}
	}
}
