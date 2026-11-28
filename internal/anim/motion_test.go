package anim_test

import (
	"math"
	"testing"

	"github.com/Elpulgo/critterkingdom/internal/anim"
	"github.com/Elpulgo/critterkingdom/internal/geom"
)

// onePartSheet is a sheet with a single body part, enough to watch one motion.
const onePartSheet = `{"pxPerWorldPixel": 8, "views": {"down": {"parts": [
	{"name": "body", "image": "down_body.png", "x": -6, "y": -10, "w": 12, "h": 10, "pivotX": 0, "pivotY": -5}
]}}}`

// near reports whether a and b are equal up to floating-point noise.
func near(a, b float64) bool { return math.Abs(a-b) < 1e-9 }

// bodyPoseAt returns the body's pose when the idle animation is only the given
// motion, t seconds in.
func bodyPoseAt(t *testing.T, motion string, at float64) anim.PartPose {
	t.Helper()
	layout, err := anim.ParseLayout([]byte(onePartSheet))
	if err != nil {
		t.Fatalf("ParseLayout: %v", err)
	}
	c, err := anim.ParseCharacter(layout, []byte(`{"sprite": "test/body", "animations": {"idle": [`+motion+`], "walk": []}}`))
	if err != nil {
		t.Fatalf("ParseCharacter: %v", err)
	}
	return c.Idle.Pose(layout.View(anim.ViewDown), at)[0]
}

func TestMotions_MoveTheirPartsByTheirFormulaAtAGivenTime(t *testing.T) {
	tests := []struct {
		name   string
		motion string
		t      float64
		offset geom.Vec
		turn   float64
		scaleX float64
		scaleY float64
	}{
		{"breathe at a quarter period is tallest and thinnest",
			`{"motion": "breathe", "size": 0.1, "period": 2}`, 0.5, geom.Vec{}, 0, 0.9, 1.1},
		{"breathe at three quarters is shortest and widest",
			`{"motion": "breathe", "size": 0.1, "period": 2}`, 1.5, geom.Vec{}, 0, 1.1, 0.9},
		{"bob at half a period lifts by its size",
			`{"motion": "bob", "size": 1, "period": 0.3}`, 0.15, geom.Vec{Y: -1}, 0, 1, 1},
		{"bob at a quarter period lifts by half its size",
			`{"motion": "bob", "size": 1, "period": 0.3}`, 0.075, geom.Vec{Y: -0.5}, 0, 1, 1},
		{"bob at a whole period is back on the ground",
			`{"motion": "bob", "size": 1, "period": 0.3}`, 0.3, geom.Vec{}, 0, 1, 1},
		{"patter at a quarter period lifts by its size",
			`{"motion": "patter", "size": 2, "period": 1}`, 0.25, geom.Vec{Y: -2}, 0, 1, 1},
		{"patter stays on the ground at three quarters",
			`{"motion": "patter", "size": 2, "period": 1}`, 0.75, geom.Vec{}, 0, 1, 1},
		{"patter stays on the ground early in the second half",
			`{"motion": "patter", "size": 2, "period": 1}`, 0.6, geom.Vec{}, 0, 1, 1},
		{"patter with phase 0.5 is on the ground at a quarter period",
			`{"motion": "patter", "size": 2, "period": 1, "phase": 0.5}`, 0.25, geom.Vec{}, 0, 1, 1},
		{"patter with phase 0.5 lifts by its size at three quarters",
			`{"motion": "patter", "size": 2, "period": 1, "phase": 0.5}`, 0.75, geom.Vec{Y: -2}, 0, 1, 1},
		{"wiggle at a quarter period turns by its size in degrees",
			`{"motion": "wiggle", "size": 10, "period": 1}`, 0.25, geom.Vec{}, 0.17453292519943295, 1, 1},
		{"wiggle at three quarters turns the other way",
			`{"motion": "wiggle", "size": 10, "period": 1}`, 0.75, geom.Vec{}, -0.17453292519943295, 1, 1},
		{"blink is open at the start",
			`{"motion": "blink", "period": 4, "duration": 0.2}`, 0, geom.Vec{}, 0, 1, 1},
		{"blink is open just before its last duration",
			`{"motion": "blink", "period": 4, "duration": 0.2}`, 3.7, geom.Vec{}, 0, 1, 1},
		{"blink is closed in its last duration",
			`{"motion": "blink", "period": 4, "duration": 0.2}`, 3.85, geom.Vec{}, 0, 1, 0.1},
		{"blink opens again when the next period starts",
			`{"motion": "blink", "period": 4, "duration": 0.2}`, 4, geom.Vec{}, 0, 1, 1},
		{"blink closes again at the end of the second period",
			`{"motion": "blink", "period": 4, "duration": 0.2}`, 7.9, geom.Vec{}, 0, 1, 0.1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := bodyPoseAt(t, tt.motion, tt.t)
			if !near(got.Offset.X, tt.offset.X) || !near(got.Offset.Y, tt.offset.Y) ||
				!near(got.Turn, tt.turn) || !near(got.ScaleX, tt.scaleX) || !near(got.ScaleY, tt.scaleY) {
				t.Errorf("pose at t=%v: offset %v turn %v scale %v,%v; want offset %v turn %v scale %v,%v",
					tt.t, got.Offset, got.Turn, got.ScaleX, got.ScaleY, tt.offset, tt.turn, tt.scaleX, tt.scaleY)
			}
		})
	}
}
