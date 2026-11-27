package world

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2/vector"

	"github.com/Elpulgo/critterkingdom/internal/engine"
	"github.com/Elpulgo/critterkingdom/internal/geom"
)

// pen draws shapes given in world pixels on a canvas, at full screen
// resolution.
type pen struct {
	c      engine.Canvas
	camera geom.Vec
	path   vector.Path
}

// begin points the pen at a canvas that shows the view with its top-left
// corner at camera.
func (p *pen) begin(c engine.Canvas, camera geom.Vec) {
	p.c, p.camera = c, camera
}

// at returns where the world point x, y lands on the canvas.
func (p *pen) at(x, y float64) (float32, float32) {
	return p.c.At(geom.Vec{X: x - p.camera.X, Y: y - p.camera.Y})
}

// rect fills a rectangle. It is not antialiased, because its edges lie on
// whole screen pixels when it is on whole world pixels and the camera comes
// from engine.SmoothCameraFor.
func (p *pen) rect(x, y, w, h float64, clr color.Color) {
	left, top := p.at(x, y)
	s := float32(p.c.Scale)
	vector.FillRect(p.c.Image, left, top, float32(w)*s, float32(h)*s, clr, false)
}

// ellipse fills an ellipse centred on x, y with radii rx and ry.
func (p *pen) ellipse(x, y, rx, ry float64, clr color.Color) {
	// k puts the control points of a cubic Bézier that follows a quarter
	// of a circle.
	const k = 0.5522847
	cx, cy := p.at(x, y)
	a, b := float32(rx*p.c.Scale), float32(ry*p.c.Scale)
	p.path.MoveTo(cx+a, cy)
	p.path.CubicTo(cx+a, cy+k*b, cx+k*a, cy+b, cx, cy+b)
	p.path.CubicTo(cx-k*a, cy+b, cx-a, cy+k*b, cx-a, cy)
	p.path.CubicTo(cx-a, cy-k*b, cx-k*a, cy-b, cx, cy-b)
	p.path.CubicTo(cx+k*a, cy-b, cx+a, cy-k*b, cx+a, cy)
	p.path.Close()
	p.fill(clr)
}

func (p *pen) circle(x, y, r float64, clr color.Color) {
	p.ellipse(x, y, r, r, clr)
}

// polygon fills the polygon through the points x0, y0, x1, y1 and so on.
func (p *pen) polygon(clr color.Color, xy ...float64) {
	for i := 0; i+1 < len(xy); i += 2 {
		x, y := p.at(xy[i], xy[i+1])
		if i == 0 {
			p.path.MoveTo(x, y)
		} else {
			p.path.LineTo(x, y)
		}
	}
	p.path.Close()
	p.fill(clr)
}

func (p *pen) fill(clr color.Color) {
	op := &vector.DrawPathOptions{AntiAlias: true}
	op.ColorScale.ScaleWithColor(clr)
	vector.FillPath(p.c.Image, &p.path, nil, op)
	p.path.Reset()
}
