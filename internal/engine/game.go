// Package engine glues the game to Ebitengine: the game loop, the input
// abstraction, the touch controls, the camera and the view that places the
// world on the screen.
package engine

import (
	"image"
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"

	"github.com/Elpulgo/critterkingdom/internal/geom"
)

// Scene is one screen of the game, such as the meadow.
type Scene interface {
	// Update advances the scene by one tick.
	Update(f Frame) error
	// DrawWorld draws the world on the canvas.
	DrawWorld(c Canvas)
	// DrawUI draws text and other interface over the world at full screen
	// resolution, so text stays sharp.
	DrawUI(screen *ebiten.Image, ui UI)
}

// Frame is what a scene gets every tick.
type Frame struct {
	Input Input
	View  View
	// DT is the length of a tick, in seconds.
	DT float64
}

// Canvas is the part of the screen that shows the world. Scenes draw on it
// at full screen resolution.
type Canvas struct {
	Image *ebiten.Image
	// Scale is how many screen pixels one world pixel covers.
	Scale float64
	// OffX and OffY are where the top-left corner of the view lands on Image.
	OffX, OffY float64
}

// At returns where the point p, in view pixels, lands on Image.
func (c Canvas) At(p geom.Vec) (x, y float32) {
	return float32(p.X*c.Scale + c.OffX), float32(p.Y*c.Scale + c.OffY)
}

// UI tells a scene how to place its interface on the screen.
type UI struct {
	View View
	// Unit is the number of screen pixels per CSS pixel. Interface sizes are
	// given in CSS pixels and multiplied by Unit, so a button or a line of
	// text has the same physical size on every screen.
	Unit float64
}

// Options configure a Game.
type Options struct {
	// Font is used by the debug overlay.
	Font *text.GoTextFaceSource
	// TouchPad turns on the floating touch pad on the left half of the
	// screen. Without it, a finger anywhere on the world walks the critter
	// there.
	TouchPad bool
}

// borderColor fills the screen around the world on very wide or tall screens.
var borderColor = color.NRGBA{0x18, 0x14, 0x25, 0xff}

// debugTextSize is the size of the debug overlay text, in CSS pixels.
const debugTextSize = 13

// Game runs a scene. It implements ebiten.Game.
type Game struct {
	scene Scene
	input inputReader
	debug debugOverlay

	screenW, screenH int
	unit             float64
	view             View
}

// NewGame returns a game that runs scene.
func NewGame(scene Scene, opts Options) *Game {
	g := &Game{
		scene: scene,
		input: inputReader{touch: touchControls{padEnabled: opts.TouchPad}},
		unit:  1,
		view:  viewFor(minWorldWidth, minWorldHeight),
	}
	if opts.Font != nil {
		g.debug.face = NewFace(opts.Font, debugTextSize, 500)
	}
	return g
}

// Layout is never called, because Game implements LayoutF. It is here to
// satisfy ebiten.Game.
func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	w, h := g.LayoutF(float64(outsideWidth), float64(outsideHeight))
	return int(w), int(h)
}

// LayoutF makes the screen image as large as the window in device pixels, so
// text and interface are drawn at full resolution, and picks the world view
// for that size.
func (g *Game) LayoutF(outsideWidth, outsideHeight float64) (float64, float64) {
	unit := 1.0
	if m := ebiten.Monitor(); m != nil && m.DeviceScaleFactor() > 0 {
		unit = m.DeviceScaleFactor()
	}
	w := math.Max(1, math.Ceil(outsideWidth*unit))
	h := math.Max(1, math.Ceil(outsideHeight*unit))
	g.screenW, g.screenH = int(w), int(h)
	g.unit = unit
	g.view = viewFor(g.screenW, g.screenH)
	return w, h
}

// Update reads the input and advances the scene by one tick.
func (g *Game) Update() error {
	in := g.input.read(g.view, g.touchLayout())
	if debugEnabled {
		g.debug.update(in, g.input.touch.count())
	}
	tps := ebiten.TPS()
	if tps <= 0 {
		tps = ebiten.DefaultTPS
	}
	return g.scene.Update(Frame{Input: in, View: g.view, DT: 1 / float64(tps)})
}

// Draw draws the world and then the interface on top, both at full screen
// resolution.
func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(borderColor)
	g.drawWorld(screen)
	g.scene.DrawUI(screen, UI{View: g.view, Unit: g.unit})
	g.input.touch.draw(screen, g.touchLayout())
	if debugEnabled {
		g.debug.draw(screen, g)
	}
}

// drawWorld lets the scene draw on the part of the screen that shows the
// world, so nothing spills onto the border.
func (g *Game) drawWorld(screen *ebiten.Image) {
	v := g.view
	area := image.Rect(v.OffX, v.OffY, v.OffX+v.W*v.Scale, v.OffY+v.H*v.Scale)
	c := Canvas{
		Image: screen.SubImage(area).(*ebiten.Image),
		Scale: float64(v.Scale),
		OffX:  float64(v.OffX),
		OffY:  float64(v.OffY),
	}
	g.scene.DrawWorld(c)
	if debugEnabled {
		g.debug.drawWorld(c, g.scene)
	}
}

func (g *Game) touchLayout() touchLayout {
	return touchLayout{screenW: float64(g.screenW), screenH: float64(g.screenH), unit: g.unit}
}
