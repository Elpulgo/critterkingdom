package engine

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"github.com/Elpulgo/critterkingdom/internal/geom"
)

// Touch control sizes, in CSS pixels. They are multiplied by the device scale
// factor so they have the same physical size on every screen; 64 CSS pixels
// is about a centimetre on a tablet, which small fingers hit reliably.
const (
	pawButtonRadius = 48
	pawButtonMargin = 28
	padRadius       = 64
	padKnobRadius   = 28
)

var (
	buttonColor        = color.NRGBA{0xff, 0xff, 0xff, 0x70}
	buttonPressedColor = color.NRGBA{0xff, 0xff, 0xff, 0xb0}
	pawPrintColor      = color.NRGBA{0x3e, 0x27, 0x31, 0xc0}
	padRingColor       = color.NRGBA{0xff, 0xff, 0xff, 0x90}
	padKnobColor       = color.NRGBA{0xff, 0xff, 0xff, 0xb0}
	padHintColor       = color.NRGBA{0xff, 0xff, 0xff, 0x40}
)

// touchLayout places the touch controls on a screen of the given size, in
// screen pixels. unit is the number of screen pixels per CSS pixel.
type touchLayout struct {
	screenW, screenH float64
	unit             float64
}

func (l touchLayout) pawCenter() geom.Vec {
	d := (pawButtonRadius + pawButtonMargin) * l.unit
	return geom.Vec{X: l.screenW - d, Y: l.screenH - d}
}

// onPawButton reports whether a finger at p presses the paw button. The hit
// area reaches a little past the drawn button, because small fingers miss.
func (l touchLayout) onPawButton(p geom.Vec) bool {
	return p.Sub(l.pawCenter()).Len() <= (pawButtonRadius+pawButtonMargin/2)*l.unit
}

// inPadZone reports whether a finger at p starts the floating touch pad.
func (l touchLayout) inPadZone(p geom.Vec) bool {
	return p.X < l.screenW/2
}

// padHintCenter is where the idle touch pad hint is drawn.
func (l touchLayout) padHintCenter() geom.Vec {
	d := (padRadius + pawButtonMargin) * l.unit
	return geom.Vec{X: d, Y: l.screenH - d}
}

// fingerRole is what a finger does. It is decided when the finger touches the
// screen and kept until it lifts, so sliding off a button never surprises.
type fingerRole int

const (
	walkFinger fingerRole = iota // the critter walks towards it
	padFinger                    // steers with the floating touch pad
	pawFinger                    // pressed the paw button
	idleFinger                   // does nothing
)

type finger struct {
	role  fingerRole
	start geom.Vec // where it touched down, in screen pixels
	pos   geom.Vec // where it is now, in screen pixels
	seq   int      // when it touched down; higher is newer
}

type touchPoint struct {
	id  ebiten.TouchID
	pos geom.Vec
}

// touchInput is what the fingers on the screen asked for during one tick.
type touchInput struct {
	move     geom.Vec
	pointing bool
	point    geom.Vec // screen pixels
	paw      bool
	// tap is true on the tick a finger lands anywhere on the screen, whatever
	// role it gets, and tapAt is where, in screen pixels. When several fingers
	// land on the same tick, tapAt is the first of them.
	tap   bool
	tapAt geom.Vec
}

// touchControls turns fingers into input: a finger on the world walks the
// critter there, a finger on the paw button presses paw, and, when the touch
// pad is enabled, a finger on the left half of the screen steers with a pad
// that appears where it touched down.
type touchControls struct {
	padEnabled bool
	// used becomes true at the first touch. Only then are the buttons drawn
	// and pressable, so desktop players never see them.
	used    bool
	fingers map[ebiten.TouchID]*finger
	seq     int

	ids    []ebiten.TouchID
	points []touchPoint
}

// update reads the fingers on the screen and returns this tick's touch input.
func (t *touchControls) update(l touchLayout) touchInput {
	t.ids = ebiten.AppendTouchIDs(t.ids[:0])
	t.points = t.points[:0]
	for _, id := range t.ids {
		x, y := ebiten.TouchPositionF(id)
		t.points = append(t.points, touchPoint{id: id, pos: geom.Vec{X: x, Y: y}})
	}
	return t.step(t.points, l)
}

