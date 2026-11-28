package ui

import (
	"math"

	"github.com/Elpulgo/critterkingdom/internal/geom"
)

// Dialogue box sizes, in CSS pixels. They are multiplied by the unit (screen
// pixels per CSS pixel) so the box has the same physical size on every
// screen.
const (
	// boxMargin is the gap between the box and the screen edges, and between
	// the box and the choice buttons above it.
	boxMargin = 12
	// boxMaxWidth keeps the box readable on wide screens; it is centred.
	boxMaxWidth = 960
	// boxPadding is the gap inside the box around the portrait and text.
	boxPadding = 16
	// nameHeight is the row the villager's name sits in.
	nameHeight = 32
	// textHeight holds LinesPerPage lines of TextSize text.
	textHeight = 66
	// boxHeight is the whole box, portrait and two lines of text included.
	boxHeight = boxPadding + nameHeight + textHeight + boxPadding
	// choiceWidth and choiceHeight are a choice button's size; both are at
	// least 64 CSS pixels, the smallest touch target small fingers hit
	// reliably.
	choiceWidth  = 180
	choiceHeight = 72
	// choiceGap is the space between the two choice buttons.
	choiceGap = 16
)

// DialogueLayout is where the parts of the dialogue box go, in screen
// pixels.
type DialogueLayout struct {
	// Box is the panel across the bottom of the screen.
	Box geom.Rect
	// Portrait is the square card on the left of the box.
	Portrait geom.Rect
	// Name is the row for the villager's name, right of the portrait.
	Name geom.Rect
	// Text is the area for the page's lines, under the name.
	Text geom.Rect
	// Choices are the choice buttons, first to last, side by side above the
	// right end of the box.
	Choices []geom.Rect
}

// LayoutDialogue places the dialogue box on a screen screenW by screenH
// screen pixels, with unit screen pixels per CSS pixel, and with room for
// choices choice buttons (0 to 2).
func LayoutDialogue(screenW, screenH, unit float64, choices int) DialogueLayout {
	u := unit
	if u <= 0 {
		u = 1
	}
	w := math.Min(screenW-2*boxMargin*u, boxMaxWidth*u)
	box := geom.Rect{
		Min: geom.Vec{X: (screenW - w) / 2, Y: screenH - (boxMargin+boxHeight)*u},
	}
	box.Max = geom.Vec{X: box.Min.X + w, Y: box.Min.Y + boxHeight*u}

	inner := box.Inset(boxPadding * u)
	side := inner.Max.Y - inner.Min.Y
	portrait := geom.Rect{Min: inner.Min, Max: geom.Vec{X: inner.Min.X + side, Y: inner.Max.Y}}
	textLeft := portrait.Max.X + boxPadding*u
	name := geom.Rect{
		Min: geom.Vec{X: textLeft, Y: inner.Min.Y},
		Max: geom.Vec{X: inner.Max.X, Y: inner.Min.Y + nameHeight*u},
	}
	txt := geom.Rect{
		Min: geom.Vec{X: textLeft, Y: name.Max.Y},
		Max: inner.Max,
	}

	l := DialogueLayout{Box: box, Portrait: portrait, Name: name, Text: txt}
	bottom := box.Min.Y - boxMargin*u
	right := box.Max.X
	for i := range choices {
		// Buttons are laid out right to left from the box's right edge, so
		// the first choice ends up on the left.
		n := choices - 1 - i
		x := right - float64(n+1)*choiceWidth*u - float64(n)*choiceGap*u
		l.Choices = append(l.Choices, geom.Rect{
			Min: geom.Vec{X: x, Y: bottom - choiceHeight*u},
			Max: geom.Vec{X: x + choiceWidth*u, Y: bottom},
		})
	}
	return l
}
