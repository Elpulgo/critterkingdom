// Package world is where the critter walks around Tiled maps.
package world

import (
	"cmp"
	"fmt"
	"image/color"
	"math"
	"math/rand/v2"
	"slices"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"github.com/Elpulgo/critterkingdom/internal/combat"
	"github.com/Elpulgo/critterkingdom/internal/critter"
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
	// lanternRadius is how far the lantern the critter carries on dark maps
	// reaches, in world pixels.
	lanternRadius = 56
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
// and says hello when the paw button is pressed, moves to another map when it
// walks into a warp, and meets the map's enemies. When enemy hazards take its
// last petal it naps: it wakes at the spot it entered the map at with full
// petals, and the enemies reset.
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
	elapsed  float64 // total elapsed seconds, for tile animations and the clock

	// clock returns elapsed as game time; health and every enemy read it.
	clock     combat.Clock
	health    *combat.Health
	enemyDefs map[string]combat.EnemyDef
	onDefeat  func(combat.EnemyDefeated)
	enemies   []*entity.Enemy
	// entrance is where the critter was placed on entering the current map,
	// where it wakes from a nap.
	entrance geom.Vec

	// awaitRelease ignores pointer walking until the pointer has been let go:
	// after a warp, so a held tap does not walk the critter straight back, and
	// after a press on a villager, which walks to the villager instead.
	awaitRelease bool
	// planned is true while plannedFor holds the cell the current pointer
	// route was planned for, and toTap whether that route ends on the
	// pointer rather than on the nearest reachable cell.
	planned    bool
	plannedFor Cell
	toTap      bool
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

	// species and rng are set by AddWildCritters; with rng nil the scene has
	// no wild critters.
	species []critter.Species
	rng     *rand.Rand
	// spawner decides which critters are present in the current map's
	// zones, and wild holds the entity of each, in zone order.
	spawner *critter.Spawner
	wild    []*wildCritter
	// inView is reused by DrawWorld for the wild critters it draws.
	inView []*wildCritter
	// timeOfDay decides which species are about; it stays critter.Day until
	// day and night exist.
	timeOfDay critter.TimeOfDay
	mask      lightMask // darkness drawn over dark maps
}

// NewScene returns a world scene on the map startMap, loaded with load, with
// the critter placed at the map's first player_spawn. Warps load their target
// maps with load too. opts add enemy definitions and a defeat sink. It
// returns an error when startMap cannot be loaded, has no player_spawn object,
// or has an enemy object naming an enemy without a definition.
func NewScene(font *text.GoTextFaceSource, load MapLoader, startMap string, opts ...Option) (*Scene, error) {
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
		enemyDefs:  make(map[string]combat.EnemyDef),
		titleAlpha: 1,
		timeOfDay:  critter.Day,
		titleFace:  engine.NewFace(font, titleTextSize, titleWeight),
		bubbleFace: engine.NewFace(font, bubbleTextSize, bubbleWeight),
	}
	s.clock = s.now
	s.health = combat.NewHealth(s.clock)
	for _, opt := range opts {
		opt(s)
	}
	enemies, err := s.spawnEnemies(m)
	if err != nil {
		return nil, fmt.Errorf("start map %q: %w", startMap, err)
	}
	s.setMap(startMap, m, spawn, enemies)
	return s, nil
}

// now returns the scene's game time: the sum of every tick's Frame.DT,
// rounded to the nanosecond.
func (s *Scene) now() time.Duration {
	return time.Duration(math.Round(s.elapsed * float64(time.Second)))
}

// spawnEnemies builds the enemies of m's enemy objects, each at its start
// position with full hit points, reading the scene's clock. It returns an
// error naming the object id and the enemy id for an enemy without a
// definition.
func (s *Scene) spawnEnemies(m *Map) ([]*entity.Enemy, error) {
	var enemies []*entity.Enemy
	for _, mk := range m.Enemies() {
		def, ok := s.enemyDefs[mk.EnemyID]
		if !ok {
			return nil, fmt.Errorf("%s object %d: no enemy definition %q", objectsLayer, mk.ObjectID, mk.EnemyID)
		}
		e, err := entity.NewEnemy(def, mk.Pos, s.clock, s.onDefeat)
		if err != nil {
			return nil, fmt.Errorf("%s object %d: %w", objectsLayer, mk.ObjectID, err)
		}
		e.Walls = m.Collision
		enemies = append(enemies, e)
	}
	return enemies, nil
}