// step advances the touch controls by one tick, given every finger on the
// screen.
func (t *touchControls) step(points []touchPoint, l touchLayout) touchInput {
	if t.fingers == nil {
		t.fingers = make(map[ebiten.TouchID]*finger)
	}
	for id := range t.fingers {
		if !hasTouch(points, id) {
			delete(t.fingers, id)
		}
	}

	var in touchInput
	var newest *finger
	for _, p := range points {
		f, ok := t.fingers[p.id]
		if !ok {
			f = t.touchDown(p.pos, l)
			t.fingers[p.id] = f
			in.paw = in.paw || f.role == pawFinger
			if !in.tap {
				in.tap = true
				in.tapAt = p.pos
			}
		}
		f.pos = p.pos
		switch f.role {
		case padFinger:
			in.move = snapDirection(f.pos.Sub(f.start).Scale(1 / (padRadius * l.unit)))
		case walkFinger:
			if newest == nil || f.seq > newest.seq {
				newest = f
			}
		}
	}
	if newest != nil {
		in.pointing = true
		in.point = newest.pos
	}
	t.used = t.used || len(points) > 0
	return in
}

// touchDown returns a new finger at pos with the role it gets from where it
// touched the screen.
func (t *touchControls) touchDown(pos geom.Vec, l touchLayout) *finger {
	t.seq++
	f := &finger{role: walkFinger, start: pos, pos: pos, seq: t.seq}
	switch {
	case t.used && l.onPawButton(pos):
		f.role = pawFinger
	case t.padEnabled && l.inPadZone(pos) && t.fingerWith(padFinger) != nil:
		f.role = idleFinger
	case t.padEnabled && l.inPadZone(pos):
		f.role = padFinger
	}
	return f
}

func (t *touchControls) fingerWith(role fingerRole) *finger {
	for _, f := range t.fingers {
		if f.role == role {
			return f
		}
	}
	return nil
}

func (t *touchControls) count() int {
	return len(t.fingers)
}

func hasTouch(points []touchPoint, id ebiten.TouchID) bool {
	for _, p := range points {
		if p.id == id {
			return true
		}
	}
	return false
}

// draw draws the paw button and the touch pad, once the screen has been
// touched.
func (t *touchControls) draw(dst *ebiten.Image, l touchLayout) {
	if !t.used {
		return
	}
	u := float32(l.unit)

	c := l.pawCenter()
	r := pawButtonRadius * u
	clr := buttonColor
	if t.fingerWith(pawFinger) != nil {
		clr = buttonPressedColor
		r *= 0.94
	}
	vector.FillCircle(dst, float32(c.X), float32(c.Y), r, clr, true)
	drawPawPrint(dst, float32(c.X), float32(c.Y), r)

	if f := t.fingerWith(padFinger); f != nil {
		knob := f.start.Add(f.pos.Sub(f.start).ClampLen(padRadius * l.unit))
		vector.StrokeCircle(dst, float32(f.start.X), float32(f.start.Y), padRadius*u, 3*u, padRingColor, true)
		vector.FillCircle(dst, float32(knob.X), float32(knob.Y), padKnobRadius*u, padKnobColor, true)
		return
	}
	if t.padEnabled {
		h := l.padHintCenter()
		vector.StrokeCircle(dst, float32(h.X), float32(h.Y), padRadius*u, 2*u, padHintColor, true)
		vector.FillCircle(dst, float32(h.X), float32(h.Y), padKnobRadius*u, padHintColor, true)
	}
}

// drawPawPrint draws a paw print that fits in a circle of radius r.
func drawPawPrint(dst *ebiten.Image, cx, cy, r float32) {
	vector.FillCircle(dst, cx, cy+0.18*r, 0.36*r, pawPrintColor, true)
	toes := [4][2]float32{{-0.46, -0.16}, {-0.17, -0.46}, {0.17, -0.46}, {0.46, -0.16}}
	for _, toe := range toes {
		vector.FillCircle(dst, cx+toe[0]*r, cy+toe[1]*r, 0.15*r, pawPrintColor, true)
	}
}
