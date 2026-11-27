package combat

import (
	"math"
	"testing"

	"github.com/Elpulgo/critterkingdom/internal/geom"
)

func TestKnockback_PushesTheTargetStraightAwayFromTheAttacker(t *testing.T) {
	tests := []struct {
		name     string
		attacker geom.Vec
		target   geom.Vec
		want     geom.Vec
	}{
		{
			"attacker on the left pushes right",
			geom.Vec{X: 0, Y: 0},
			geom.Vec{X: 5, Y: 0},
			geom.Vec{X: 12, Y: 0},
		},
		{
			"attacker on the right pushes left",
			geom.Vec{X: 20, Y: 10},
			geom.Vec{X: 12, Y: 10},
			geom.Vec{X: -12, Y: 0},
		},
		{
			"attacker above pushes down",
			geom.Vec{X: 10, Y: 4},
			geom.Vec{X: 10, Y: 9},
			geom.Vec{X: 0, Y: 12},
		},
		{
			"attacker below pushes up",
			geom.Vec{X: 10, Y: 30},
			geom.Vec{X: 10, Y: 15},
			geom.Vec{X: 0, Y: -12},
		},
		{
			"attacker on a 3-4-5 diagonal pushes along the diagonal",
			geom.Vec{X: 0, Y: 0},
			geom.Vec{X: 3, Y: 4},
			geom.Vec{X: 7.2, Y: 9.6},
		},
		{
			"attacker and target on the same spot do not push at all",
			geom.Vec{X: 14, Y: 22},
			geom.Vec{X: 14, Y: 22},
			geom.Vec{X: 0, Y: 0},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Knockback(tt.attacker, tt.target)
			if math.Abs(got.X-tt.want.X) > 1e-9 || math.Abs(got.Y-tt.want.Y) > 1e-9 {
				t.Errorf("Knockback(%v, %v) = %v, want %v", tt.attacker, tt.target, got, tt.want)
			}
		})
	}
}
