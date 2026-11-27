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

// Placeholder art drawn with shapes, in colours from the Endesga 32 palette.
// Sprites drawn in Aseprite replace it later.
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

// drawMeadow draws grass, tufts and flowers, with a hedge around the edge.
// Decorations only go inside inner. The same seed gives the same meadow
// every time.
func drawMeadow(dst *ebiten.Image, inner geom.Rect) {
	dst.Fill(grass)
	rng := rand.New(rand.NewPCG(1, 2))
	for range 70 {
		x, y := randomSpot(rng, inner)
		drawTuft(dst, x, y)
	}
	petals := []color.NRGBA{petalWhite, petalPink}
	for range 40 {
		x, y := randomSpot(rng, inner)
		drawFlower(dst, x, y, petals[rng.IntN(len(petals))])
	}
	b := dst.Bounds()
	drawHedge(dst, float32(b.Dx()), float32(b.Dy()))
}

func randomSpot(rng *rand.Rand, r geom.Rect) (x, y float32) {
	x = float32(math.Floor(r.Min.X + rng.Float64()*(r.Max.X-r.Min.X)))
	y = float32(math.Floor(r.Min.Y + rng.Float64()*(r.Max.Y-r.Min.Y)))
	return x, y
}

func drawTuft(dst *ebiten.Image, x, y float32) {
	vector.FillRect(dst, x-2, y-2, 1, 2, grassDark, false)
	vector.FillRect(dst, x, y-3, 1, 3, grassDark, false)
	vector.FillRect(dst, x+2, y-2, 1, 2, grassDark, false)
}

func drawFlower(dst *ebiten.Image, x, y float32, petal color.NRGBA) {
	vector.FillRect(dst, x-1, y, 3, 1, petal, false)
	vector.FillRect(dst, x, y-1, 1, 3, petal, false)
	vector.FillRect(dst, x, y, 1, 1, flowerHeart, false)
}

// drawHedge draws a band of hedge around the edge of a w by h meadow, with
// round bushes along its inner side.
func drawHedge(dst *ebiten.Image, w, h float32) {
	t := float32(hedgeWidth)
	vector.FillRect(dst, 0, 0, w, t, hedgeDark, false)
	vector.FillRect(dst, 0, h-t, w, t, hedgeDark, false)
	vector.FillRect(dst, 0, 0, t, h, hedgeDark, false)
	vector.FillRect(dst, w-t, 0, t, h, hedgeDark, false)
	for x := t / 2; x < w; x += 12 {
		drawBush(dst, x, t-2)
		drawBush(dst, x, h-t+2)
	}
	for y := t / 2; y < h; y += 12 {
		drawBush(dst, t-2, y)
		drawBush(dst, w-t+2, y)
	}
}

func drawBush(dst *ebiten.Image, x, y float32) {
	vector.FillCircle(dst, x, y, 5, grassDark, false)
	vector.FillRect(dst, x-3, y-3, 2, 1, grass, false)
}

// drawCritter draws the placeholder critter, a round orange fox, with its
// feet at x, y and its body lifted off the ground by lift pixels.
func drawCritter(dst *ebiten.Image, x, y, lift float32, facing entity.Facing) {
	// The shadow stays on the ground while the critter hops.
	vector.FillRect(dst, x-4, y-2, 8, 1, shadow, false)
	vector.FillRect(dst, x-5, y-1, 10, 2, shadow, false)
	vector.FillRect(dst, x-4, y+1, 8, 1, shadow, false)

	by := y - bodyRadius - lift // centre of the body
	fillTriangle(dst, x-6, by-3, x-5, by-10, x-1, by-6, fur)
	fillTriangle(dst, x+6, by-3, x+5, by-10, x+1, by-6, fur)

	switch facing {
	case entity.FacingUp:
		vector.FillCircle(dst, x, by, bodyRadius, fur, false)
		vector.FillCircle(dst, x, by+5, 2, cream, false)
	case entity.FacingLeft:
		vector.FillCircle(dst, x+7, by+3, 2, cream, false)
		vector.FillCircle(dst, x, by, bodyRadius, fur, false)
		vector.FillCircle(dst, x-2, by+2, 3, cream, false)
		vector.FillRect(dst, x-4, by-2, 1, 2, ink, false)
		vector.FillRect(dst, x-6, by, 1, 1, cocoa, false)
	case entity.FacingRight:
		vector.FillCircle(dst, x-7, by+3, 2, cream, false)
		vector.FillCircle(dst, x, by, bodyRadius, fur, false)
		vector.FillCircle(dst, x+2, by+2, 3, cream, false)
		vector.FillRect(dst, x+3, by-2, 1, 2, ink, false)
		vector.FillRect(dst, x+5, by, 1, 1, cocoa, false)
	default:
		vector.FillCircle(dst, x, by, bodyRadius, fur, false)
		vector.FillCircle(dst, x, by+3, 4, cream, false)
		vector.FillRect(dst, x-3, by-2, 1, 2, ink, false)
		vector.FillRect(dst, x+2, by-2, 1, 2, ink, false)
		vector.FillRect(dst, x-1, by, 2, 1, cocoa, false)
	}
}

func fillTriangle(dst *ebiten.Image, x0, y0, x1, y1, x2, y2 float32, clr color.Color) {
	var p vector.Path
	p.MoveTo(x0, y0)
	p.LineTo(x1, y1)
	p.LineTo(x2, y2)
	p.Close()
	op := &vector.DrawPathOptions{}
	op.ColorScale.ScaleWithColor(clr)
	vector.FillPath(dst, &p, nil, op)
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
