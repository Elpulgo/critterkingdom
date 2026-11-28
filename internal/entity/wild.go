package entity

import (
	"math"
	"math/rand/v2"

	"github.com/Elpulgo/critterkingdom/internal/geom"
)

const (
	// WildHopDistance is the furthest a wild critter goes in one hop, in
	// world pixels.
	WildHopDistance = 12.0
	// wildHopMinDistance is the shortest hop a wild critter aims for, so a
	// hop always shows as a move.
	wildHopMinDistance = 4.0
	// wildIdleMin and wildIdleMax bound how long a wild critter stands still
	// between hops, in seconds.
	wildIdleMin = 0.8
	wildIdleMax = 2.5

	// NoticeRadius is how close the player must come, in world pixels, for an
	// idle or landing wild critter to notice it.
	NoticeRadius = 32.0
	// NoticeMargin is how far beyond NoticeRadius the player must go before a
	// noticing critter loses interest, so a player standing near the edge
	// does not make it flicker between noticing and idling.
	NoticeMargin = 8.0
)

// WildState is what a wild critter is doing.
type WildState int

// The wild critter's states.
const (
	// WildIdle stands still.
	WildIdle WildState = iota
	// WildHopping is one hop in an arc.
	WildHopping
	// WildNoticing stands still and faces the player.
	WildNoticing
)

// Wild is a wild critter that stands about and hops around inside its zone,
// and stops to look at the player when it comes close.
// Its randomness comes from the rng it is given, so the same seed and the
// same ticks always move it the same way.
type Wild struct {
	// Pos is where the critter's feet touch the ground, in world pixels.
	Pos    geom.Vec
	Facing Facing
	// Walls, when non-nil, keeps the critter's hops out of solid scenery.
	Walls Walls
	// Rare marks a critter of a rare species, which sparkles.
	Rare bool

	zone  geom.Rect
	rng   *rand.Rand
	state WildState
	left  float64  // seconds left in the current state
	vel   geom.Vec // world pixels per second during a hop
}

// NewWild returns a critter standing idle at pos, moved inside zone if it is
// outside, that draws its rests and hops from rng.
func NewWild(pos geom.Vec, zone geom.Rect, rng *rand.Rand) *Wild {
	w := &Wild{Pos: zone.Clamp(pos), zone: zone, rng: rng}
	w.rest()
	return w
}

// State returns what the critter is doing.
func (w *Wild) State() WildState {
	return w.state
}

// Box returns the critter's 12 by 10 feet rectangle in world pixels, with its
// bottom centre at Pos.
func (w *Wild) Box() geom.Rect {
	return feetBox(w.Pos)
}

// Noticing reports whether the critter has noticed the player and is
// standing still facing it.
func (w *Wild) Noticing() bool {
	return w.state == WildNoticing
}

// HopHeight returns how high above the ground the critter is, in world pixels.
func (w *Wild) HopHeight() float64 {
	if w.state != WildHopping || w.left <= 0 {
		return 0
	}
	return hopHeight * math.Sin(math.Pi*(1-w.left/HopTime))
}

// Update moves the critter on by dt seconds. playerPos is where the player's
// feet are. An idle critter, or one landing a hop, with the player within
// NoticeRadius notices it; a hop in the air is always finished first. A
// noticing critter keeps facing the player and goes back to resting once the
// player is further than NoticeRadius plus NoticeMargin.
func (w *Wild) Update(dt float64, playerPos geom.Vec) {
	toPlayer := playerPos.Sub(w.Pos)
	switch w.state {
	case WildIdle:
		if toPlayer.Len() <= NoticeRadius {
			w.notice(toPlayer)
			return
		}
		w.left -= dt
		if w.left <= 0 {
			w.hop()
		}
	case WildHopping:
		w.move(w.vel.Scale(math.Min(dt, w.left)))
		w.left -= dt
		if w.left > 0 {
			return
		}
		if toPlayer = playerPos.Sub(w.Pos); toPlayer.Len() <= NoticeRadius {
			w.notice(toPlayer)
			return
		}
		w.rest()
	case WildNoticing:
		if toPlayer.Len() > NoticeRadius+NoticeMargin {
			w.rest()
			return
		}
		w.Facing = w.Facing.Toward(toPlayer)
	}
}

// notice stops the critter and turns it toward the player, which is
// toPlayer away from its feet.
func (w *Wild) notice(toPlayer geom.Vec) {
	w.state = WildNoticing
	w.vel = geom.Vec{}
	w.left = 0
	w.Facing = w.Facing.Toward(toPlayer)
}

// rest starts standing still for a random while.
func (w *Wild) rest() {
	w.state = WildIdle
	w.vel = geom.Vec{}
	w.left = wildIdleMin + w.rng.Float64()*(wildIdleMax-wildIdleMin)
}

// hop starts a hop in a random direction. A landing point outside the zone is
// pulled back to its edge; walls are dealt with as the hop moves.
func (w *Wild) hop() {
	angle := w.rng.Float64() * 2 * math.Pi
	dist := wildHopMinDistance + w.rng.Float64()*(WildHopDistance-wildHopMinDistance)
	aim := geom.Vec{X: math.Cos(angle), Y: math.Sin(angle)}
	landing := w.zone.Clamp(w.Pos.Add(aim.Scale(dist)))

	w.state = WildHopping
	w.left = HopTime
	w.vel = landing.Sub(w.Pos).Scale(1 / HopTime)
	w.Facing = w.Facing.Toward(aim)
}

// move shifts the critter by delta, kept inside its zone and, with Walls
// set, stopped flush against solid scenery.
func (w *Wild) move(delta geom.Vec) {
	next := w.zone.Clamp(w.Pos.Add(delta))
	delta = next.Sub(w.Pos)
	if w.Walls != nil {
		delta = w.Walls.Slide(w.Box(), delta)
	}
	w.Pos = w.zone.Clamp(w.Pos.Add(delta))
}
