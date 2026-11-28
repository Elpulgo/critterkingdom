package engine_test

import (
	"slices"
	"strings"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/Elpulgo/critterkingdom/internal/engine"
	"github.com/Elpulgo/critterkingdom/internal/geom"
	"github.com/Elpulgo/critterkingdom/internal/world"
)

// The room is a 6x4-tile map, 96x64 world pixels, with the critter starting
// at (24,40) facing down and the baker standing 20 px in front of it at
// (24,60), so a paw press talks to the baker.
const room = `{
	"orientation": "orthogonal", "infinite": false,
	"width": 6, "height": 4, "tilewidth": 128, "tileheight": 128,
	"properties": [{"name": "biome", "type": "string", "value": "village"}],
	"layers": [
		{"name": "collision", "type": "objectgroup", "objects": []},
		{"name": "objects", "type": "objectgroup", "objects": [
			{"id": 1, "name": "start", "type": "player_spawn", "x": 192, "y": 320, "point": true},
			{"id": 2, "type": "npc", "x": 192, "y": 480, "point": true,
				"properties": [{"name": "id", "type": "string", "value": "baker_badger"}]}
		]}
	]
}`

// roomView fits the whole room, so the camera stays at 0,0 and a pointer's
// view position is its world position.
var roomView = engine.View{Scale: 1, W: 96, H: 64}

func newRoomScene(t *testing.T) *world.Scene {
	t.Helper()
	s, err := world.NewScene(nil, func(string) (*world.Map, error) { return world.LoadMap([]byte(room)) }, "room")
	if err != nil {
		t.Fatalf("NewScene: %v", err)
	}
	return s
}

// tick is one tick at 60 ticks per second with the input in.
func tick(in engine.Input) engine.Frame {
	return engine.Frame{Input: in, View: roomView, DT: 1.0 / 60}
}

// critterState returns the lines of the scene's debug overlay that say where
// the critter stands, which way it faces, and where it walks to, if anywhere.
func critterState(s *world.Scene) []string {
	var got []string
	for _, l := range s.DebugLines() {
		if strings.HasPrefix(l, "critter ") || strings.HasPrefix(l, "walk to ") {
			got = append(got, l)
		}
	}
	return got
}

// tapToClose stands in for an overlay such as the dialogue box, which is
// built in another package on top of the stack: it closes on the first tap.
type tapToClose struct{ done bool }

func (o *tapToClose) Update(f engine.Frame) error {
	o.done = o.done || f.Input.Tap
	return nil
}
func (o *tapToClose) DrawWorld(engine.Canvas)         {}
func (o *tapToClose) DrawUI(*ebiten.Image, engine.UI) {}
func (o *tapToClose) Done() bool                      { return o.done }

func update(t *testing.T, s *engine.Stack, f engine.Frame) {
	t.Helper()
	if err := s.Update(f); err != nil {
		t.Fatalf("Update: %v", err)
	}
}

func TestStack_AnOverlayPausesTheWorldSoWalkingAndPawDoNothing(t *testing.T) {
	scene := newRoomScene(t)
	stack := engine.NewStack(scene)
	stack.Push(&tapToClose{})
	before := critterState(scene)

	walk := engine.Input{Move: geom.Vec{X: 1}, Pointing: true, Point: geom.Vec{X: 80, Y: 40}}
	paw := walk
	paw.Paw = true
	update(t, stack, tick(paw))
	for range 30 {
		update(t, stack, tick(walk))
	}

	if got := critterState(scene); !slices.Equal(got, before) {
		t.Errorf("critter after walking under the overlay = %q; want it to stand as before, %q", got, before)
	}
	if in, ok := scene.TakeInteraction(); ok {
		t.Errorf("paw under the overlay started an interaction with %q; want none", in.NPC)
	}
	if !stack.Overlaid() {
		t.Errorf("the overlay is gone without having finished")
	}
}

func TestStack_APressThatClosesAnOverlayDoesNotWalkTheCritterUntilPressedAgain(t *testing.T) {
	scene := newRoomScene(t)
	stack := engine.NewStack(scene)
	stack.Push(&tapToClose{})
	before := critterState(scene)
	there := geom.Vec{X: 80, Y: 40}

	// The press lands, closes the overlay and is held.
	update(t, stack, tick(engine.Input{Tap: true, Pointing: true, Point: there}))
	if stack.Overlaid() {
		t.Fatalf("the overlay is still up after the tap that closed it")
	}
	for range 30 {
		update(t, stack, tick(engine.Input{Pointing: true, Point: there}))
	}
	if got := critterState(scene); !slices.Equal(got, before) {
		t.Fatalf("critter while the closing press is held = %q; want it to stand as before, %q", got, before)
	}

	// Let go and press again: now the critter walks there.
	update(t, stack, tick(engine.Input{}))
	update(t, stack, tick(engine.Input{Tap: true, Pointing: true, Point: there}))
	if got := critterState(scene); slices.Equal(got, before) {
		t.Errorf("critter after a new press = %q; want it walking towards %v", got, there)
	}
}
