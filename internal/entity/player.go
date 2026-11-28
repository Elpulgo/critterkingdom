// Package entity holds the things that live in the world: the player's
// critter and wild critters; villagers, enemies and items join later.
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

	target  geom.Vec
	walking bool    // walking to target
	moved   bool    // moved during the last update
	hop     float64 // seconds left of the current hop
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

// feetBox returns the 12 by 10 feet rectangle of a critter whose feet are at
// pos, with its bottom centre at pos.
func feetBox(pos geom.Vec) geom.Rect {
	return geom.Rect{
		Min: geom.Vec{X: pos.X - boxHalfWidth, Y: pos.Y - boxHeight},
		Max: geom.Vec{X: pos.X + boxHalfWidth, Y: pos.Y},
	}
}

// WalkTo makes the critter walk in a straight line to target and stop there.
func (p *Player) WalkTo(target geom.Vec) {
	p.target = target
	p.walking = true
}

// Place puts the critter's feet at pos, as when it arrives on another map. It
// stops any walk, since a walk target belongs to the map the critter left;
// facing and a hop under way are kept.
func (p *Player) Place(pos geom.Vec) {
	p.Pos = pos
	p.walking = false
}

// Target returns the spot the critter is walking to, and false when it is not
// walking anywhere.
func (p *Player) Target() (geom.Vec, bool) {
	return p.target, p.walking
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
// stick or the touch pad, walks that way and cancels any walk target. The
// critter's feet never leave area, and a target outside it is walked to the
// nearest point inside.
func (p *Player) Update(dt float64, dir geom.Vec, area geom.Rect) {
	p.hop = math.Max(0, p.hop-dt)

	step := WalkSpeed * dt
	next := p.Pos
	switch {
	case !dir.IsZero():
		p.walking = false
		next = p.Pos.Add(dir.Scale(step))
	case p.walking:
		target := area.Clamp(p.target)
		to := target.Sub(p.Pos)
		if dist := to.Len(); dist > step {
			next = p.Pos.Add(to.Scale(step / dist))
		} else {
			next = target
			p.walking = false
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
	if p.walking && next.Sub(p.Pos).Len() < 0.1*intended {
		p.walking = false
	}
	p.moved = next != p.Pos
	p.Pos = next
}
