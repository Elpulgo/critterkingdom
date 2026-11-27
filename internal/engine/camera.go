package engine

import (
	"math"

	"github.com/Elpulgo/critterkingdom/internal/geom"
)

// CameraFor returns the top-left corner, in world pixels, of a w by h view
// that keeps focus in the middle without showing anything outside bounds.
// When bounds is smaller than the view along an axis, the view centres bounds
// along that axis instead. The corner is rounded down to whole pixels so the
// world does not shimmer while it scrolls.
func CameraFor(focus geom.Vec, w, h int, bounds geom.Rect) geom.Vec {
	return geom.Vec{
		X: cameraAxis(focus.X, float64(w), bounds.Min.X, bounds.Max.X),
		Y: cameraAxis(focus.Y, float64(h), bounds.Min.Y, bounds.Max.Y),
	}
}

// SmoothCameraFor is CameraFor for a world drawn straight onto the screen,
// with scale screen pixels to a world pixel. The corner is rounded down to
// whole screen pixels instead of whole world pixels, so the view glides
// smoothly and the world still does not shimmer while it scrolls.
func SmoothCameraFor(focus geom.Vec, w, h int, bounds geom.Rect, scale int) geom.Vec {
	s := float64(scale)
	onScreen := geom.Rect{Min: bounds.Min.Scale(s), Max: bounds.Max.Scale(s)}
	return CameraFor(focus.Scale(s), w*scale, h*scale, onScreen).Scale(1 / s)
}

func cameraAxis(focus, size, lo, hi float64) float64 {
	if hi-lo <= size {
		return math.Floor(lo - (size-(hi-lo))/2)
	}
	return math.Floor(math.Max(lo, math.Min(focus-size/2, hi-size)))
}
