package world

import "github.com/Elpulgo/critterkingdom/internal/geom"

// lightCore is the share of a light's radius that is fully lit; the soft edge
// band runs from there out to the radius.
const lightCore = 0.45

// Light is a round light source on a dark map, such as the lantern the critter
// carries. Pos and Radius are in world pixels.
type Light struct {
	Pos    geom.Vec
	Radius float64
}

// Brightness is how strongly the light lights point p, from 0 (dark) to 1
// (fully lit). It is 1 within lightCore of the radius, falls smoothly to 0 at
// the radius along a smoothstep curve, and is 0 from the radius outward. A
// light with no positive radius lights nothing.
func (l Light) Brightness(p geom.Vec) float64 {
	if !(l.Radius > 0) {
		return 0
	}
	t := (p.Sub(l.Pos).Len()/l.Radius - lightCore) / (1 - lightCore)
	switch {
	case t <= 0:
		return 1
	case t >= 1:
		return 0
	}
	return 1 - t*t*(3-2*t)
}
