// Package world is where the critter walks. For now that is a single meadow;
// maps made in Tiled replace it later.
package world

import (
	"fmt"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"github.com/Elpulgo/critterkingdom/internal/engine"
	"github.com/Elpulgo/critterkingdom/internal/entity"
	"github.com/Elpulgo/critterkingdom/internal/geom"
)

const (
	tileSize   = 16
	meadowCols = 40
	meadowRows = 24
	hedgeWidth = 16
	// edgeGap keeps the critter's feet this far from the edge of the meadow,
	// so it stands in front of the hedge rather than inside it.
	edgeGap = hedgeWidth + 4
	// focusLift raises the camera's focus from the critter's feet to the
	// middle of its body.
	focusLift = 8
	// stepTime is how long each step of the walk bob lasts, in seconds.
	stepTime = 0.15
)

const (
	title          = "Critter Kingdom"
	greeting       = "Hej!"
	titleTextSize  = 34 // CSS pixels
	titleWeight    = 600
	bubbleTextSize = 26 // CSS pixels
	bubbleWeight   = 500
	// greetTime is how long the greeting stays up after a hop, in seconds.
	greetTime = 1.5
	// greetFade is how long the greeting takes to fade out, in seconds.
	greetFade = 0.2
	// titleWalk is how long the critter walks before the title fades, in
	// seconds, and titleFade is how long the fade takes.
	titleWalk = 0.6
	titleFade = 1.0
)

// Scene is the meadow: a field of grass inside a hedge where the critter
// walks around, and hops and says hello when the paw button is pressed.
type Scene struct {
	bounds   geom.Rect // the whole meadow, in world pixels
	walkable geom.Rect // where the critter's feet may go
	ground   *ebiten.Image
	player   *entity.Player
	camera   geom.Vec // top-left corner of the view, in world pixels

	walked     float64 // seconds the critter has spent walking
	greet      float64 // seconds left of the greeting
	titleAlpha float64

	titleFace  *text.GoTextFace
	bubbleFace *text.GoTextFace
}

// NewScene returns the meadow with the critter in the middle.
func NewScene(font *text.GoTextFaceSource) *Scene {
	bounds := geom.Rect{Max: geom.Vec{X: meadowCols * tileSize, Y: meadowRows * tileSize}}
	return &Scene{
		bounds:     bounds,
		walkable:   bounds.Inset(edgeGap),
		player:     entity.NewPlayer(geom.Vec{X: bounds.Max.X / 2, Y: bounds.Max.Y / 2}),
		titleAlpha: 1,
		titleFace:  engine.NewFace(font, titleTextSize, titleWeight),
		bubbleFace: engine.NewFace(font, bubbleTextSize, bubbleWeight),
	}
}

// Update moves the critter and the camera by one tick.
func (s *Scene) Update(f engine.Frame) error {
	in := f.Input
	if in.Move.IsZero() && in.Pointing {
		s.player.WalkTo(s.camera.Add(in.Point))
	}
	if in.Paw && s.player.Hop() {
		s.greet = greetTime
	}
	s.player.Update(f.DT, in.Move, s.walkable)
	if s.player.Moved() {
		s.walked += f.DT
	}
	s.greet = math.Max(0, s.greet-f.DT)
	if s.walked >= titleWalk {
		s.titleAlpha = math.Max(0, s.titleAlpha-f.DT/titleFade)
	}
	focus := s.player.Pos.Sub(geom.Vec{Y: focusLift})
	s.camera = engine.CameraFor(focus, f.View.W, f.View.H, s.bounds)
	return nil
}

// DrawWorld draws the meadow and the critter.
func (s *Scene) DrawWorld(world *ebiten.Image) {
	if s.ground == nil {
		s.ground = ebiten.NewImage(int(s.bounds.Max.X-s.bounds.Min.X), int(s.bounds.Max.Y-s.bounds.Min.Y))
		drawMeadow(s.ground, s.bounds.Inset(hedgeWidth+6))
	}
	world.Fill(hedgeDark)
	var op ebiten.DrawImageOptions
	op.GeoM.Translate(s.bounds.Min.X-s.camera.X, s.bounds.Min.Y-s.camera.Y)
	world.DrawImage(s.ground, &op)

	feet := s.feetInView()
	lift := math.Round(s.player.HopHeight())
	if s.player.Moved() && int(s.walked/stepTime)%2 == 1 {
		lift++
	}
	drawCritter(world, float32(feet.X), float32(feet.Y), float32(lift), s.player.Facing)
}

