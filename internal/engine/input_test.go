package engine

import (
	"math"
	"testing"

	"github.com/Elpulgo/critterkingdom/internal/geom"
)

// d is each part of a diagonal direction of length 1.
const d = math.Sqrt2 / 2

func TestSnapDirection_SnapsAStickToTheNearestOfEightDirections(t *testing.T) {
	tests := []struct {
		name  string
		stick geom.Vec
		want  geom.Vec
	}{
		{"a light push inside the dead zone does not walk", geom.Vec{X: 0.1, Y: 0.1}, geom.Vec{}},
		{"mostly right walks right", geom.Vec{X: 0.9, Y: 0.3}, geom.Vec{X: 1}},
		{"nearer the diagonal than right walks diagonally", geom.Vec{X: 0.7, Y: 0.5}, geom.Vec{X: d, Y: d}},
		{"mostly left walks left, from above too", geom.Vec{X: -0.9, Y: -0.1}, geom.Vec{X: -1}},
		{"mostly up walks up", geom.Vec{X: -0.2, Y: -0.9}, geom.Vec{Y: -1}},
		{"down and left walks diagonally", geom.Vec{X: -0.5, Y: 0.45}, geom.Vec{X: -d, Y: d}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := snapDirection(tt.stick); !near(got, tt.want) {
				t.Errorf("snapDirection(%v) = %v, want %v", tt.stick, got, tt.want)
			}
		})
	}
}

func TestKeyDirection_DiagonalsAreAsFastAsStraightLinesAndOppositeKeysCancel(t *testing.T) {
	tests := []struct {
		name                  string
		up, down, left, right bool
		want                  geom.Vec
	}{
		{"up and right walk diagonally at full speed", true, false, false, true, geom.Vec{X: d, Y: -d}},
		{"left and right cancel", false, false, true, true, geom.Vec{}},
		{"left, right and down walk down", false, true, true, true, geom.Vec{Y: 1}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := keyDirection(tt.up, tt.down, tt.left, tt.right); !near(got, tt.want) {
				t.Errorf("keyDirection = %v, want %v", got, tt.want)
			}
		})
	}
}

func near(a, b geom.Vec) bool {
	return math.Abs(a.X-b.X) < 1e-9 && math.Abs(a.Y-b.Y) < 1e-9
}
