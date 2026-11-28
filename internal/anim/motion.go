package anim

import (
	"math"

	"github.com/Elpulgo/critterkingdom/internal/geom"
)

// MotionKind names one of the shared motions every character can use.
type MotionKind string

// The shared motions. Lean, squash and flash come with fights.
const (
	// Breathe squashes its parts about their pivots: taller and thinner,
	// then shorter and wider. Size is the scale fraction.
	Breathe MotionKind = "breathe"
	// Blink closes its parts (scaleY 0.1) for the last Duration seconds of
	// each period, so they are open at the start of a period.
	Blink MotionKind = "blink"
	// Bob lifts its parts smoothly by up to Size world pixels once per
	// period, highest at half a period and never below their rest place.
	Bob MotionKind = "bob"
	// Patter lifts its parts by up to Size world pixels during the first
	// half of each period and leaves them on the ground for the second half.
	Patter MotionKind = "patter"
	// Wiggle turns its parts about their pivots by up to Size degrees each
	// way.
	Wiggle MotionKind = "wiggle"
)

// blinkClosedScaleY is how far a blink squashes its parts.
const blinkClosedScaleY = 0.1

// blinkMinWeight is the least weight at which a blink still closes its parts:
// a blink cannot be half done, so it plays while its animation is the larger
// share of the pose.
const blinkMinWeight = 0.5

// knownMotion reports whether k is one of the shared motions.
func knownMotion(k MotionKind) bool {
	switch k {
	case Breathe, Blink, Bob, Patter, Wiggle:
		return true
	}
	return false
}

// Motion is one shared motion as a character's data picks it.
type Motion struct {
	Kind MotionKind
	// Parts are the part names the motion moves, in every view that has
	// them. No parts means the whole character.
	Parts []string
	// Size is how much the motion moves: a scale fraction for breathe,
	// world pixels for bob and patter, degrees for wiggle. Blink ignores it.
	Size float64
	// Period is the length of one cycle in seconds, always positive.
	Period float64
	// Phase shifts the cycle by a fraction of a period.
	Phase float64
	// Duration is how many seconds a blink keeps its parts closed at the
	// end of each period. Only blink uses it.
	Duration float64
}

// moves reports whether the motion moves the named part.
func (m Motion) moves(part string) bool {
	if len(m.Parts) == 0 {
		return true
	}
	for _, p := range m.Parts {
		if p == part {
			return true
		}
	}
	return false
}

// applyTo adds the motion at time t (seconds) to pose, its amount scaled by
// weight: 1 is full strength and 0 leaves the pose alone. A blink is not
// scaled; it closes its parts only while weight is at least half. Offsets and
// turns add up and scales multiply, so several motions can move one part.
func (m Motion) applyTo(pose *PartPose, t, weight float64) {
	u := t/m.Period + m.Phase
	switch m.Kind {
	case Breathe:
		s := weight * m.Size * math.Sin(2*math.Pi*u)
		pose.ScaleX *= 1 - s
		pose.ScaleY *= 1 + s
	case Blink:
		if weight < blinkMinWeight {
			return
		}
		into := (u - math.Floor(u)) * m.Period // seconds into this period
		if into >= m.Period-m.Duration {
			pose.ScaleY *= blinkClosedScaleY
		}
	case Bob:
		s := math.Sin(math.Pi * u)
		pose.Offset = pose.Offset.Add(geom.Vec{Y: -weight * m.Size * s * s})
	case Patter:
		lift := weight * m.Size * math.Max(0, math.Sin(2*math.Pi*u))
		pose.Offset = pose.Offset.Add(geom.Vec{Y: -lift})
	case Wiggle:
		pose.Turn += weight * m.Size * math.Pi / 180 * math.Sin(2*math.Pi*u)
	}
}