// DrawUI draws the title and the greeting bubble at full screen resolution.
func (s *Scene) DrawUI(screen *ebiten.Image, ui engine.UI) {
	if s.titleAlpha > 0 {
		s.drawTitle(screen, ui)
	}
	if s.greet > 0 {
		s.drawGreeting(screen, ui)
	}
}

func (s *Scene) drawTitle(screen *ebiten.Image, ui engine.UI) {
	u := ui.Unit
	s.titleFace.Size = titleTextSize * u
	x, y := float64(screen.Bounds().Dx())/2, 24*u
	engine.DrawText(screen, title, s.titleFace, x+3*u, y+3*u, text.AlignCenter, fade(cocoa, 0.6*s.titleAlpha))
	engine.DrawText(screen, title, s.titleFace, x, y, text.AlignCenter, fade(white, s.titleAlpha))
}

// drawGreeting draws a speech bubble above the critter's head, kept on the
// screen when the critter is near an edge.
func (s *Scene) drawGreeting(screen *ebiten.Image, ui engine.UI) {
	u := ui.Unit
	s.bubbleFace.Size = bubbleTextSize * u
	padX, padY, tail, margin := 14*u, 6*u, 10*u, 8*u
	w := text.Advance(greeting, s.bubbleFace) + 2*padX
	h := engine.LineHeight(s.bubbleFace) + 2*padY

	head := s.feetInView().Sub(geom.Vec{Y: critterHeight + math.Round(s.player.HopHeight()) + 1})
	hx, hy := ui.View.ViewToScreen(head)
	sw := float64(screen.Bounds().Dx())
	x := math.Max(w/2+margin, math.Min(hx, sw-w/2-margin))
	bottom := math.Max(hy-tail, h+margin)

	alpha := s.greet / greetFade
	fillBubble(screen, float32(x), float32(bottom), float32(w), float32(h), float32(12*u), float32(hx), float32(tail), fade(white, alpha))
	engine.DrawText(screen, greeting, s.bubbleFace, x, bottom-h+padY, text.AlignCenter, fade(cocoa, alpha))
}

// feetInView returns where the critter's feet are in the view, in whole
// pixels.
func (s *Scene) feetInView() geom.Vec {
	p := s.player.Pos.Sub(s.camera)
	return geom.Vec{X: math.Floor(p.X), Y: math.Floor(p.Y)}
}

// DebugLines adds the critter's position, the camera and the walk target to
// the debug overlay.
func (s *Scene) DebugLines() []string {
	p := s.player
	lines := []string{
		fmt.Sprintf("critter %.1f,%.1f facing %s", p.Pos.X, p.Pos.Y, p.Facing),
		fmt.Sprintf("camera %.0f,%.0f", s.camera.X, s.camera.Y),
	}
	if t, ok := p.Target(); ok {
		lines = append(lines, fmt.Sprintf("walk to %.0f,%.0f", t.X, t.Y))
	}
	return lines
}

// DrawDebug outlines where the critter may walk and marks its feet and its
// walk target.
func (s *Scene) DrawDebug(world *ebiten.Image) {
	topLeft := s.walkable.Min.Sub(s.camera)
	size := s.walkable.Max.Sub(s.walkable.Min)
	vector.StrokeRect(world, float32(topLeft.X), float32(topLeft.Y), float32(size.X), float32(size.Y), 1, debugColor, false)

	feet := s.feetInView()
	vector.FillRect(world, float32(feet.X)-1, float32(feet.Y)-1, 2, 2, debugColor, false)

	if t, ok := s.player.Target(); ok {
		c := t.Sub(s.camera)
		x, y := float32(c.X), float32(c.Y)
		vector.StrokeLine(world, x-3, y-3, x+3, y+3, 1, debugColor, false)
		vector.StrokeLine(world, x-3, y+3, x+3, y-3, 1, debugColor, false)
	}
}
