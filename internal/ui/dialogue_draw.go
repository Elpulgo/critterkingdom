package ui

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"github.com/Elpulgo/critterkingdom/internal/engine"
	"github.com/Elpulgo/critterkingdom/internal/geom"
)

// Placeholder interface art drawn in code, in the Art Bible's colours from
// the Endesga 32 palette, until the panel, portrait and icon art is
// exported.
var (
	panelCream    = color.NRGBA{0xea, 0xd4, 0xaa, 0xff}
	warmTaupe     = color.NRGBA{0xc2, 0x85, 0x69, 0xff}
	portraitPeach = color.NRGBA{0xe8, 0xb7, 0x96, 0xff}
	cocoa         = color.NRGBA{0x3e, 0x27, 0x31, 0xff}
	dropShadow    = color.NRGBA{0x18, 0x14, 0x25, 0x3c}
	white         = color.NRGBA{0xff, 0xff, 0xff, 0xff}
	ink           = color.NRGBA{0x18, 0x14, 0x25, 0xff}
	headGrey      = color.NRGBA{0x8b, 0x9b, 0xb4, 0xff}
	headOutline   = color.NRGBA{0x5a, 0x69, 0x88, 0xff}
	cheekPink     = color.NRGBA{0xf6, 0x75, 0x7a, 0x90}
	sunnyYellow   = color.NRGBA{0xfe, 0xe7, 0x61, 0xff}
	leafGreen     = color.NRGBA{0x3e, 0x89, 0x48, 0xff}
	poppyRed      = color.NRGBA{0xe4, 0x3b, 0x44, 0xff}
)

// Sizes of the drawing, in CSS pixels; they are multiplied by the unit.
const (
	// nameSize and nameWeight are the villager's name: a little smaller
	// than the dialogue text, so it fits its row, and bolder.
	nameSize   = 24
	nameWeight = 600
	// choiceSize and choiceWeight are a choice button's text. Text too
	// wide for its button is drawn smaller to fit.
	choiceSize   = 24
	choiceWeight = 600

	panelRadius   = 16
	panelOutline  = 3
	cardRadius    = 12
	cardOutline   = 2
	buttonRadius  = 14
	buttonOutline = 3
	// highlightOutline is the outline of the highlighted choice button.
	highlightOutline = 4
	// shadowDrop is how far the soft shadow under the panel and the buttons
	// is offset downwards.
	shadowDrop = 4

	// iconSize is the round choice icon, buttonInset the gap between a
	// button's edge and its icon or text.
	iconSize    = 40
	buttonInset = 14

	// arrowWidth and arrowHeight are the "next" arrow, drawn in the panel's
	// bottom padding under the end of the text.
	arrowWidth  = 14
	arrowHeight = 9
)

// DrawUI draws the dialogue box at the places LayoutDialogue gives: the
// panel, the portrait card with the villager's face, the name, the part of
// the page typed so far and, on a question, the choice buttons. A small
// arrow shows that a tap goes on once the page is fully shown.
func (d *Dialogue) DrawUI(screen *ebiten.Image, ui engine.UI) {
	if d.runner.Done() {
		return
	}
	u := ui.Unit
	if u <= 0 {
		u = 1
	}
	b := screen.Bounds()
	choices := d.Choices()
	l := LayoutDialogue(float64(b.Dx()), float64(b.Dy()), u, len(choices))

	fillRoundRect(screen, offset(l.Box, shadowDrop*u), panelRadius*u, dropShadow)
	fillRoundRect(screen, l.Box, panelRadius*u, panelCream)
	strokeRoundRect(screen, l.Box, panelRadius*u, panelOutline*u, warmTaupe)

	fillRoundRect(screen, l.Portrait, cardRadius*u, portraitPeach)
	strokeRoundRect(screen, l.Portrait, cardRadius*u, cardOutline*u, warmTaupe)
	drawHead(screen, l.Portrait, d.Face())

	d.nameFace.Size = nameSize * u
	nameTop := l.Name.Min.Y + (l.Name.Max.Y-l.Name.Min.Y-engine.LineHeight(d.nameFace))/2
	engine.DrawText(screen, d.Name(), d.nameFace, l.Name.Min.X, nameTop, text.AlignStart, cocoa)

	// The face is sized to the unit by Update, which wraps the text with it.
	engine.DrawText(screen, d.Text(), d.face, l.Text.Min.X, l.Text.Min.Y, text.AlignStart, cocoa)

	if len(choices) == 0 {
		if d.typer != nil && d.typer.Done() {
			drawNextArrow(screen, l, u)
		}
		return
	}
	hi := d.Highlighted()
	for i, c := range choices {
		d.drawChoice(screen, l.Choices[i], c, i == hi, u)
	}
}

