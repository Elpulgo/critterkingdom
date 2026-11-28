// Package entity holds the things that live in the world: the player's
// critter, the villagers and wild critters; enemies and items join later.
package entity

import (
	"math"

	"github.com/Elpulgo/critterkingdom/internal/geom"
)

const (
	// WalkSpeed is the critter's brisk walk in world pixels per second, four
	// and a half tiles. There is no run button.
	WalkSpeed = 72.0
	// HopTime is how long a happy hop lasts, in seconds.
	HopTime = 0.35
	// hopHeight is how high a hop goes, in world pixels.
	hopHeight = 6.0

	// boxHalfWidth and boxHeight size the critter's 12 by 10 feet collision
	// box, whose bottom centre sits at Pos.
	boxHalfWidth = 6.0
	boxHeight    = 10.0
)

// Walls stops the critter from walking through solid scenery and slides it
// along walls.
type Walls interface {
	Slide(box geom.Rect, delta geom.Vec) geom.Vec
}

// Player is the critter the child plays.
type Player struct {
	// Pos is where the critter's feet touch the ground, in world pixels.
	Pos    geom.Vec
	Facing Facing
	// Walls, when non-nil, constrains movement against solid scenery.
	Walls Walls

	route []geom.Vec // points still to walk to, in order; empty when not walking
	moved bool       // moved during the last update
	hop   float64    // seconds left of the current hop
}

// NewPlayer returns a critter standing at pos, facing the player.
func NewPlayer(pos geom.Vec) *Player {
	return &Player{Pos: pos}
}

// Box returns the critter's 12 by 10 collision rectangle in world pixels, with
// its bottom centre at Pos.
func (p *Player) Box() geom.Rect {
	return feetBox(p.Pos)
}

// feetBox returns a character's 12 by 10 collision rectangle with its bottom
// centre at the feet position pos.
func feetBox(pos geom.Vec) geom.Rect {
	return geom.Rect{
		Min: geom.Vec{X: pos.X - boxHalfWidth, Y: pos.Y - boxHeight},
		Max: geom.Vec{X: pos.X + boxHalfWidth, Y: pos.Y},
	}
}

// WalkTo makes the critter walk in a straight line to target and stop there.
func (p *Player) WalkTo(target geom.Vec) {
	p.WalkAlong([]geom.Vec{target})
}

// WalkAlong makes the critter walk in a straight line to each point in turn,
// passing through every one, and stop on the last. An empty route stops the
// walk. The points are copied, so the caller may reuse the slice.
func (p *Player) WalkAlong(points []geom.Vec) {
	p.route = append(p.route[:0:0], points...)
}

// Retarget moves the last point of the walk route to target, keeping the
// points before it, as when a held pointer moves within the spot it was
// pointing at. It does nothing when the critter is not walking.
func (p *Player) Retarget(target geom.Vec) {
	if n := len(p.route); n > 0 {
		p.route[n-1] = target
	}
}

// Place puts the critter's feet at pos, as when it arrives on another map. It
// stops any walk, since a walk target belongs to the map the critter left;
// facing and a hop under way are kept.
func (p *Player) Place(pos geom.Vec) {
	p.Pos = pos
	p.route = nil
}

// Target returns the spot the critter is walking to, the last point of its
// route, and false when it is not walking anywhere.
func (p *Player) Target() (geom.Vec, bool) {
	if len(p.route) == 0 {
		return geom.Vec{}, false
	}
	return p.route[len(p.route)-1], true
}

// Moved reports whether the critter moved during the last update.
func (p *Player) Moved() bool {
	return p.moved
}

// Hop starts a happy hop. It does nothing and returns false while a hop is
// already under way.
func (p *Player) Hop() bool {
	if p.hop > 0 {
		return false
	}
	p.hop = HopTime
	return true
}

// HopHeight returns how high above the ground the critter is, in world pixels.
func (p *Player) HopHeight() float64 {
	if p.hop <= 0 {
		return 0
	}
	return hopHeight * math.Sin(math.Pi*(1-p.hop/HopTime))
}

// Update moves the critter for dt seconds. A non-zero dir, from the keys, a
// stick or the touch pad, walks that way and cancels the whole walk route. A
// walk heads for the route's next point and, on reaching it, goes on to the
// one after from the next update; a step that walls block (moving less than a
// tenth of what was meant) drops the whole route. The critter's feet never
// leave area, and a route point outside it is walked to the nearest point
// inside.
func (p *Player) Update(dt float64, dir geom.Vec, area geom.Rect) {
	p.hop = math.Max(0, p.hop-dt)

	step := WalkSpeed * dt
	next := p.Pos
	switch {
	case !dir.IsZero():
		p.route = nil
		next = p.Pos.Add(dir.Scale(step))
	case len(p.route) > 0:
		target := area.Clamp(p.route[0])
		to := target.Sub(p.Pos)
		if dist := to.Len(); dist > step {
			next = p.Pos.Add(to.Scale(step / dist))
		} else {
			next = target
			p.route = p.route[1:]
		}
	}

	// The critter faces the way it tries to go, even when the edge of the
	// area stops it, so pushing against a hedge still turns it.
	if want := next.Sub(p.Pos); !want.IsZero() {
		p.Facing = p.Facing.Toward(want)
	}
	next = area.Clamp(next)
	intended := next.Sub(p.Pos).Len()
	if p.Walls != nil {
		next = p.Pos.Add(p.Walls.Slide(p.Box(), next.Sub(p.Pos)))
	}
	if len(p.route) > 0 && next.Sub(p.Pos).Len() < 0.1*intended {
		p.route = nil
	}
	p.moved = next != p.Pos
	p.Pos = next
}
