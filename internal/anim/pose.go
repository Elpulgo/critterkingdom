package anim

import (
	"math"

	"github.com/Elpulgo/critterkingdom/internal/geom"
)

// PartPose is where one part sits at one moment: the part at rest (its box
// and pivot from the feet) and how the motions move it.
type PartPose struct {
	Part
	// Offset moves the part, in world pixels (negative Y is up).
	Offset geom.Vec
	// Turn turns the part about its pivot, in radians. Y grows downwards, so
	// a positive turn is clockwise on screen.
	Turn float64
	// ScaleX and ScaleY squash the part about its pivot; 1 is its drawn size.
	ScaleX, ScaleY float64
	// Mirrored flips the posed part about the feet (X becomes -X), for the
	// side view drawn facing the other way.
	Mirrored bool
}

// restPose returns the part where it is drawn, unmoved.
func restPose(p Part) PartPose {
	return PartPose{Part: p, ScaleX: 1, ScaleY: 1}
}

// Apply maps a point of the part at rest, in world pixels from the feet, to
// its posed place: squashed and then turned about the pivot, moved by the
// offset, and finally mirrored about the feet when Mirrored is set.
func (pp PartPose) Apply(p geom.Vec) geom.Vec {
	d := p.Sub(pp.Pivot)
	d = geom.Vec{X: d.X * pp.ScaleX, Y: d.Y * pp.ScaleY}
	sin, cos := math.Sincos(pp.Turn)
	d = geom.Vec{X: d.X*cos - d.Y*sin, Y: d.X*sin + d.Y*cos}
	q := pp.Pivot.Add(d).Add(pp.Offset)
	if pp.Mirrored {
		q.X = -q.X
	}
	return q
}
