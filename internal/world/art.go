package world

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"github.com/Elpulgo/critterkingdom/internal/entity"
)

// Placeholder art drawn in code with smooth shapes, in colours from the
// Endesga 32 palette. Sprites made from vector drawings replace it later.
var (
	grass       = color.NRGBA{0x63, 0xc7, 0x4d, 0xff}
	grassDark   = color.NRGBA{0x3e, 0x89, 0x48, 0xff}
	dirtPath    = color.NRGBA{0xd7, 0x76, 0x43, 0xff}
	waterBlue   = color.NRGBA{0x2c, 0xe8, 0xf5, 0xff}
	waterDeep   = color.NRGBA{0x00, 0x99, 0xdb, 0xff}
	stoneGrey   = color.NRGBA{0x8b, 0x9b, 0xb4, 0xff}
	stoneLight  = color.NRGBA{0xc0, 0xcb, 0xdc, 0xff}
	barkBrown   = color.NRGBA{0x73, 0x3e, 0x39, 0xff}
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
	missingTile = color.NRGBA{0xff, 0x00, 0xff, 0xff}
)

const (
	// critterHeight is how far the tips of the critter's ears are above its
	// feet, in world pixels.
	critterHeight = 17
	// bodyRadius is the radius of the critter's round body.
	bodyRadius = 7
	// reach is how far a prop shape can extend beyond its object rectangle, in
	// world pixels. Props further than this outside the view are not drawn.
	reach = 6
)

// drawTileArt draws the art for a tile or prop of the given type filling the
// w by h rectangle at world position x, y. col and row pick between variants
// of the same type, such as the petal colour of a flower. An unknown type is
// a magenta square, so a missing drawing stands out.
func drawTileArt(p *pen, kind string, x, y, w, h float64, col, row int) {
	cx, cy := x+w/2, y+h/2
	switch kind {
	case "grass":
		p.rect(x, y, w, h, grass)
	case "path":
		p.rect(x, y, w, h, dirtPath)
		p.ellipse(cx-3, cy-2, 2.2, 1.2, cream)
		p.ellipse(cx+3, cy+3, 1.8, 1.0, cream)
	case "water_1":
		p.rect(x, y, w, h, waterDeep)
		p.ellipse(cx-2, cy-3, 3.5, 1.1, waterBlue)
		p.ellipse(cx+2, cy+3, 2.5, 0.9, waterBlue)
	case "water_2":
		p.rect(x, y, w, h, waterDeep)
		p.ellipse(cx, cy-1, 4.0, 1.2, waterBlue)
		p.ellipse(cx-1, cy+4, 2.2, 0.8, waterBlue)
	case "water_3":
		p.rect(x, y, w, h, waterDeep)
		p.ellipse(cx+2, cy-2, 3.2, 1.0, waterBlue)
		p.ellipse(cx-2, cy+2, 3.8, 1.1, waterBlue)
	case "tuft":
		drawTuft(p, cx, cy+2)
	case "flower":
		petal := petalWhite
		if (col+row)%2 != 0 {
			petal = petalPink
		}
		drawFlower(p, cx, cy, petal)
	case "bush":
		drawBush(p, cx, cy)
	case "rock":
		drawRock(p, cx, cy)
	case "tree":
		drawTree(p, cx, cy)
	case "tree_top":
		drawTreeTop(p, cx, cy)
	case "floor":
		drawFloor(p, x, y, w, h)
	case "wall":
		drawWall(p, x, y, w, h, row)
	case "roof":
		drawRoof(p, x, y, w, h)
	case "door":
		drawDoor(p, x, y, w, h)
	default:
		p.rect(x, y, w, h, missingTile)
	}
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

// drawBush draws a round bush centred on x, y.
func drawBush(p *pen, x, y float64) {
	p.circle(x, y, 6, grassDark)
	p.ellipse(x-2, y-2.5, 1.6, 0.9, grass)
}

// drawRock draws a grey boulder on its shadow, centred on x, y.
func drawRock(p *pen, x, y float64) {
	p.ellipse(x, y+2, 6, 2.5, shadow)
	p.ellipse(x, y, 5.5, 4.2, stoneGrey)
	p.ellipse(x-1.5, y-1.5, 2.5, 1.4, stoneLight)
}

// drawTree draws a tree trunk with the bottom of its crown, centred on x, y.
// The rest of the crown is a tree_top tile in the above layer.
func drawTree(p *pen, x, y float64) {
	p.ellipse(x, y+4, 6, 2.5, shadow)
	p.rect(x-2.5, y-4, 5, 8, barkBrown)
	p.circle(x, y-4, 5.5, grassDark)
}

// drawTreeTop draws the part of a tree crown the critter walks behind.
func drawTreeTop(p *pen, x, y float64) {
	p.circle(x, y+2, 7.5, grassDark)
	p.circle(x-2, y-1, 5, grass)
}

// drawFloor draws wooden floorboards: a cream floor with three planks
// running across the tile.
func drawFloor(p *pen, x, y, w, h float64) {
	p.rect(x, y, w, h, cream)
	seam := fade(barkBrown, 0.35)
	for i := 1; i < 3; i++ {
		p.rect(x, y+h*float64(i)/3-0.25, w, 0.5, seam)
	}
	// One butt joint per plank, staggered so the boards do not line up.
	p.rect(x+w*0.3, y, 0.5, h/3, seam)
	p.rect(x+w*0.75, y+h/3, 0.5, h/3, seam)
	p.rect(x+w*0.15, y+h*2/3, 0.5, h/3, seam)
}

// drawWall draws a stone wall of two brick courses. row offsets the vertical
// joints of every other tile row, so a tall wall looks bonded.
func drawWall(p *pen, x, y, w, h float64, row int) {
	p.rect(x, y, w, h, stoneGrey)
	mortar := stoneLight
	p.rect(x, y+h/2-0.3, w, 0.6, mortar)
	p.rect(x, y+h-0.6, w, 0.6, mortar)
	shift := 0.0
	if row%2 != 0 {
		shift = w / 4
	}
	p.rect(x+w/4+shift-0.3, y, 0.6, h/2, mortar)
	p.rect(x+w*3/4-shift-0.3, y+h/2, 0.6, h/2, mortar)
}

// drawRoof draws red roof shingles: two rows of rounded tiles with a dark
// line under each row.
func drawRoof(p *pen, x, y, w, h float64) {
	p.rect(x, y, w, h, petalPink)
	edge := fade(cocoa, 0.45)
	for i := 0; i < 2; i++ {
		bottom := y + h*float64(i+1)/2
		p.rect(x, bottom-0.6, w, 0.6, edge)
		for j := 0; j < 4; j++ {
			p.ellipse(x+w*(float64(j)+0.5)/4, bottom-h/8, w/10, h/14, fade(white, 0.25))
		}
	}
}

// drawDoor draws a wooden door with a rounded top and a brass knob, set in a
// stone wall.
func drawDoor(p *pen, x, y, w, h float64) {
	p.rect(x, y, w, h, stoneGrey)
	dw := w * 0.6
	dx := x + (w-dw)/2
	top := y + h*0.25
	p.circle(x+w/2, top, dw/2, barkBrown)
	p.rect(dx, top, dw, y+h-top, barkBrown)
	p.rect(x+w/2-0.25, top-dw/2+1, 0.5, y+h-top+dw/2-1, cocoa)
	p.circle(dx+dw*0.8, y+h*0.62, 0.8, flowerHeart)
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
