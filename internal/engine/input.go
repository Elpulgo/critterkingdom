package engine

import (
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"

	"github.com/Elpulgo/critterkingdom/internal/geom"
)

// Input is what the player asked for during one tick, merged from every
// device, so scenes never need to know whether a child uses a tablet, a
// keyboard, a gamepad or a mouse.
type Input struct {
	// Move is a walk direction from the keys, a stick, the d-pad or the touch
	// pad. It is zero or one of eight directions of length 1.
	Move geom.Vec
	// Pointing is true while a finger or the left mouse button is held down
	// on the world, and Point is where, in view pixels.
	Pointing bool
	Point    geom.Vec
	// Paw is true on the tick the paw button was pressed.
	Paw bool
	// Menu is true on the tick the menu button was pressed.
	Menu bool
	// Tap is true on the tick a finger lands anywhere on the screen (the
	// world, the touch pad half or the paw button) or the left mouse button
	// goes down, and TapAt is where, in screen pixels. It ignores what the
	// finger does after that, so interface such as a dialogue box can treat
	// any touch as a tap. While the scene has an overlay up, tapping is all
	// a new finger does: it neither presses paw nor steers nor points.
	Tap   bool
	TapAt geom.Vec
}

// stickDeadZone is how far a stick or the touch pad must be pushed, from 0 to
// 1, before the critter walks.
const stickDeadZone = 0.25

// diagonal is the length of each part of a diagonal direction of length 1.
const diagonal = math.Sqrt2 / 2

// eightWay holds the eight walk directions, clockwise from the right as seen
// on screen, where Y grows downwards.
var eightWay = [8]geom.Vec{
	{X: 1, Y: 0},
	{X: diagonal, Y: diagonal},
	{X: 0, Y: 1},
	{X: -diagonal, Y: diagonal},
	{X: -1, Y: 0},
	{X: -diagonal, Y: -diagonal},
	{X: 0, Y: -1},
	{X: diagonal, Y: -diagonal},
}

// snapDirection turns a stick position, with each part from -1 to 1, into the
// nearest of the eight walk directions. It returns zero inside the dead zone.
func snapDirection(v geom.Vec) geom.Vec {
	if v.Len() < stickDeadZone {
		return geom.Vec{}
	}
	sector := int(math.Round(math.Atan2(v.Y, v.X)/(math.Pi/4))) & 7
	return eightWay[sector]
}

// keyDirection turns four held direction buttons into a walk direction.
// Opposite buttons cancel each other out.
func keyDirection(up, down, left, right bool) geom.Vec {
	var v geom.Vec
	if up {
		v.Y--
	}
	if down {
		v.Y++
	}
	if left {
		v.X--
	}
	if right {
		v.X++
	}
	return v.Normalized()
}

// inputReader reads every input device once per tick.
type inputReader struct {
	touch    touchControls
	gamepads []ebiten.GamepadID
}

// read returns this tick's input. The keyboard wins over a gamepad, which
// wins over the touch pad. A finger on the world wins over the mouse.
func (r *inputReader) read(view View, l touchLayout) Input {
	touch := r.touch.update(l)
	keyMove, keyPaw, keyMenu := readKeys()
	padMove, padPaw, padMenu := r.readGamepads()

	in := Input{
		Move: keyMove,
		Paw:  keyPaw || padPaw || touch.paw,
		Menu: keyMenu || padMenu,
	}
	if in.Move.IsZero() {
		in.Move = padMove
	}
	if in.Move.IsZero() {
		in.Move = touch.move
	}

	switch {
	case touch.pointing:
		in.Pointing = true
		in.Point = view.ScreenToView(touch.point.X, touch.point.Y)
	case ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft):
		in.Pointing = true
		in.Point = view.ScreenToView(ebiten.CursorPositionF())
	}

	switch {
	case touch.tap:
		in.Tap = true
		in.TapAt = touch.tapAt
	case inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft):
		x, y := ebiten.CursorPositionF()
		in.Tap = true
		in.TapAt = geom.Vec{X: x, Y: y}
	}
	return in
}

func readKeys() (move geom.Vec, paw, menu bool) {
	move = keyDirection(
		anyKeyPressed(ebiten.KeyArrowUp, ebiten.KeyW),
		anyKeyPressed(ebiten.KeyArrowDown, ebiten.KeyS),
		anyKeyPressed(ebiten.KeyArrowLeft, ebiten.KeyA),
		anyKeyPressed(ebiten.KeyArrowRight, ebiten.KeyD),
	)
	paw = anyKeyJustPressed(ebiten.KeySpace, ebiten.KeyEnter, ebiten.KeyE)
	menu = anyKeyJustPressed(ebiten.KeyEscape, ebiten.KeyTab)
	return move, paw, menu
}

// readGamepads reads every gamepad with the standard layout. The d-pad wins
// over the left stick.
func (r *inputReader) readGamepads() (move geom.Vec, paw, menu bool) {
	r.gamepads = ebiten.AppendGamepadIDs(r.gamepads[:0])
	for _, id := range r.gamepads {
		if !ebiten.IsStandardGamepadLayoutAvailable(id) {
			continue
		}
		if move.IsZero() {
			move = keyDirection(
				ebiten.IsStandardGamepadButtonPressed(id, ebiten.StandardGamepadButtonLeftTop),
				ebiten.IsStandardGamepadButtonPressed(id, ebiten.StandardGamepadButtonLeftBottom),
				ebiten.IsStandardGamepadButtonPressed(id, ebiten.StandardGamepadButtonLeftLeft),
				ebiten.IsStandardGamepadButtonPressed(id, ebiten.StandardGamepadButtonLeftRight),
			)
		}
		if move.IsZero() {
			move = snapDirection(geom.Vec{
				X: ebiten.StandardGamepadAxisValue(id, ebiten.StandardGamepadAxisLeftStickHorizontal),
				Y: ebiten.StandardGamepadAxisValue(id, ebiten.StandardGamepadAxisLeftStickVertical),
			})
		}
		paw = paw || inpututil.IsStandardGamepadButtonJustPressed(id, ebiten.StandardGamepadButtonRightBottom)
		menu = menu ||
			inpututil.IsStandardGamepadButtonJustPressed(id, ebiten.StandardGamepadButtonCenterRight) ||
			inpututil.IsStandardGamepadButtonJustPressed(id, ebiten.StandardGamepadButtonRightTop)
	}
	return move, paw, menu
}

func anyKeyPressed(keys ...ebiten.Key) bool {
	for _, k := range keys {
		if ebiten.IsKeyPressed(k) {
			return true
		}
	}
	return false
}

func anyKeyJustPressed(keys ...ebiten.Key) bool {
	for _, k := range keys {
		if inpututil.IsKeyJustPressed(k) {
			return true
		}
	}
	return false
}
