package entity

import (
	"math/rand/v2"

	"github.com/Elpulgo/critterkingdom/internal/geom"
)

const (
	// NPCWalkSpeed is how fast a villager strolls, in world pixels per second:
	// a third of the critter's walk.
	NPCWalkSpeed = WalkSpeed / 3

	// npcIdleMin and npcIdleMax bound how long a villager stands still between
	// walks, in seconds.
	npcIdleMin = 1.0
	npcIdleMax = 3.0
)

// NPC is a villager. It idles for a while, then walks to a random spot inside
// its wander area, then idles again.
type NPC struct {
	// ID names the villager, as in the map's npc object.
	ID string
	// Pos is where the villager's feet touch the ground, in world pixels.
	Pos    geom.Vec
	Facing Facing
	// Walls, when non-nil, constrains movement against solid scenery.
	Walls Walls

	area    geom.Rect
	rng     *rand.Rand
	idle    float64  // seconds left standing still
	target  geom.Vec // spot the villager walks to
	walking bool
}

// NewNPC returns the villager id standing at pos, facing the player, that
// wanders inside area using rng for its choices. A zero-size area makes a
// villager that stands still. rng must not be nil.
func NewNPC(id string, pos geom.Vec, area geom.Rect, rng *rand.Rand) *NPC {
	n := &NPC{ID: id, Pos: pos, area: area, rng: rng}
	n.rest()
	return n
}

// Box returns the villager's collision rectangle in world pixels, the same
// size as the critter's, with its bottom centre at Pos.
func (n *NPC) Box() geom.Rect {
	return feetBox(n.Pos)
}

// Update moves the villager for dt seconds. playerPos is where the critter's
// feet are.
func (n *NPC) Update(dt float64, playerPos geom.Vec) {
	_ = playerPos // facing the critter comes with the next change

	if !n.walking {
		n.idle -= dt
		if n.idle > 0 || !n.canWander() {
			return
		}
		n.target = geom.Vec{
			X: n.area.Min.X + n.rng.Float64()*(n.area.Max.X-n.area.Min.X),
			Y: n.area.Min.Y + n.rng.Float64()*(n.area.Max.Y-n.area.Min.Y),
		}
		n.walking = true
	}

	step := NPCWalkSpeed * dt
	to := n.target.Sub(n.Pos)
	next := n.target
	arrived := true
	if dist := to.Len(); dist > step {
		next = n.Pos.Add(to.Scale(step / dist))
		arrived = false
	}
	// The area is a rectangle and both ends of the walk are inside it, so the
	// clamp only guards against rounding; sliding along a wall shortens each
	// axis of the move and so keeps the feet inside too.
	next = n.area.Clamp(next)
	want := next.Sub(n.Pos)
	n.Facing = n.Facing.Toward(want)
	got := want
	if n.Walls != nil {
		got = n.Walls.Slide(n.Box(), want)
	}
	n.Pos = n.Pos.Add(got)

	// A walk blocked by a wall ends rather than rubbing against it.
	if arrived || got.Len() < 0.1*want.Len() {
		n.rest()
	}
}

// canWander reports whether the villager's area leaves it room to walk.
func (n *NPC) canWander() bool {
	return n.area.Max.X > n.area.Min.X || n.area.Max.Y > n.area.Min.Y
}

// rest stops any walk and stands the villager still for 1 to 3 seconds.
func (n *NPC) rest() {
	n.walking = false
	n.idle = npcIdleMin + n.rng.Float64()*(npcIdleMax-npcIdleMin)
}
