// Package world is where the critter walks around Tiled maps.
package world

import (
	"cmp"
	"fmt"
	"math"
	"math/rand/v2"
	"slices"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"github.com/Elpulgo/critterkingdom/internal/engine"
	"github.com/Elpulgo/critterkingdom/internal/entity"
	"github.com/Elpulgo/critterkingdom/internal/geom"
	"github.com/Elpulgo/critterkingdom/internal/i18n"
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

// The tap box of a villager, in world pixels from its feet: a press inside it
// taps the villager rather than the ground behind it.
const (
	tapHalfWidth = 10
	tapAbove     = 24
	tapBelow     = 2
)

// Interaction is an interaction the critter started with a villager. NPC is
// the villager's id.
type Interaction struct {
	NPC string
}

// Scene is the world scene: a Tiled map where the critter walks around,
// collides with solid scenery, talks to the villager in front of it or hops
// and says hello when the paw button is pressed, and moves to another map
// when it walks into a warp.
type Scene struct {
	load     MapLoader
	mapName  string // the name m was loaded under
	m        *Map
	bounds   geom.Rect // the whole map, in world pixels
	walkable geom.Rect // where the critter's feet may go so its box stays inside bounds
	player   *entity.Player
	npcs     []*entity.NPC // the villagers of m, in the order of m.NPCs
	camera   geom.Vec      // top-left corner of the view, in world pixels
	viewSize geom.Vec      // size of the view, in world pixels
	pen      pen
	bob      float64 // how far the walk bob lifts the critter, in world pixels
	elapsed  float64 // total elapsed seconds, for tile animations

	// awaitRelease ignores pointer walking until the pointer has been let go:
	// after a warp, so a held tap does not walk the critter straight back, and
	// after a press on a villager, which walks to the villager instead.
	awaitRelease bool
	// wasPointing is whether the pointer was held on the tick before, so a
	// press can be told from a hold.
	wasPointing bool
	// tapTarget is the villager the critter walks to because it was tapped,
	// or nil. Reaching it starts an interaction.
	tapTarget *entity.NPC

	// pending is the interaction started and not yet taken with
	// TakeInteraction, or nil. lastNPC is the id of the villager last
	// interacted with, for the debug overlay.
	pending *Interaction
	lastNPC string

	walked     float64 // seconds the critter has spent walking
	greet      float64 // seconds left of the greeting
	titleAlpha float64

	titleFace  *text.GoTextFace
	bubbleFace *text.GoTextFace
}

// NewScene returns a world scene on the map startMap, loaded with load, with
// the critter placed at the map's first player_spawn. Warps load their target
// maps with load too. It returns an error when startMap cannot be loaded or
// has no player_spawn object.
func NewScene(font *text.GoTextFaceSource, load MapLoader, startMap string) (*Scene, error) {
	if load == nil {
		return nil, fmt.Errorf("map loader is nil")
	}
	m, err := load(startMap)
	if err != nil {
		return nil, fmt.Errorf("start map %q: %w", startMap, err)
	}
	if m == nil {
		return nil, fmt.Errorf("start map %q: loader returned no map", startMap)
	}
	spawn, ok := m.DefaultSpawn()
	if !ok {
		return nil, fmt.Errorf("start map %q has no player_spawn object in %s layer", startMap, objectsLayer)
	}
	s := &Scene{
		load:       load,
		player:     entity.NewPlayer(spawn),
		titleAlpha: 1,
		titleFace:  engine.NewFace(font, titleTextSize, titleWeight),
		bubbleFace: engine.NewFace(font, bubbleTextSize, bubbleWeight),
	}
	s.setMap(startMap, m)
	return s, nil
}

// setMap makes m, loaded as name, the scene's map: its size bounds the
// critter and the camera, its collision grid stops the critter and the
// villagers, and its npc objects replace the villagers of the map before.
func (s *Scene) setMap(name string, m *Map) {
	size := m.Size()
	s.mapName = name
	s.m = m
	s.bounds = geom.Rect{Max: size}
	s.walkable = geom.Rect{
		Min: geom.Vec{X: feetMarginX, Y: feetMarginTop},
		Max: geom.Vec{X: size.X - feetMarginX, Y: size.Y},
	}
	s.player.Walls = m.Collision
	// A tapped villager belongs to the map before.
	s.tapTarget = nil
	spawns := m.NPCs()
	s.npcs = make([]*entity.NPC, 0, len(spawns))
	for _, sp := range spawns {
		// Seeding from the object id gives each villager its own wander that
		// is the same every time the map is entered.
		rng := rand.New(rand.NewPCG(uint64(sp.ObjectID), 0))
		n := entity.NewNPC(sp.ID, sp.Pos, sp.Area, rng)
		n.Walls = m.Collision
		s.npcs = append(s.npcs, n)
	}
}

// followWarp moves the critter to the target spawn of the warp its feet are
// in, if any, on the warp's target map. It returns the error from ResolveWarp
// when the target map or spawn does not exist.
func (s *Scene) followWarp() error {
	w, ok := s.m.WarpAt(s.player.Pos)
	if !ok {
		return nil
	}
	m, pos, err := ResolveWarp(s.load, w)
	if err != nil {
		return err
	}
	s.setMap(w.TargetMap, m)
	s.player.Place(pos)
	s.awaitRelease = true
	return nil
}

// Update moves the critter, the villagers and the camera by one tick, starts
// an interaction when the critter reaches a villager it was tapped towards,
// and moves the critter to another map when its feet end the tick inside a
// warp.
// It returns an error when that warp's target map or spawn does not exist.
func (s *Scene) Update(f engine.Frame) error {
	s.elapsed += f.DT
	in := f.Input
	pressed := in.Pointing && !s.wasPointing
	s.wasPointing = in.Pointing
	if !in.Pointing {
		s.awaitRelease = false
	}
	s.steer(in, pressed)
	if in.Paw {
		s.paw()
	}
	s.player.Update(f.DT, in.Move, s.walkable)
	if err := s.followWarp(); err != nil {
		return err
	}
	if s.tapTarget != nil && !s.reachTapTarget() {
		if _, walking := s.player.Target(); !walking {
			// The walk ended before getting there: a wall stopped it, or the
			// stick or a warp ended it.
			s.tapTarget = nil
		}
	}
	for _, n := range s.npcs {
		n.Update(f.DT, s.player.Pos)
	}
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

// steer turns this tick's movement input into where the critter goes. The
// stick or keys win and drop a tapped villager. A press on a villager makes
// it the tap target, which the critter walks to until it is in reach; a
// press anywhere else drops the target and, like a held pointer, walks the
// critter to the point.
func (s *Scene) steer(in engine.Input, pressed bool) {
	if !in.Move.IsZero() {
		s.tapTarget = nil
		return
	}
	if pressed && !s.awaitRelease {
		point := s.camera.Add(in.Point)
		s.tapTarget = s.villagerAt(point)
		if s.tapTarget != nil {
			// The rest of this press walks to the villager, not the point.
			s.awaitRelease = true
		}
	}
	if s.tapTarget != nil {
		if !s.reachTapTarget() {
			s.player.WalkTo(s.tapTarget.Pos)
		}
		return
	}
	if in.Pointing && !s.awaitRelease {
		s.player.WalkTo(s.camera.Add(in.Point))
	}
}

// villagerAt returns the villager whose tap box holds the world point p, the
// one standing lowest in the view when boxes overlap, as it is drawn on top,
// or nil when there is none.
func (s *Scene) villagerAt(p geom.Vec) *entity.NPC {
	var hit *entity.NPC
	for _, n := range s.npcs {
		box := geom.Rect{
			Min: geom.Vec{X: n.Pos.X - tapHalfWidth, Y: n.Pos.Y - tapAbove},
			Max: geom.Vec{X: n.Pos.X + tapHalfWidth, Y: n.Pos.Y + tapBelow},
		}
		if box.Contains(p) && (hit == nil || n.Pos.Y >= hit.Pos.Y) {
			hit = n
		}
	}
	return hit
}

// reachTapTarget checks whether the critter's feet are within reach of the
// tapped villager. If so, the critter stops, faces it and starts an
// interaction with it, the target is dropped and it returns true.
func (s *Scene) reachTapTarget() bool {
	n := s.tapTarget
	off := n.Pos.Sub(s.player.Pos)
	if off.Len() > entity.InteractReach {
		return false
	}
	s.player.Place(s.player.Pos)
	s.player.Facing = s.player.Facing.Toward(off)
	s.interact(n)
	s.tapTarget = nil
	return true
}

// paw starts an interaction with the villager the critter would pick, or
// hops and greets when there is none.
func (s *Scene) paw() {
	if i, ok := s.pickVillager(); ok {
		s.interact(s.npcs[i])
		return
	}
	if s.player.Hop() {
		s.greet = greetTime
	}
}

// pickVillager returns the index in s.npcs of the villager the paw would
// interact with now, and false when there is none.
func (s *Scene) pickVillager() (int, bool) {
	if len(s.npcs) == 0 {
		return 0, false
	}
	feet := make([]geom.Vec, len(s.npcs))
	for i, n := range s.npcs {
		feet[i] = n.Pos
	}
	return entity.PickInteractable(s.player.Pos, s.player.Facing, feet)
}

// interact starts an interaction with n, replacing one not yet taken.
func (s *Scene) interact(n *entity.NPC) {
	s.pending = &Interaction{NPC: n.ID}
	s.lastNPC = n.ID
}

// TakeInteraction returns the interaction the critter started since the last
// call, and false when it started none.
func (s *Scene) TakeInteraction() (Interaction, bool) {
	if s.pending == nil {
		return Interaction{}, false
	}
	in := *s.pending
	s.pending = nil
	return in, true
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

// DrawWorld draws the map layers, the villagers and the critter.
func (s *Scene) DrawWorld(c engine.Canvas) {
	c.Image.Fill(ink)
	s.pen.begin(c, s.camera)
	view := geom.Rect{Min: s.camera, Max: s.camera.Add(s.viewSize)}
	c0, r0, c1, r1 := s.m.VisibleCells(view)
	elapsed := time.Duration(s.elapsed * float64(time.Second))

	drawTileLayer(&s.pen, s.m, s.m.TileLayer("ground"), c0, r0, c1, r1, elapsed)
	drawTileLayer(&s.pen, s.m, s.m.TileLayer("decor"), c0, r0, c1, r1, elapsed)
	drawProps(&s.pen, s.m, s.m.ObjectLayer("props"), view, elapsed)
	s.drawCharacters()
	drawTileLayer(&s.pen, s.m, s.m.TileLayer("above"), c0, r0, c1, r1, elapsed)
}

// drawCharacters draws the villagers and the critter from the back of the
// view to the front, by the Y of their feet, so whoever stands lower in the
// view overlaps whoever stands behind. The critter is drawn last of those
// with the same Y.
func (s *Scene) drawCharacters() {
	order := make([]int, len(s.npcs))
	for i := range order {
		order[i] = i
	}
	slices.SortStableFunc(order, func(a, b int) int {
		return cmp.Compare(s.npcs[a].Pos.Y, s.npcs[b].Pos.Y)
	})
	feet := s.player.Pos
	critterDrawn := false
	for _, i := range order {
		n := s.npcs[i]
		if !critterDrawn && n.Pos.Y > feet.Y {
			drawCritter(&s.pen, feet.X, feet.Y, s.player.HopHeight()+s.bob, s.player.Facing)
			critterDrawn = true
		}
		drawVillager(&s.pen, n.Pos.X, n.Pos.Y, n.Facing, i)
	}
	if !critterDrawn {
		drawCritter(&s.pen, feet.X, feet.Y, s.player.HopHeight()+s.bob, s.player.Facing)
	}
}

// DrawUI draws the title, the "…" bubble over the villager the paw would
// pick and the greeting bubble at full screen resolution, in the current
// language of i18n.
func (s *Scene) DrawUI(screen *ebiten.Image, ui engine.UI) {
	if s.titleAlpha > 0 {
		s.drawTitle(screen, ui)
	}
	if i, ok := s.pickVillager(); ok {
		s.drawTalkBubble(screen, ui, s.npcs[i])
	}
	if s.greet > 0 {
		s.drawGreeting(screen, ui)
	}
}

func (s *Scene) drawTitle(screen *ebiten.Image, ui engine.UI) {
	u := ui.Unit
	title := i18n.T("ui.title")
	s.titleFace.Size = titleTextSize * u
	x, y := float64(screen.Bounds().Dx())/2, 24*u
	engine.DrawText(screen, title, s.titleFace, x+3*u, y+3*u, text.AlignCenter, fade(cocoa, 0.6*s.titleAlpha))
	engine.DrawText(screen, title, s.titleFace, x, y, text.AlignCenter, fade(white, s.titleAlpha))
}

// drawGreeting draws a speech bubble above the critter's head, kept on the
// screen when the critter is near an edge.
func (s *Scene) drawGreeting(screen *ebiten.Image, ui engine.UI) {
	u := ui.Unit
	greeting := i18n.T("ui.greeting")
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

// drawTalkBubble draws a small bubble with three dots above n's head, kept
// on the screen when n is near an edge. The dots are drawn rather than
// written, so the bubble needs no glyph or text.
func (s *Scene) drawTalkBubble(screen *ebiten.Image, ui engine.UI, n *entity.NPC) {
	u := ui.Unit
	w, h, tail, margin := 44*u, 22*u, 7*u, 8*u
	dot, gap := 3*u, 10*u

	head := n.Pos.Sub(s.camera).Sub(geom.Vec{Y: villagerHeight + 1})
	hx, hy := ui.View.ViewToScreen(head)
	sw := float64(screen.Bounds().Dx())
	x := math.Max(w/2+margin, math.Min(hx, sw-w/2-margin))
	bottom := math.Max(hy-tail, h+margin)

	fillBubble(screen, float32(x), float32(bottom), float32(w), float32(h), float32(10*u), float32(hx), float32(tail), white)
	cy := float32(bottom - h/2)
	for k := -1.0; k <= 1; k++ {
		vector.FillCircle(screen, float32(x+k*gap), cy, float32(dot), cocoa, true)
	}
}

// feetInView returns where the critter's feet are in the view.
func (s *Scene) feetInView() geom.Vec {
	return s.player.Pos.Sub(s.camera)
}

// DebugLines adds the map's name, the critter's position, the camera, the
// number of villagers, the villager last interacted with and the walk target
// to the debug overlay.
func (s *Scene) DebugLines() []string {
	p := s.player
	lines := []string{
		"map " + s.mapName,
		fmt.Sprintf("critter %.1f,%.1f facing %s", p.Pos.X, p.Pos.Y, p.Facing),
		fmt.Sprintf("camera %.0f,%.0f", s.camera.X, s.camera.Y),
		fmt.Sprintf("villagers %d", len(s.npcs)),
	}
	if s.lastNPC != "" {
		lines = append(lines, "talked to "+s.lastNPC)
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
