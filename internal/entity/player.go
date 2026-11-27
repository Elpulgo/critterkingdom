// Package entity holds the things that live in the world. For now that is the
// player's critter; villagers, wild critters, enemies and items join later.
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
)

// Player is the critter the child plays.
type Player struct {
	// Pos is where the critter's feet touch the ground, in world pixels.
	Pos    geom.Vec
	Facing Facing

	target  geom.Vec
	walking bool    // walking to target
	moved   bool    // moved during the last update
	hop     float64 // seconds left of the current hop
}

// NewPlayer returns a critter standing at pos, facing the player.
func NewPlayer(pos geom.Vec) *Player {
	return &Player{Pos: pos}
}

// WalkTo makes the critter walk in a straight line to target and stop there.
func (p *Player) WalkTo(target geom.Vec) {
	p.target = target
	p.walking = true
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
	p.moved = next != p.Pos
	p.Pos = next
}
