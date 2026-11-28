package world

import (
	"math"
	"testing"

	"github.com/Elpulgo/critterkingdom/internal/geom"
)

func TestLight_Brightness_IsFullInCoreFadesSoftlyAndIsZeroFromRadius(t *testing.T) {
	// Radius 100, so the core ends at 45 and the edge band runs 45..100.
	light := Light{Pos: geom.Vec{X: 10, Y: 20}, Radius: 100}

	// Rows are in order of increasing distance from the light.
	cases := []struct {
		name   string
		offset geom.Vec
		want   float64
	}{
		{"centre", geom.Vec{}, 1},
		{"inside the core, diagonally", geom.Vec{X: -24, Y: 32}, 1}, // distance 40
		{"core edge", geom.Vec{X: 45}, 1},
		{"quarter into the edge band", geom.Vec{Y: 58.75}, 0.84375},
		{"middle of the edge band", geom.Vec{X: 72.5}, 0.5},
		{"three quarters into the edge band", geom.Vec{X: -86.25}, 0.15625},
		{"at the radius", geom.Vec{Y: -100}, 0},
		{"beyond the radius", geom.Vec{X: 90, Y: 120}, 0}, // distance 150
	}

	prev := math.Inf(1)
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := light.Brightness(light.Pos.Add(tc.offset))
			if math.Abs(got-tc.want) > 1e-9 {
				t.Errorf("Brightness = %v, want %v", got, tc.want)
			}
			if got > prev {
				t.Errorf("Brightness = %v is brighter than %v closer to the light", got, prev)
			}
			prev = got
		})
	}
}

func TestLight_Brightness_ZeroRadiusLightsNothing(t *testing.T) {
	for _, radius := range []float64{0, -10} {
		light := Light{Pos: geom.Vec{X: 5, Y: 5}, Radius: radius}
		for _, p := range []geom.Vec{{X: 5, Y: 5}, {X: 6, Y: 5}, {X: 500, Y: 500}} {
			if got := light.Brightness(p); got != 0 {
				t.Errorf("Light{Radius: %v}.Brightness(%v) = %v, want 0", radius, p, got)
			}
		}
	}
}
