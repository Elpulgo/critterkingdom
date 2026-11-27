package world

import (
	"image/color"
	"math"
	"math/rand/v2"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"github.com/Elpulgo/critterkingdom/internal/entity"
	"github.com/Elpulgo/critterkingdom/internal/geom"
)

// Placeholder art drawn in code with smooth shapes, in colours from the
// Endesga 32 palette. Sprites made from vector drawings replace it later.
var (
	grass       = color.NRGBA{0x63, 0xc7, 0x4d, 0xff}
	grassDark   = color.NRGBA{0x3e, 0x89, 0x48, 0xff}
	hedgeDark   = color.NRGBA{0x26, 0x5c, 0x42, 0xff}
	flowerHeart = color.NRGBA{0xfe, 0xae, 0x34, 0xff}
	petalWhite  = color.NRGBA{0xff, 0xff, 0xff, 0xff}
	petalPink   = color.NRGBA{0xf6, 0x75, 0x7a, 0xff}
	fur         = color.NRGBA{0xf7, 0x76, 0x22, 0xff}
	cream       = color.NRGBA{0xea, 0xd4, 0xaa, 0xff}
	ink         = color.NRGBA{0x18, 0x14, 0x25, 0xff}
	cocoa       = color.NRGBA{0x3e, 0x27, 0x31, 0xff}
	shadow      = color.NRGBA{0x00, 0x00, 0x00, 0x3c}
	white       = color.NRGBA{0xff, 0xff, 0xff, 0xff}
	debugColor  = color.NRGBA{0xfe, 0xe7, 0x61, 0xff}
)

const (
	// critterHeight is how far the tips of the critter's ears are above its
	// feet, in world pixels.
	critterHeight = 17
	// bodyRadius is the radius of the critter's round body.
	bodyRadius = 7
)

// meadow says where the tufts of grass, the flowers and the round bushes of
// the hedge go, in world pixels. The same seed gives the same meadow every
// time.
type meadow struct {
	size    geom.Vec
	tufts   []geom.Vec
	flowers []flower
	bushes  []geom.Vec
}

type flower struct {
	at    geom.Vec
	petal color.NRGBA
}

// newMeadow lays out a meadow of the given size with a hedge around the
// edge. Tufts and flowers only go inside inner.
func newMeadow(size geom.Vec, inner geom.Rect) *meadow {
	m := &meadow{size: size}
	rng := rand.New(rand.NewPCG(1, 2))
	for range 70 {
		m.tufts = append(m.tufts, randomSpot(rng, inner))
	}
	petals := []color.NRGBA{petalWhite, petalPink}
	for range 40 {
		at := randomSpot(rng, inner)
		m.flowers = append(m.flowers, flower{at: at, petal: petals[rng.IntN(len(petals))]})
	}
	t := float64(hedgeWidth)
	for x := t / 2; x < size.X; x += 12 {
		m.bushes = append(m.bushes, geom.Vec{X: x, Y: t - 2}, geom.Vec{X: x, Y: size.Y - t + 2})
	}
	for y := t / 2; y < size.Y; y += 12 {
		m.bushes = append(m.bushes, geom.Vec{X: t - 2, Y: y}, geom.Vec{X: size.X - t + 2, Y: y})
	}
	return m
}

func randomSpot(rng *rand.Rand, r geom.Rect) geom.Vec {
	return geom.Vec{
		X: r.Min.X + rng.Float64()*(r.Max.X-r.Min.X),
		Y: r.Min.Y + rng.Float64()*(r.Max.Y-r.Min.Y),
	}
}

// reach is how far a tuft, a flower or a bush reaches from its spot, in
// world pixels. Ones further than this outside the view are not drawn.
const reach = 6

// drawMeadow draws the grass, tufts and flowers of the meadow, with a hedge
// around the edge. It skips the tufts, flowers and bushes outside visible.
func drawMeadow(p *pen, m *meadow, visible geom.Rect) {
	near := visible.Inset(-reach)
	p.rect(0, 0, m.size.X, m.size.Y, grass)
	for _, t := range m.tufts {
		if near.Contains(t) {
			drawTuft(p, t.X, t.Y)
		}
	}
	for _, f := range m.flowers {
		if near.Contains(f.at) {
			drawFlower(p, f.at.X, f.at.Y, f.petal)
		}
	}
	drawHedge(p, m, near)
}

// drawTuft draws three pointed blades of grass standing on y.
func drawTuft(p *pen, x, y float64) {
	p.polygon(grassDark, x-2.6, y, x-1.4, y, x-2.4, y-2.4)
	p.polygon(grassDark, x-0.6, y, x+0.6, y, x, y-3.4)
	p.polygon(grassDark, x+1.4, y, x+2.6, y, x+2.4, y-2.4)
}

