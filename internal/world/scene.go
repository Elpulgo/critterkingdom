// Package world is where the critter walks around Tiled maps.
package world

import (
	"fmt"
	"math"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"github.com/Elpulgo/critterkingdom/internal/engine"
	"github.com/Elpulgo/critterkingdom/internal/entity"
	"github.com/Elpulgo/critterkingdom/internal/geom"
)

const (
	// feetMarginX and feetMarginTop keep the critter's 12 by 10 feet box inside
	// the map bounds when its feet position is clamped to walkable.
	feetMarginX   = 6
	feetMarginTop = 10
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

// Scene is the world scene: a Tiled map where the critter walks around,
// collides with solid scenery, and hops and says hello when the paw button is
// pressed.
type Scene struct {
	m        *Map
	bounds   geom.Rect // the whole map, in world pixels
	walkable geom.Rect // where the critter's feet may go so its box stays inside bounds
	player   *entity.Player
	camera   geom.Vec // top-left corner of the view, in world pixels
	viewSize geom.Vec // size of the view, in world pixels
	pen      pen
	bob      float64 // how far the walk bob lifts the critter, in world pixels
	elapsed  float64 // total elapsed seconds, for tile animations

	walked     float64 // seconds the critter has spent walking
	greet      float64 // seconds left of the greeting
	titleAlpha float64

	titleFace  *text.GoTextFace
	bubbleFace *text.GoTextFace
}

// NewScene returns a world scene for m with the critter placed at the first
// player_spawn object in the objects layer. It returns an error when m has no
// player_spawn object.
func NewScene(font *text.GoTextFaceSource, m *Map) (*Scene, error) {
	if m == nil {
		return nil, fmt.Errorf("map is nil")
	}
	spawn, ok := findPlayerSpawn(m)
	if !ok {
		return nil, fmt.Errorf("map has no player_spawn object in objects layer")
	}
	size := m.Size()
	bounds := geom.Rect{Max: size}
	walkable := geom.Rect{
		Min: geom.Vec{X: feetMarginX, Y: feetMarginTop},
		Max: geom.Vec{X: size.X - feetMarginX, Y: size.Y},
	}
	player := entity.NewPlayer(spawn)
	player.Walls = m.Collision
	return &Scene{
		m:          m,
		bounds:     bounds,
		walkable:   walkable,
		player:     player,
		titleAlpha: 1,
		titleFace:  engine.NewFace(font, titleTextSize, titleWeight),
		bubbleFace: engine.NewFace(font, bubbleTextSize, bubbleWeight),
	}, nil
}

// findPlayerSpawn returns the position of the first object of type
// player_spawn in the objects layer of m, and false when there is none.
func findPlayerSpawn(m *Map) (geom.Vec, bool) {
	objs := m.ObjectLayer("objects")
	if objs == nil {
		return geom.Vec{}, false
	}
	for _, o := range objs.Objects {
		if o.Type == "player_spawn" {
			return geom.Vec{X: o.X, Y: o.Y}, true
		}
	}
	return geom.Vec{}, false
}

// Update moves the critter and the camera by one tick.
func (s *Scene) Update(f engine.Frame) error {
	s.elapsed += f.DT
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

// DrawWorld draws the map layers and the critter.
func (s *Scene) DrawWorld(c engine.Canvas) {
	c.Image.Fill(ink)
	s.pen.begin(c, s.camera)
	view := geom.Rect{Min: s.camera, Max: s.camera.Add(s.viewSize)}
	c0, r0, c1, r1 := s.m.VisibleCells(view)
	elapsed := time.Duration(s.elapsed * float64(time.Second))

	drawTileLayer(&s.pen, s.m, s.m.TileLayer("ground"), c0, r0, c1, r1, elapsed)
	drawTileLayer(&s.pen, s.m, s.m.TileLayer("decor"), c0, r0, c1, r1, elapsed)
	drawProps(&s.pen, s.m, s.m.ObjectLayer("props"), view, elapsed)
	feet := s.player.Pos
	drawCritter(&s.pen, feet.X, feet.Y, s.player.HopHeight()+s.bob, s.player.Facing)
	drawTileLayer(&s.pen, s.m, s.m.TileLayer("above"), c0, r0, c1, r1, elapsed)
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