// drawChoice draws one choice button in r: its icon on the left, if it has
// one, and its text centred in the rest. The highlighted button is yellow
// with a thick cocoa outline.
func (d *Dialogue) drawChoice(screen *ebiten.Image, r geom.Rect, c Choice, highlighted bool, u float64) {
	fill, edge, width := panelCream, warmTaupe, buttonOutline*u
	if highlighted {
		fill, edge, width = sunnyYellow, cocoa, highlightOutline*u
	}
	fillRoundRect(screen, offset(r, shadowDrop*u), buttonRadius*u, dropShadow)
	fillRoundRect(screen, r, buttonRadius*u, fill)
	strokeRoundRect(screen, r, buttonRadius*u, width, edge)

	left := r.Min.X + buttonInset*u
	cy := (r.Min.Y + r.Max.Y) / 2
	if drawIcon(screen, c.Icon, left+iconSize*u/2, cy, iconSize*u/2, u) {
		left += iconSize*u + buttonInset*u/2
	}
	right := r.Max.X - buttonInset*u

	d.choiceFace.Size = choiceSize * u
	if w := text.Advance(c.Text, d.choiceFace); w > right-left {
		d.choiceFace.Size *= (right - left) / w
	}
	top := cy - engine.LineHeight(d.choiceFace)/2
	engine.DrawText(screen, c.Text, d.choiceFace, (left+right)/2, top, text.AlignCenter, cocoa)
}

// drawIcon draws the choice icon called name as a disc of radius r centred
// on cx, cy, and reports whether it knows the name: "check" is a white tick
// on green and "cross" a white cross on red.
func drawIcon(screen *ebiten.Image, name string, cx, cy, r, u float64) bool {
	var p vector.Path
	switch name {
	case "check":
		vector.FillCircle(screen, float32(cx), float32(cy), float32(r), leafGreen, true)
		p.MoveTo(float32(cx-0.45*r), float32(cy+0.02*r))
		p.LineTo(float32(cx-0.12*r), float32(cy+0.35*r))
		p.LineTo(float32(cx+0.48*r), float32(cy-0.3*r))
	case "cross":
		vector.FillCircle(screen, float32(cx), float32(cy), float32(r), poppyRed, true)
		k := 0.36 * r
		p.MoveTo(float32(cx-k), float32(cy-k))
		p.LineTo(float32(cx+k), float32(cy+k))
		p.MoveTo(float32(cx+k), float32(cy-k))
		p.LineTo(float32(cx-k), float32(cy+k))
	default:
		return false
	}
	strokePath(screen, &p, 5*u, white)
	return true
}

// drawNextArrow draws a small cocoa arrow pointing down in the panel's
// bottom padding, under the right end of the text area.
func drawNextArrow(screen *ebiten.Image, l DialogueLayout, u float64) {
	cx := l.Text.Max.X - arrowWidth*u/2
	top := l.Box.Max.Y - (boxPadding-1)*u
	var p vector.Path
	p.MoveTo(float32(cx-arrowWidth*u/2), float32(top))
	p.LineTo(float32(cx+arrowWidth*u/2), float32(top))
	p.LineTo(float32(cx), float32(top+arrowHeight*u))
	p.Close()
	fillPath(screen, &p, cocoa)
}

// drawHead draws a placeholder villager head centred on the portrait card
// card: a round grey head with two round ears, a cream muzzle, eyes and a
// mouth that follows face. "happy" smiles, "sad" frowns under drooping
// brows, and anything else, "neutral" included, is a short flat mouth.
func drawHead(screen *ebiten.Image, card geom.Rect, face string) {
	side := card.Max.X - card.Min.X
	cx := (card.Min.X + card.Max.X) / 2
	cy := card.Min.Y + side*0.56
	r := side * 0.3
	line := side * 0.035

	for _, s := range []float64{-1, 1} {
		ex, ey := cx+s*r*0.72, cy-r*0.82
		fillEllipse(screen, ex, ey, r*0.36+line, r*0.36+line, headOutline)
		fillEllipse(screen, ex, ey, r*0.36, r*0.36, headGrey)
		fillEllipse(screen, ex, ey, r*0.18, r*0.18, portraitPeach)
	}
	fillEllipse(screen, cx, cy, r+line, r+line, headOutline)
	fillEllipse(screen, cx, cy, r, r, headGrey)
	fillEllipse(screen, cx, cy+r*0.32, r*0.62, r*0.5, panelCream)

	eyeY := cy - r*0.18
	for _, s := range []float64{-1, 1} {
		ex := cx + s*r*0.38
		fillEllipse(screen, ex, eyeY, r*0.11, r*0.16, ink)
		fillEllipse(screen, ex-r*0.03, eyeY-r*0.06, r*0.04, r*0.04, white)
		if face == "happy" {
			fillEllipse(screen, cx+s*r*0.62, cy+r*0.2, r*0.13, r*0.08, cheekPink)
		}
		if face == "sad" {
			// The inner end of each brow sits higher: worried, not cross.
			var p vector.Path
			p.MoveTo(float32(cx+s*r*0.55), float32(eyeY-r*0.22))
			p.LineTo(float32(cx+s*r*0.22), float32(eyeY-r*0.36))
			strokePath(screen, &p, line, cocoa)
		}
	}
	fillEllipse(screen, cx, cy+r*0.12, r*0.12, r*0.08, cocoa)

	mouthY := cy + r*0.42
	var p vector.Path
	switch face {
	case "happy":
		p.Arc(float32(cx), float32(mouthY-r*0.14), float32(r*0.24), 0.15*math.Pi, 0.85*math.Pi, vector.Clockwise)
	case "sad":
		p.Arc(float32(cx), float32(mouthY+r*0.2), float32(r*0.2), 1.2*math.Pi, 1.8*math.Pi, vector.Clockwise)
	default:
		p.MoveTo(float32(cx-r*0.16), float32(mouthY))
		p.LineTo(float32(cx+r*0.16), float32(mouthY))
	}
	strokePath(screen, &p, line, cocoa)
}

