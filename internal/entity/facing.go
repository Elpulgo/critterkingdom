package entity

import (
	"math"

	"github.com/Elpulgo/critterkingdom/internal/geom"
)

// Facing is the way a character looks. Characters walk in eight directions
// but face only four, so every sprite needs four views.
type Facing int

// The four facings. Down, towards the player, is the zero value.
const (
	FacingDown Facing = iota
	FacingUp
	FacingLeft
	FacingRight
)

// keepFacing is how large the movement along the current facing must stay,
// relative to the movement across it, for a character to keep that facing.
// Below 1 it keeps the facing a little past the diagonal, so walking at about
// 45 degrees does not flicker between two facings.
const keepFacing = 0.75

// Toward returns the facing of a character that faces f and moves along v.
// A diagonal move keeps the current facing when that is one of the two
// directions involved; otherwise the character faces the larger part of the
// move, sideways when both parts are equal. A zero v keeps f.
func (f Facing) Toward(v geom.Vec) Facing {
	if v.IsZero() || f.allows(v) {
		return f
	}
	return facingAlong(v)
}

// facingAlong returns the facing towards the larger part of a non-zero v,
// sideways when both parts are equal.
func facingAlong(v geom.Vec) Facing {
	if math.Abs(v.X) >= math.Abs(v.Y) {
		if v.X > 0 {
			return FacingRight
		}
		return FacingLeft
	}
	if v.Y > 0 {
		return FacingDown
	}
	return FacingUp
}

// dir returns the unit vector f points along, with Y growing downwards.
func (f Facing) dir() geom.Vec {
	switch f {
	case FacingUp:
		return geom.Vec{Y: -1}
	case FacingLeft:
		return geom.Vec{X: -1}
	case FacingRight:
		return geom.Vec{X: 1}
	default:
		return geom.Vec{Y: 1}
	}
}

// allows reports whether a character facing f can keep facing that way while
// moving along v.
func (f Facing) allows(v geom.Vec) bool {
	along, across := v.Y, v.X
	if f == FacingLeft || f == FacingRight {
		along, across = v.X, v.Y
	}
	if f == FacingUp || f == FacingLeft {
		along = -along
	}
	return along > 0 && along >= keepFacing*math.Abs(across)
}

// String returns the facing's name.
func (f Facing) String() string {
	switch f {
	case FacingUp:
		return "up"
	case FacingLeft:
		return "left"
	case FacingRight:
		return "right"
	default:
		return "down"
	}
}
