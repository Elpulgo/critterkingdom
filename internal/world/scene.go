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
	// bobPace is how fast the walk bob swings, in radians per second, so it
	// rises and falls once every two steps. It is also the fastest the bob
	// may change, in world pixels per second, which lets it follow the swing
	// and settle softly when the critter stops.
	bobPace = math.Pi / (2 * stepTime)
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
	meadow   *meadow
	player   *entity.Player
	camera   geom.Vec // top-left corner of the view, in world pixels
	viewSize geom.Vec // size of the view, in world pixels
	pen      pen
	bob      float64 // how far the walk bob lifts the critter, in world pixels

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
		meadow:     newMeadow(bounds.Max.Sub(bounds.Min), bounds.Inset(hedgeWidth+6)),
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
	s.updateBob(f.DT)

	s.viewSize = geom.Vec{X: float64(f.View.W), Y: float64(f.View.H)}
	focus := s.player.Pos.Sub(geom.Vec{Y: focusLift})
	s.camera = engine.SmoothCameraFor(focus, f.View.W, f.View.H, s.bounds, f.View.Scale)
	return nil
}

// updateBob swings the walk bob while the critter walks and lets it
// settle back to the ground when it stops.
func (s *Scene) updateBob(dt float64) {
	target := 0.0
	if s.player.Moved() {
		swing := math.Sin(bobPace * s.walked)
		target = swing * swing
	}
	s.bob += math.Max(-bobPace*dt, math.Min(bobPace*dt, target-s.bob))
}

// DrawWorld draws the meadow and the critter.
func (s *Scene) DrawWorld(c engine.Canvas) {
	c.Image.Fill(hedgeDark)
	s.pen.begin(c, s.camera)
	drawMeadow(&s.pen, s.meadow, geom.Rect{Min: s.camera, Max: s.camera.Add(s.viewSize)})
	feet := s.player.Pos
	drawCritter(&s.pen, feet.X, feet.Y, s.player.HopHeight()+s.bob, s.player.Facing)
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

	head := s.feetInView().Sub(geom.Vec{Y: critterHeight + s.player.HopHeight() + 1})
	hx, hy := ui.View.ViewToScreen(head)
	sw := float64(screen.Bounds().Dx())
	x := math.Max(w/2+margin, math.Min(hx, sw-w/2-margin))
	bottom := math.Max(hy-tail, h+margin)

	alpha := s.greet / greetFade
	fillBubble(screen, float32(x), float32(bottom), float32(w), float32(h), float32(12*u), float32(hx), float32(tail), fade(white, alpha))
	engine.DrawText(screen, greeting, s.bubbleFace, x, bottom-h+padY, text.AlignCenter, fade(cocoa, alpha))
}

// feetInView returns where the critter's feet are in the view.
func (s *Scene) feetInView() geom.Vec {
	return s.player.Pos.Sub(s.camera)
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
// walk target. Sizes are in world pixels, so the marks grow with the view.
func (s *Scene) DrawDebug(c engine.Canvas) {
	k := float32(c.Scale)
	left, top := c.At(s.walkable.Min.Sub(s.camera))
	size := s.walkable.Max.Sub(s.walkable.Min)
	vector.StrokeRect(c.Image, left, top, float32(size.X)*k, float32(size.Y)*k, k, debugColor, true)

	fx, fy := c.At(s.feetInView())
	vector.FillRect(c.Image, fx-k, fy-k, 2*k, 2*k, debugColor, true)

	if t, ok := s.player.Target(); ok {
		x, y := c.At(t.Sub(s.camera))
		vector.StrokeLine(c.Image, x-3*k, y-3*k, x+3*k, y+3*k, k, debugColor, true)
		vector.StrokeLine(c.Image, x-3*k, y+3*k, x+3*k, y-3*k, k, debugColor, true)
	}
}
