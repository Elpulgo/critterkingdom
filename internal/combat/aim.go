package combat

import (
	"math"

	"github.com/Elpulgo/critterkingdom/internal/geom"
)

const (
	// AimRange is how far ahead, in world pixels, the aim-assist cone reaches.
	AimRange = 40.0
	// AimHalfAngle is the half-width of the aim-assist cone, in radians (45 degrees).
	AimHalfAngle = math.Pi / 4
)

// Aim returns the unit vector toward the nearest enemy inside the aim-assist
// cone in front of origin along facing, or facing normalized when no enemy is
// inside the cone. A zero facing has no cone and returns the zero vector.
func Aim(origin, facing geom.Vec, enemies []geom.Vec) geom.Vec {
	f := facing.Normalized()
	if f.IsZero() {
		return geom.Vec{}
	}

	best := f
	bestDist := math.Inf(1)
	for _, enemy := range enemies {
		to := enemy.Sub(origin)
		dist := to.Len()
		if dist == 0 || dist > AimRange {
			continue
		}
		dir := to.Normalized()
		dot := dir.X*f.X + dir.Y*f.Y
		dot = math.Max(-1, math.Min(1, dot))
		if math.Acos(dot) > AimHalfAngle {
			continue
		}
		if dist < bestDist {
			bestDist = dist
			best = dir
		}
	}
	return best
}