// setMap makes m, loaded as name, the scene's map, with the critter entering
// it at entrance and enemies as its enemies: its size bounds the critter, the
// enemies and the camera, its collision grid stops them and the villagers, and
// its npc objects replace the villagers of the map before.
func (s *Scene) setMap(name string, m *Map, entrance geom.Vec, enemies []*entity.Enemy) {
	size := m.Size()
	s.mapName = name
	s.m = m
	s.bounds = geom.Rect{Max: size}
	s.walkable = geom.Rect{
		Min: geom.Vec{X: feetMarginX, Y: feetMarginTop},
		Max: geom.Vec{X: size.X - feetMarginX, Y: size.Y},
	}
	s.player.Walls = m.Collision
	s.player.Place(entrance)
	s.entrance = entrance
	s.enemies = enemies
	s.resetWildCritters()
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
// in, if any, on the warp's target map, with that map's enemies. It returns
// the error from ResolveWarp when the target map or spawn does not exist, and
// an error naming the map when one of its enemies has no definition.
func (s *Scene) followWarp() error {
	w, ok := s.m.WarpAt(s.player.Pos)
	if !ok {
		return nil
	}
	m, pos, err := ResolveWarp(s.load, w)
	if err != nil {
		return err
	}
	enemies, err := s.spawnEnemies(m)
	if err != nil {
		return fmt.Errorf("warp to map %q: %w", w.TargetMap, err)
	}
	s.setMap(w.TargetMap, m, pos, enemies)
	s.awaitRelease = true
	s.planned = false
	return nil
}

// Update moves the critter, the villagers, the enemies and the camera by one
// tick, starts an interaction when the critter reaches a villager it was
// tapped towards, and moves the critter to another map when its feet end the
// tick inside a warp. Enemy hazards touching the critter take its petals, and
// the tick they take the last one the critter naps. It returns an error when
// that warp's target map or spawn does not exist, or the map has an enemy
// without a definition.
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
	s.updateWildCritters(f.DT)
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
	s.updateEnemies()
	if s.health.Napping() {
		if err := s.nap(); err != nil {
			return err
		}
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

// walkToPointer walks the critter to p, the pointer's world position, around
// solid scenery. It plans a new route when the critter is not walking or p is
// in another cell than the current route was planned for; otherwise only the
// route's end follows p, when the route ends on the pointer, moved to where
// the critter's box fits in p's cell.
func (s *Scene) walkToPointer(p geom.Vec) {
	tap := s.walkable.Clamp(p)
	g := s.m.Collision
	if g == nil || g.TileSize <= 0 {
		s.player.WalkTo(tap)
		return
	}
	box := s.player.Box()
	r := router{grid: g, box: box.Max.Sub(box.Min)}
	cell := g.cellAt(tap)
	if _, walking := s.player.Target(); walking && s.planned && cell == s.plannedFor {
		if s.toTap {
			// The route ends in this open cell; end it where the box fits,
			// as plan does.
			s.player.Retarget(r.fit(tap))
		}
		return
	}
	route, toTap := r.plan(s.player.Pos, tap)
	s.player.WalkAlong(route)
	s.planned, s.plannedFor, s.toTap = true, cell, toTap
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
		s.walkToPointer(s.camera.Add(in.Point))
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

// updateEnemies steps every enemy toward the critter, lets each hazard that
// touches the critter's hurt-box hit it for the enemy's damage, and drops the
// enemies that have run off after their defeat.
func (s *Scene) updateEnemies() {
	hurt := s.player.HurtBox()
	kept := s.enemies[:0]
	for _, e := range s.enemies {
		e.Update(s.player.Pos, s.walkable)
		for _, h := range e.Hazards() {
			if combat.Overlap(h, hurt) {
				s.health.Hit(e.Def().Damage)
			}
		}
		if !e.Gone() {
			kept = append(kept, e)
		}
	}
	clear(s.enemies[len(kept):])
	s.enemies = kept
}

// nap wakes the critter at the spot it entered the map at, with full petals,
// and rebuilds every enemy of the map at its start. Pointer walking waits for
// a release, as after a warp.
func (s *Scene) nap() error {
	enemies, err := s.spawnEnemies(s.m)
	if err != nil {
		return fmt.Errorf("map %q: %w", s.mapName, err)
	}
	s.player.Place(s.entrance)
	// A tapped villager is no longer walked to after a nap.
	s.tapTarget = nil
	s.enemies = enemies
	s.health.Wake()
	s.awaitRelease = true
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

// Lights returns the light sources of the current map: on a dark map the
// lantern the critter carries, centred on its body at the camera's focus and
// unmoved by a hop; on any other map none.
func (s *Scene) Lights() []Light {
	if !s.m.Dark {
		return nil
	}
	return []Light{{Pos: s.player.Pos.Sub(geom.Vec{Y: focusLift}), Radius: lanternRadius}}
}

// DrawWorld draws the map layers and the critter, and on a dark map covers
// them with darkness except where the lights from Lights shine.
func (s *Scene) DrawWorld(c engine.Canvas) {
	c.Image.Fill(ink)
	s.pen.begin(c, s.camera)
	view := geom.Rect{Min: s.camera, Max: s.camera.Add(s.viewSize)}
	c0, r0, c1, r1 := s.m.VisibleCells(view)
	elapsed := time.Duration(s.elapsed * float64(time.Second))

	drawTileLayer(&s.pen, s.m, s.m.TileLayer("ground"), c0, r0, c1, r1, elapsed)
	drawTileLayer(&s.pen, s.m, s.m.TileLayer("decor"), c0, r0, c1, r1, elapsed)
	drawProps(&s.pen, s.m, s.m.ObjectLayer("props"), view, elapsed)
	s.drawCritters(view)
	s.drawCharacters()
	drawTileLayer(&s.pen, s.m, s.m.TileLayer("above"), c0, r0, c1, r1, elapsed)
	if s.m.Dark {
		s.mask.draw(c, s.camera, s.viewSize, s.Lights())
	}
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

// drawCritters draws the player and the wild critters in view, the ones
// whose feet are higher up the map first, so a critter in front of another
// is drawn over it. The player is drawn over a wild critter with its feet at
// the same height.
func (s *Scene) drawCritters(view geom.Rect) {
	near := view.Inset(-wildReach)
	s.inView = s.inView[:0]
	for _, w := range s.wild {
		if near.Contains(w.Pos) {
			s.inView = append(s.inView, w)
		}
	}
	slices.SortStableFunc(s.inView, func(a, b *wildCritter) int {
		return cmp.Compare(a.Pos.Y, b.Pos.Y)
	})

	feet := s.player.Pos
	playerDrawn := false
	for _, w := range s.inView {
		if !playerDrawn && w.Pos.Y > feet.Y {
			drawCritter(&s.pen, feet.X, feet.Y, s.player.HopHeight()+s.bob, s.player.Facing)
			playerDrawn = true
		}
		s.drawWild(w)
	}
	if !playerDrawn {
		drawCritter(&s.pen, feet.X, feet.Y, s.player.HopHeight()+s.bob, s.player.Facing)
	}
}

// drawWild draws one wild critter, with twinkling stars around it if it is
// rare and a "!" above it while it is noticing the player.
func (s *Scene) drawWild(w *wildCritter) {
	lift := w.HopHeight()
	drawWildCritter(&s.pen, w.Pos.X, w.Pos.Y, lift, w.Facing, speciesColor(w.species.ID))
	if w.Rare {
		// The spawn id shifts each critter's twinkle so rare critters side
		// by side do not twinkle in step.
		drawSparkle(&s.pen, w.Pos.X, w.Pos.Y-wildRadius-lift, s.elapsed, float64(w.id))
	}
	if w.Noticing() {
		drawNotice(&s.pen, w.Pos.X, w.Pos.Y-noticeTop)
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
// critter's petals, the number of villagers, the villager last interacted
// with, the walk target and the number of wild critters to the debug overlay.
func (s *Scene) DebugLines() []string {
	p := s.player
	lines := []string{
		"map " + s.mapName,
		fmt.Sprintf("critter %.1f,%.1f facing %s", p.Pos.X, p.Pos.Y, p.Facing),
		fmt.Sprintf("camera %.0f,%.0f", s.camera.X, s.camera.Y),
		fmt.Sprintf("petals %d/%d", s.health.Petals(), s.health.Max()),
		fmt.Sprintf("villagers %d", len(s.npcs)),
	}
	if s.lastNPC != "" {
		lines = append(lines, "talked to "+s.lastNPC)
	}
	if t, ok := p.Target(); ok {
		lines = append(lines, fmt.Sprintf("walk to %.0f,%.0f", t.X, t.Y))
	}
	lines = append(lines, fmt.Sprintf("wild critters %d", len(s.wild)))
	return lines
}

// DrawDebug outlines where the critter may walk and the map's spawn zones,
// and marks the critter's feet and its walk target. Sizes are in world
// pixels, so the marks grow with the view.
func (s *Scene) DrawDebug(c engine.Canvas) {
	k := float32(c.Scale)
	s.strokeWorldRect(c, s.walkable, debugColor)
	for _, z := range s.m.SpawnZones() {
		s.strokeWorldRect(c, z.Rect, zoneDebug)
	}

	fx, fy := c.At(s.feetInView())
	vector.FillRect(c.Image, fx-k, fy-k, 2*k, 2*k, debugColor, true)

	if t, ok := s.player.Target(); ok {
		x, y := c.At(t.Sub(s.camera))
		vector.StrokeLine(c.Image, x-3*k, y-3*k, x+3*k, y+3*k, k, debugColor, true)
		vector.StrokeLine(c.Image, x-3*k, y+3*k, x+3*k, y-3*k, k, debugColor, true)
	}
}

// strokeWorldRect outlines r, given in world pixels, with a line one world
// pixel wide.
func (s *Scene) strokeWorldRect(c engine.Canvas, r geom.Rect, clr color.Color) {
	k := float32(c.Scale)
	left, top := c.At(r.Min.Sub(s.camera))
	size := r.Max.Sub(r.Min)
	vector.StrokeRect(c.Image, left, top, float32(size.X)*k, float32(size.Y)*k, k, clr, true)
}