// drawFlower draws four round petals around a heart.
func drawFlower(p *pen, x, y float64, petal color.NRGBA) {
	p.circle(x-0.9, y, 0.7, petal)
	p.circle(x+0.9, y, 0.7, petal)
	p.circle(x, y-0.9, 0.7, petal)
	p.circle(x, y+0.9, 0.7, petal)
	p.circle(x, y, 0.6, flowerHeart)
}

// drawHedge draws a band of hedge around the edge of the meadow, with round
// bushes along its inner side. Only the bushes inside near are drawn.
func drawHedge(p *pen, m *meadow, near geom.Rect) {
	t := float64(hedgeWidth)
	w, h := m.size.X, m.size.Y
	p.rect(0, 0, w, t, hedgeDark)
	p.rect(0, h-t, w, t, hedgeDark)
	p.rect(0, 0, t, h, hedgeDark)
	p.rect(w-t, 0, t, h, hedgeDark)
	for _, b := range m.bushes {
		if near.Contains(b) {
			drawBush(p, b.X, b.Y)
		}
	}
}

func drawBush(p *pen, x, y float64) {
	p.circle(x, y, 5, grassDark)
	p.ellipse(x-2, y-2.5, 1.2, 0.6, grass)
}

// drawCritter draws the placeholder critter, a round orange fox, with its
// feet at the world point x, y and its body lifted lift world pixels off the
// ground.
func drawCritter(p *pen, x, y, lift float64, facing entity.Facing) {
	// The shadow stays on the ground while the critter hops.
	p.ellipse(x, y, 5, 2, shadow)

	by := y - bodyRadius - lift // centre of the body
	p.polygon(fur, x-6, by-3, x-5, by-10, x-1, by-6)
	p.polygon(fur, x+6, by-3, x+5, by-10, x+1, by-6)

	switch facing {
	case entity.FacingUp:
		p.circle(x, by, bodyRadius, fur)
		p.circle(x, by+5, 2, cream)
	case entity.FacingLeft:
		p.circle(x+7, by+3, 2, cream)
		p.circle(x, by, bodyRadius, fur)
		p.circle(x-2, by+2, 3, cream)
		p.ellipse(x-3.5, by-1, 0.6, 1.05, ink)
		p.circle(x-5.5, by+0.5, 0.6, cocoa)
	case entity.FacingRight:
		p.circle(x-7, by+3, 2, cream)
		p.circle(x, by, bodyRadius, fur)
		p.circle(x+2, by+2, 3, cream)
		p.ellipse(x+3.5, by-1, 0.6, 1.05, ink)
		p.circle(x+5.5, by+0.5, 0.6, cocoa)
	default:
		p.circle(x, by, bodyRadius, fur)
		p.circle(x, by+3, 4, cream)
		p.ellipse(x-2.5, by-1, 0.6, 1.05, ink)
		p.ellipse(x+2.5, by-1, 0.6, 1.05, ink)
		p.ellipse(x, by+0.5, 1.1, 0.6, cocoa)
	}
}

// fillBubble fills a speech bubble with rounded corners of radius r. The
// bubble is w by h, centred on x, with its bottom edge at bottom. Its tail is
// tail long and points down at tailX, kept clear of the corners.
func fillBubble(dst *ebiten.Image, x, bottom, w, h, r, tailX, tail float32, clr color.Color) {
	left, right, top := x-w/2, x+w/2, bottom-h
	tailX = max(left+r+tail, min(tailX, right-r-tail))
	var p vector.Path
	p.MoveTo(left+r, top)
	p.LineTo(right-r, top)
	p.ArcTo(right, top, right, top+r, r)
	p.LineTo(right, bottom-r)
	p.ArcTo(right, bottom, right-r, bottom, r)
	p.LineTo(tailX+tail, bottom)
	p.LineTo(tailX, bottom+tail)
	p.LineTo(tailX-tail, bottom)
	p.LineTo(left+r, bottom)
	p.ArcTo(left, bottom, left, bottom-r, r)
	p.LineTo(left, top+r)
	p.ArcTo(left, top, left+r, top, r)
	p.Close()
	op := &vector.DrawPathOptions{AntiAlias: true}
	op.ColorScale.ScaleWithColor(clr)
	vector.FillPath(dst, &p, nil, op)
}

// fade returns c with its opacity multiplied by a, from 0 to 1.
func fade(c color.NRGBA, a float64) color.NRGBA {
	c.A = uint8(math.Round(float64(c.A) * math.Max(0, math.Min(1, a))))
	return c
}
