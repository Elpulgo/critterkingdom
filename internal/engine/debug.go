package engine

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// debugScene is implemented by scenes that add their own lines and drawings
// to the debug overlay.
type debugScene interface {
	// DebugLines returns extra lines of text for the overlay.
	DebugLines() []string
	// DrawDebug draws on the world image after the scene has drawn the world.
	DrawDebug(world *ebiten.Image)
}

var debugBackdrop = color.NRGBA{0x18, 0x14, 0x25, 0xb0}

// debugOverlay shows the frame rate, the screen and world sizes and the
// latest input on top of the game. It is only shown in builds with the debug
// tag, and F3 hides and shows it.
type debugOverlay struct {
	hidden  bool
	face    *text.GoTextFace
	paws    int
	menus   int
	last    Input
	fingers int
	lines   []string
}

func (d *debugOverlay) update(in Input, fingers int) {
	if inpututil.IsKeyJustPressed(ebiten.KeyF3) {
		d.hidden = !d.hidden
	}
	if in.Paw {
		d.paws++
	}
	if in.Menu {
		d.menus++
	}
	d.last = in
	d.fingers = fingers
}

func (d *debugOverlay) drawWorld(world *ebiten.Image, scene Scene) {
	if d.hidden {
		return
	}
	if s, ok := scene.(debugScene); ok {
		s.DrawDebug(world)
	}
}

func (d *debugOverlay) draw(screen *ebiten.Image, g *Game) {
	if d.hidden || d.face == nil {
		return
	}
	in, v := d.last, g.view
	point := "-"
	if in.Pointing {
		point = fmt.Sprintf("%.0f,%.0f", in.Point.X, in.Point.Y)
	}
	d.lines = append(d.lines[:0],
		fmt.Sprintf("fps %.0f  tps %.0f", ebiten.ActualFPS(), ebiten.ActualTPS()),
		fmt.Sprintf("screen %dx%d  dpr %.2f", g.screenW, g.screenH, g.unit),
		fmt.Sprintf("world %dx%d  scale %d", v.W, v.H, v.Scale),
		fmt.Sprintf("move %+.2f,%+.2f  point %s", in.Move.X, in.Move.Y, point),
		fmt.Sprintf("paw %d  menu %d  fingers %d", d.paws, d.menus, d.fingers),
	)
	if s, ok := g.scene.(debugScene); ok {
		d.lines = append(d.lines, s.DebugLines()...)
	}

	d.face.Size = debugTextSize * g.unit
	lh := LineHeight(d.face)
	pad := 6 * g.unit
	w := 0.0
	for _, l := range d.lines {
		w = max(w, text.Advance(l, d.face))
	}
	h := lh * float64(len(d.lines))
	vector.FillRect(screen, 0, 0, float32(w+2*pad), float32(h+2*pad), debugBackdrop, false)
	for i, l := range d.lines {
		DrawText(screen, l, d.face, pad, pad+float64(i)*lh, text.AlignStart, color.White)
	}
}
