package anim_test

import (
	"math"
	"testing"

	"github.com/Elpulgo/critterkingdom/internal/anim"
	"github.com/Elpulgo/critterkingdom/internal/geom"
)

func TestPartPoseApply_SquashesAndTurnsAboutThePivotThenMovesThenMirrorsAboutTheFeet(t *testing.T) {
	pivot := geom.Vec{X: 0, Y: -5}
	point := geom.Vec{X: 2, Y: -5} // two world pixels right of the pivot

	tests := []struct {
		name string
		pose anim.PartPose
		want geom.Vec
	}{
		{"a part at rest leaves the point where it is",
			anim.PartPose{ScaleX: 1, ScaleY: 1}, geom.Vec{X: 2, Y: -5}},
		{"a quarter turn swings the point below the pivot",
			anim.PartPose{Turn: math.Pi / 2, ScaleX: 1, ScaleY: 1}, geom.Vec{X: 0, Y: -3}},
		{"the offset moves the turned point",
			anim.PartPose{Offset: geom.Vec{X: 1, Y: -1}, Turn: math.Pi / 2, ScaleX: 1, ScaleY: 1}, geom.Vec{X: 1, Y: -4}},
		{"squash comes before the turn",
			anim.PartPose{Turn: math.Pi / 2, ScaleX: 2, ScaleY: 0.5}, geom.Vec{X: 0, Y: -1}},
		{"mirroring negates X from the feet after everything else",
			anim.PartPose{Offset: geom.Vec{X: 1, Y: -1}, Turn: math.Pi / 2, ScaleX: 1, ScaleY: 1, Mirrored: true}, geom.Vec{X: -1, Y: -4}},
		{"mirroring a part at rest flips it to the other side of the feet",
			anim.PartPose{ScaleX: 1, ScaleY: 1, Mirrored: true}, geom.Vec{X: -2, Y: -5}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.pose.Pivot = pivot
			got := tt.pose.Apply(point)
			if !near(got.X, tt.want.X) || !near(got.Y, tt.want.Y) {
				t.Errorf("Apply(%v) = %v, want %v", point, got, tt.want)
			}
		})
	}
}
