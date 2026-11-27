package engine

import "github.com/Elpulgo/critterkingdom/internal/geom"

// The world is drawn at a low resolution and scaled up by a whole number, so
// pixel art stays crisp on every screen. The scale is the largest that still
// shows at least minWorldWidth by minWorldHeight world pixels. The view then
// grows to fill the screen, up to maxWorldWidth by maxWorldHeight, so wide
// and tall screens see more of the world instead of black bars.
const (
	minWorldWidth  = 240
	minWorldHeight = 180
	maxWorldWidth  = 480
	maxWorldHeight = 240
)

// View says how the world image maps onto the screen.
type View struct {
	// Scale is how many screen pixels one world pixel covers.
	Scale int
	// W and H are the size of the visible world, in world pixels.
	W, H int
	// OffX and OffY are where the top-left corner of the world image lands on
	// the screen, in screen pixels. They are negative when the scaled world
	// image is a little larger than the screen.
	OffX, OffY int
}

// viewFor returns the view for a screen of sw by sh pixels.
func viewFor(sw, sh int) View {
	scale := max(1, min(sh/minWorldHeight, sw/minWorldWidth))
	w := min(max(1, ceilDiv(sw, scale)), maxWorldWidth)
	h := min(max(1, ceilDiv(sh, scale)), maxWorldHeight)
	return View{
		Scale: scale,
		W:     w,
		H:     h,
		OffX:  (sw - w*scale) / 2,
		OffY:  (sh - h*scale) / 2,
	}
}

// ScreenToView converts a point on the screen to view pixels.
func (v View) ScreenToView(x, y float64) geom.Vec {
	s := float64(v.Scale)
	return geom.Vec{X: (x - float64(v.OffX)) / s, Y: (y - float64(v.OffY)) / s}
}

// ViewToScreen converts a point in view pixels to the screen.
func (v View) ViewToScreen(p geom.Vec) (x, y float64) {
	s := float64(v.Scale)
	return p.X*s + float64(v.OffX), p.Y*s + float64(v.OffY)
}

func ceilDiv(a, b int) int {
	return (a + b - 1) / b
}