// offset returns r moved down by dy.
func offset(r geom.Rect, dy float64) geom.Rect {
	d := geom.Vec{Y: dy}
	return geom.Rect{Min: r.Min.Add(d), Max: r.Max.Add(d)}
}

// roundRect returns the outline of r with corners rounded to radius.
func roundRect(r geom.Rect, radius float64) *vector.Path {
	left, top := float32(r.Min.X), float32(r.Min.Y)
	right, bottom := float32(r.Max.X), float32(r.Max.Y)
	rad := float32(math.Min(radius, math.Min(r.Max.X-r.Min.X, r.Max.Y-r.Min.Y)/2))
	var p vector.Path
	p.MoveTo(left+rad, top)
	p.LineTo(right-rad, top)
	p.ArcTo(right, top, right, top+rad, rad)
	p.LineTo(right, bottom-rad)
	p.ArcTo(right, bottom, right-rad, bottom, rad)
	p.LineTo(left+rad, bottom)
	p.ArcTo(left, bottom, left, bottom-rad, rad)
	p.LineTo(left, top+rad)
	p.ArcTo(left, top, left+rad, top, rad)
	p.Close()
	return &p
}

func fillRoundRect(dst *ebiten.Image, r geom.Rect, radius float64, clr color.Color) {
	fillPath(dst, roundRect(r, radius), clr)
}

// strokeRoundRect outlines r, rounded to radius, with a line width wide
// that lies inside r, so the outline does not grow the shape.
func strokeRoundRect(dst *ebiten.Image, r geom.Rect, radius, width float64, clr color.Color) {
	strokePath(dst, roundRect(r.Inset(width/2), radius-width/2), width, clr)
}

// fillEllipse fills an ellipse centred on cx, cy with radii rx and ry.
func fillEllipse(dst *ebiten.Image, cx, cy, rx, ry float64, clr color.Color) {
	// k puts the control points of a cubic Bézier that follows a quarter
	// of a circle.
	const k = 0.5522847
	x, y, a, b := float32(cx), float32(cy), float32(rx), float32(ry)
	var p vector.Path
	p.MoveTo(x+a, y)
	p.CubicTo(x+a, y+k*b, x+k*a, y+b, x, y+b)
	p.CubicTo(x-k*a, y+b, x-a, y+k*b, x-a, y)
	p.CubicTo(x-a, y-k*b, x-k*a, y-b, x, y-b)
	p.CubicTo(x+k*a, y-b, x+a, y-k*b, x+a, y)
	p.Close()
	fillPath(dst, &p, clr)
}

func fillPath(dst *ebiten.Image, p *vector.Path, clr color.Color) {
	op := &vector.DrawPathOptions{AntiAlias: true}
	op.ColorScale.ScaleWithColor(clr)
	vector.FillPath(dst, p, nil, op)
}

// strokePath draws p with a line width wide, with round ends and corners.
func strokePath(dst *ebiten.Image, p *vector.Path, width float64, clr color.Color) {
	so := &vector.StrokeOptions{
		Width:    float32(width),
		LineCap:  vector.LineCapRound,
		LineJoin: vector.LineJoinRound,
	}
	op := &vector.DrawPathOptions{AntiAlias: true}
	op.ColorScale.ScaleWithColor(clr)
	vector.StrokePath(dst, p, so, op)
}
