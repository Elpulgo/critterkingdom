package entity

import "github.com/Elpulgo/critterkingdom/internal/geom"

// InteractReach is how close, feet to feet in world pixels, something must be
// for the critter to interact with it: 1.5 tiles.
const InteractReach = 24.0

// interactCone is the smallest share of a candidate's distance that must lie
// along the critter's facing: 0.5 is cos 60°, so the cone in front of the
// critter is 120 degrees wide.
const interactCone = 0.5

// PickInteractable returns the index of the candidate the critter, with its
// feet at pos and facing facing, would interact with: the nearest one within
// InteractReach that lies inside the 120 degree cone in front of it. A
// candidate behind the critter, straight beside it or on its feet is never
// picked. On equal distances the lower index wins. ok is false when no
// candidate qualifies.
func PickInteractable(pos geom.Vec, facing Facing, candidates []geom.Vec) (index int, ok bool) {
	ahead := facing.dir()
	best := 0.0
	for i, c := range candidates {
		off := c.Sub(pos)
		dist := off.Len()
		if dist > InteractReach {
			continue
		}
		along := off.X*ahead.X + off.Y*ahead.Y
		if along <= 0 || along < interactCone*dist {
			continue
		}
		if !ok || dist < best {
			index, best, ok = i, dist, true
		}
	}
	return index, ok
}
