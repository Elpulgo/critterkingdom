package engine

import (
	"testing"

	"github.com/Elpulgo/critterkingdom/internal/geom"
)

func TestCameraFor_CentresTheFocusButNeverShowsPastTheEdges(t *testing.T) {
	meadow := geom.Rect{Max: geom.Vec{X: 640, Y: 384}}
	tests := []struct {
		name  string
		focus geom.Vec
		want  geom.Vec
	}{
		{"in the middle the focus is centred", geom.Vec{X: 320, Y: 192}, geom.Vec{X: 160, Y: 102}},
		{"near the top-left corner the view stops at the edges", geom.Vec{X: 10, Y: 10}, geom.Vec{}},
		{"near the bottom-right corner the view stops at the edges", geom.Vec{X: 630, Y: 380}, geom.Vec{X: 320, Y: 204}},
		{"the view snaps to whole pixels", geom.Vec{X: 300.6, Y: 100.4}, geom.Vec{X: 140, Y: 10}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CameraFor(tt.focus, 320, 180, meadow); got != tt.want {
				t.Errorf("CameraFor(%v) = %v, want %v", tt.focus, got, tt.want)
			}
		})
	}
}

func TestCameraFor_CentresAMapSmallerThanTheView(t *testing.T) {
	small := geom.Rect{Max: geom.Vec{X: 200, Y: 100}}
	if got, want := CameraFor(geom.Vec{X: 10, Y: 90}, 320, 180, small), (geom.Vec{X: -60, Y: -40}); got != want {
		t.Errorf("CameraFor = %v, want %v", got, want)
	}
}

func TestSmoothCameraFor_MovesInWholeScreenPixels(t *testing.T) {
	meadow := geom.Rect{Max: geom.Vec{X: 640, Y: 384}}
	// At 4 screen pixels to a world pixel, the corner moves in quarter world pixels.
	if got, want := SmoothCameraFor(geom.Vec{X: 300.6, Y: 100.4}, 320, 180, meadow, 4), (geom.Vec{X: 140.5, Y: 10.25}); got != want {
		t.Errorf("SmoothCameraFor = %v, want %v", got, want)
	}
}
