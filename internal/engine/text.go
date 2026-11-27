package engine

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
)

// NewFace returns a face of the variable font src at size pixels, with a
// weight from 300 (light) to 700 (bold). Change Size when the device scale
// factor changes; the glyph cache is kept per size.
func NewFace(src *text.GoTextFaceSource, size float64, weight float32) *text.GoTextFace {
	f := &text.GoTextFace{Source: src, Size: size}
	f.SetVariation(text.MustParseTag("wght"), weight)
	return f
}

// LineHeight returns the distance from one line of text in face to the next.
func LineHeight(face text.Face) float64 {
	m := face.Metrics()
	return m.HAscent + m.HDescent + m.HLineGap
}

// DrawText draws s in face with the top of its first line at y. align says
// whether x is the start, the centre or the end of each line.
func DrawText(dst *ebiten.Image, s string, face text.Face, x, y float64, align text.Align, clr color.Color) {
	op := &text.DrawOptions{}
	op.GeoM.Translate(x, y)
	op.ColorScale.ScaleWithColor(clr)
	op.LineSpacing = LineHeight(face)
	op.PrimaryAlign = align
	text.Draw(dst, s, face, op)
}
