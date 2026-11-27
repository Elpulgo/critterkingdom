// Package engine glues the game to Ebitengine: the game loop, the input
// abstraction, the touch controls, the camera and the low-resolution world
// view that is scaled up to the screen.
package engine

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
)

// Scene is one screen of the game, such as the meadow.
type Scene interface {
	// Update advances the scene by one tick.
	Update(f Frame) error
	// DrawWorld draws the world at world resolution. The image is View.W by
	// View.H pixels and is scaled up to the screen afterwards.
	DrawWorld(world *ebiten.Image)
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
	world            *ebiten.Image
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

// Draw draws the world at world resolution, scales it up to the screen and
// draws the interface on top.
func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(borderColor)

	if g.world == nil || g.world.Bounds().Dx() != g.view.W || g.world.Bounds().Dy() != g.view.H {
		if g.world != nil {
			g.world.Deallocate()
		}
		g.world = ebiten.NewImage(g.view.W, g.view.H)
	}
	g.world.Clear()
	g.scene.DrawWorld(g.world)
	if debugEnabled {
		g.debug.drawWorld(g.world, g.scene)
	}

	var op ebiten.DrawImageOptions
	op.GeoM.Scale(float64(g.view.Scale), float64(g.view.Scale))
	op.GeoM.Translate(float64(g.view.OffX), float64(g.view.OffY))
	screen.DrawImage(g.world, &op)

	g.scene.DrawUI(screen, UI{View: g.view, Unit: g.unit})
	g.input.touch.draw(screen, g.touchLayout())
	if debugEnabled {
		g.debug.draw(screen, g)
	}
}

func (g *Game) touchLayout() touchLayout {
	return touchLayout{screenW: float64(g.screenW), screenH: float64(g.screenH), unit: g.unit}
}
