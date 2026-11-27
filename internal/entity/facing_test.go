package entity

import (
	"testing"

	"github.com/Elpulgo/critterkingdom/internal/geom"
)

func TestFacingToward_KeepsTheFacingOnADiagonalAndTurnsToTheLargerPartOtherwise(t *testing.T) {
	tests := []struct {
		name string
		from Facing
		move geom.Vec
		want Facing
	}{
		{"walking sideways turns sideways", FacingDown, geom.Vec{X: 1}, FacingRight},
		{"a diagonal keeps a sideways facing it includes", FacingRight, geom.Vec{X: 1, Y: 1}, FacingRight},
		{"the same diagonal keeps a down facing too", FacingDown, geom.Vec{X: 1, Y: 1}, FacingDown},
		{"a diagonal away from the facing turns sideways on a tie", FacingUp, geom.Vec{X: 1, Y: 1}, FacingRight},
		{"walking up from sideways turns up", FacingLeft, geom.Vec{Y: -1}, FacingUp},
		{"a little past the diagonal keeps the facing", FacingDown, geom.Vec{X: 1, Y: 0.9}, FacingDown},
		{"well past the diagonal turns", FacingDown, geom.Vec{X: 1, Y: 0.5}, FacingRight},
		{"standing still keeps the facing", FacingRight, geom.Vec{}, FacingRight},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.from.Toward(tt.move); got != tt.want {
				t.Errorf("%v.Toward(%v) = %v, want %v", tt.from, tt.move, got, tt.want)
			}
		})
	}
}
