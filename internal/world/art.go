package world

import (
	"hash/fnv"
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"github.com/Elpulgo/critterkingdom/internal/anim"
	"github.com/Elpulgo/critterkingdom/internal/combat"
	"github.com/Elpulgo/critterkingdom/internal/entity"
	"github.com/Elpulgo/critterkingdom/internal/geom"
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
	plum        = color.NRGBA{0x68, 0x38, 0x6c, 0xff}
	orchid      = color.NRGBA{0xb5, 0x50, 0x88, 0xff}
	pineGreen   = color.NRGBA{0x26, 0x5c, 0x42, 0xff}
	poppyRed    = color.NRGBA{0xe4, 0x3b, 0x44, 0xff}
	berryRed    = color.NRGBA{0xa2, 0x26, 0x33, 0xff}
	taupe       = color.NRGBA{0xc2, 0x85, 0x69, 0xff}
	debugColor  = color.NRGBA{0xfe, 0xe7, 0x61, 0xff}
	missingTile = color.NRGBA{0xff, 0x00, 0xff, 0xff}
	zoneDebug   = color.NRGBA{0xff, 0x00, 0x44, 0xff}
	sparkle     = color.NRGBA{0xfe, 0xe7, 0x61, 0xff}
)

// wildBodies are the body colours of placeholder wild critters, one picked
// per species. Greens and light blues are left out, so a critter stands out
// on grass and by water.
var wildBodies = []color.NRGBA{
	{0xbe, 0x4a, 0x2f, 0xff},
	{0xfe, 0xae, 0x34, 0xff},
	{0xb5, 0x50, 0x88, 0xff},
	{0x12, 0x4e, 0x89, 0xff},
	{0xe8, 0xb7, 0x96, 0xff},
	{0x68, 0x38, 0x6c, 0xff},
	{0xc0, 0xcb, 0xdc, 0xff},
	{0xe4, 0x3b, 0x44, 0xff},
}

// villagerCoats are the coat colours of placeholder villagers, picked by the
// villager's place among the map's villagers so neighbours look different.
var villagerCoats = []color.NRGBA{
	{0x5a, 0x69, 0x88, 0xff}, // slate blue
	{0xb8, 0x6f, 0x50, 0xff}, // clay
	{0x68, 0x38, 0x6c, 0xff}, // plum
	{0x26, 0x5c, 0x42, 0xff}, // pine
}

const (
	// critterHeight is how far the tips of the critter's ears are above its
	// feet, in world pixels.
	critterHeight = 17
	// villagerHeight is how far the tips of a villager's ears are above its
	// feet, in world pixels.
	villagerHeight = 20
	// eyeRadius is how wide each of the critter's eyes is from its middle
	// to its side, in world pixels.
	eyeRadius = 0.6
	// reach is how far a prop shape can extend beyond its object rectangle, in
	// world pixels. Props further than this outside the view are not drawn.
	reach = 6
	// wildRadius is the radius of a placeholder wild critter's body, about
	// half the size of the player's.
	wildRadius = 4
	// wildReach is how far a wild critter's drawing, hop, sparkles and
	// notice mark included, can extend from its feet, in world pixels.
	wildReach = 23
	// noticeTop is how far above a wild critter's feet the top of its notice
	// mark is, in world pixels; it sits clear of the top sparkle.
	noticeTop = 21.5
	// sparkleCount is how many stars twinkle around a rare critter.
	sparkleCount = 3
	// sparklePace is how fast a star twinkles, in radians per second.
	sparklePace = 2 * math.Pi / 1.2
	// enemyReach is how far an enemy's drawing, with its flowers and shake,
	// can extend beyond its sprite box, in world pixels.
	enemyReach = 6
	// shakeRate is how many times a second a telegraphing Thistle Beast
	// shakes from side to side, and shakeWidth how far, in world pixels.
	shakeRate  = 12
	shakeWidth = 1.0
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

// drawCritter draws the placeholder critter from its posed parts, in their
// drawing order, with its feet at the world point feet and lifted lift world
// pixels off the ground. Each part is a shape drawn in code, picked by the
// part's name, filling the part's box; the exported part drawings replace
// these shapes later.
func drawCritter(p *pen, feet geom.Vec, lift float64, poses []anim.PartPose) {
	// The shadow stays on the ground while the critter hops.
	p.ellipse(feet.X, feet.Y, 5, 2, shadow)
	origin := feet.Sub(geom.Vec{Y: lift})
	for _, pp := range poses {
		drawPart(p, origin, pp)
	}
}

// drawPart draws the placeholder shape of one posed part, its feet drawn at
// the world point origin: ears are fur triangles standing on the bottom of
// their box, eyes are dark ovals, feet are cocoa ovals, and the body, head,
// tail and any part without a shape of its own are fur ovals.
func drawPart(p *pen, origin geom.Vec, pp anim.PartPose) {
	b := pp.Box
	c := b.Min.Add(b.Max).Scale(0.5)
	w, h := b.Max.X-b.Min.X, b.Max.Y-b.Min.Y
	switch pp.Name {
	case "ear_left", "ear_right":
		p.posedPolygon(origin, pp, fur, b.Min.X, b.Max.Y, b.Max.X, b.Max.Y, c.X, b.Min.Y)
		p.posedPolygon(origin, pp, cream, b.Min.X+0.3*w, b.Max.Y, b.Max.X-0.3*w, b.Max.Y, c.X, b.Min.Y+0.4*h)
	case "eyes":
		// A box wide enough holds a pair of eyes at its ends, as seen from
		// the front; a narrow one holds the single eye of the side view.
		rx := math.Min(eyeRadius, w/2)
		if w > 4*rx {
			p.posedEllipse(origin, pp, b.Min.X+rx, c.Y, rx, h/2, ink)
			p.posedEllipse(origin, pp, b.Max.X-rx, c.Y, rx, h/2, ink)
		} else {
			p.posedEllipse(origin, pp, c.X, c.Y, rx, h/2, ink)
		}
	case "foot_left", "foot_right":
		p.posedEllipse(origin, pp, c.X, c.Y, w/2, h/2, cocoa)
	default:
		p.posedEllipse(origin, pp, c.X, c.Y, w/2, h/2, fur)
	}
}

// speciesColor returns the body colour of a placeholder critter of the
// species with the given id. The same id always gets the same colour.
func speciesColor(id string) color.NRGBA {
	h := fnv.New32a()
	h.Write([]byte(id))
	return wildBodies[h.Sum32()%uint32(len(wildBodies))]
}

// drawWildCritter draws a placeholder wild critter, a round blob in the body
// colour, with its feet at the world point x, y and its body lifted lift world
// pixels off the ground. Its eyes show which way it faces.
func drawWildCritter(p *pen, x, y, lift float64, facing entity.Facing, body color.NRGBA) {
	// The shadow stays on the ground while the critter hops.
	p.ellipse(x, y, 3.5, 1.4, shadow)

	by := y - wildRadius - lift // centre of the body
	p.ellipse(x, by, wildRadius+0.5, wildRadius, body)
	p.ellipse(x-1.2, by-1.8, 1.4, 0.8, fade(white, 0.4))

	switch facing {
	case entity.FacingUp:
		// Seen from behind: no face, just a little tail.
		p.circle(x, by+2.5, 1, fade(white, 0.6))
	case entity.FacingLeft:
		p.ellipse(x-2.6, by-0.3, 0.5, 0.9, ink)
	case entity.FacingRight:
		p.ellipse(x+2.6, by-0.3, 0.5, 0.9, ink)
	default:
		p.ellipse(x-1.5, by-0.3, 0.5, 0.9, ink)
		p.ellipse(x+1.5, by-0.3, 0.5, 0.9, ink)
	}
}

// drawSparkle draws sparkleCount four-point stars around a rare critter whose
// body is centred on x, y. Each star grows and shrinks in turn as t, in
// seconds, goes on; phase shifts the twinkle so neighbouring critters do not
// twinkle in step.
func drawSparkle(p *pen, x, y, t, phase float64) {
	for i := range sparkleCount {
		a := 2 * math.Pi * float64(i) / sparkleCount
		sx := x + 7*math.Cos(a-math.Pi/2)
		sy := y + 6*math.Sin(a-math.Pi/2)
		size := math.Sin(sparklePace*t + phase + a)
		if size <= 0 {
			continue // this star is between twinkles
		}
		drawStar(p, sx, sy, 2.2*size, sparkle)
	}
}

// drawNotice draws the "!" over a wild critter that has noticed the player,
// with the top of the mark at the world point x, top. It is a placeholder for
// the shared exclamation emote of the art bible.
func drawNotice(p *pen, x, top float64) {
	// An ink outline first, then the white mark on it, so the "!" reads on
	// grass, water and sand alike.
	drawExclaim(p, x, top, 0.5, ink)
	drawExclaim(p, x, top, 0, white)
}

// drawExclaim draws a "!" 7 world pixels tall with its top at x, top, every
// edge pushed out by grow.
func drawExclaim(p *pen, x, top, grow float64, clr color.NRGBA) {
	// A bar that narrows towards its foot, then a dot below it.
	p.polygon(clr,
		x-0.9-grow, top-grow, x+0.9+grow, top-grow,
		x+0.5+grow, top+4.6+grow, x-0.5-grow, top+4.6+grow,
	)
	p.circle(x, top+6.1, 0.85+grow, clr)
}

// drawStar draws a four-point star centred on x, y whose points are r from
// its centre.
func drawStar(p *pen, x, y, r float64, clr color.NRGBA) {
	in := r * 0.3
	p.polygon(clr,
		x, y-r, x+in, y-in,
		x+r, y, x+in, y+in,
		x, y+r, x-in, y+in,
		x-r, y, x-in, y-in,
	)
}

// drawVillager draws a placeholder villager, a round animal in a coat, with
// its feet at the world point x, y. index picks the coat colour.
func drawVillager(p *pen, x, y float64, facing entity.Facing, index int) {
	coat := villagerCoats[index%len(villagerCoats)]
	p.ellipse(x, y, 5, 2, shadow)

	// A coat for a body and a round head with two round ears on top.
	p.ellipse(x, y-5, 5, 5, coat)
	hy := y - 13 // centre of the head
	p.circle(x-3.5, hy-4.5, 1.8, stoneGrey)
	p.circle(x+3.5, hy-4.5, 1.8, stoneGrey)
	p.circle(x, hy, 5, stoneGrey)

	switch facing {
	case entity.FacingUp:
		// The back of the head, and the bow of the coat's belt.
		p.circle(x, y-5, 1.5, cream)
	case entity.FacingLeft:
		p.ellipse(x-2, hy+1, 2.6, 2.4, cream)
		p.ellipse(x-2.5, hy-0.5, 0.6, 1.05, ink)
		p.circle(x-4.6, hy+1.2, 0.7, cocoa)
	case entity.FacingRight:
		p.ellipse(x+2, hy+1, 2.6, 2.4, cream)
		p.ellipse(x+2.5, hy-0.5, 0.6, 1.05, ink)
		p.circle(x+4.6, hy+1.2, 0.7, cocoa)
	default:
		p.ellipse(x, hy+1, 3.2, 2.6, cream)
		p.ellipse(x-1.8, hy-0.5, 0.6, 1.05, ink)
		p.ellipse(x+1.8, hy-0.5, 0.6, 1.05, ink)
		p.ellipse(x, hy+1.2, 1.1, 0.6, cocoa)
		p.rect(x-0.5, y-8, 1, 6, fade(cream, 0.6))
	}
}

// drawEnemy draws the placeholder art of e, with its feet at e.Pos: a grumpy
// face while it fights, its telegraph (the slime squashes, the beast shakes,
// the mushroom puffs up), a smiling face among flowers once defeated, and a
// sleep mark above it while it sleeps. elapsed is the scene's game time in seconds, which drives the shake. An
// enemy without a drawing is a magenta square, so it stands out.
func drawEnemy(p *pen, e *entity.Enemy, elapsed float64) {
	x, y := e.Pos.X, e.Pos.Y
	phase := e.Phase()
	happy := e.Defeated()
	p.ellipse(x, y, 6, 2, shadow)
	switch e.Def().ID {
	case "sulky_slime":
		drawSlime(p, x, y, phase, happy)
	case "thistle_beast":
		if phase == combat.Telegraph {
			x += shakeWidth * math.Sin(2*math.Pi*shakeRate*elapsed)
		}
		drawThistleBeast(p, x, y, happy)
	case "grumpy_mushroom":
		drawMushroom(p, x, y, phase == combat.Telegraph, happy)
	default:
		p.rect(x-combat.EnemySize/2, y-combat.EnemySize, combat.EnemySize, combat.EnemySize, missingTile)
		return
	}
	if happy {
		drawPoofFlowers(p, x, y-combat.EnemySize/2)
	}
	if e.Asleep() {
		drawSleepMark(p, x+5, y-combat.EnemySize-1)
	}
}

// drawSleepMark draws the placeholder sleep mark, a big and a small "z"
// rising to the upper right of x, y, the bottom left corner of the big one.
func drawSleepMark(p *pen, x, y float64) {
	drawZ(p, x, y, 4, ink)
	drawZ(p, x+4.5, y-4.5, 2.5, ink)
}

// drawZ draws a "z" of size s with its bottom left corner at x, y.
func drawZ(p *pen, x, y, s float64, clr color.NRGBA) {
	const thick = 0.7
	p.rect(x, y-s, s, thick, clr)
	p.rect(x, y-thick, s, thick, clr)
	p.polygon(clr, x+s-thick, y-s+thick, x+s, y-s+thick, x+thick, y-thick, x, y-thick)
}

// drawSlime draws the Sulky Slime, an orchid blob: wide and flat while it
// squashes for its hop, lifted off the ground while it hops.
func drawSlime(p *pen, x, y float64, phase combat.Phase, happy bool) {
	rx, ry, lift := 6.5, 5.5, 0.0
	switch phase {
	case combat.Telegraph:
		rx, ry = 7.5, 4.0
	case combat.Attack:
		rx, ry, lift = 5.8, 6.2, 3
	}
	cy := y - ry - lift
	p.ellipse(x, cy, rx+1, ry+1, plum)
	p.ellipse(x, cy, rx, ry, orchid)
	p.ellipse(x-rx*0.45, cy-ry*0.45, rx*0.3, ry*0.2, petalPink)
	drawFace(p, x, cy+ry*0.15, happy)
}

// drawThistleBeast draws the Thistle Beast, a round leafy body with a crown of
// orchid thistle spikes, soft and rounded rather than sharp.
func drawThistleBeast(p *pen, x, y float64, happy bool) {
	cy := y - 6
	for i := -2; i <= 2; i++ {
		tx := x + float64(i)*2.4
		ty := cy - 6 + math.Abs(float64(i))*0.8
		p.circle(tx, ty, 2.0, plum)
		p.circle(tx, ty, 1.4, orchid)
	}
	p.ellipse(x, cy, 7.5, 6.5, pineGreen)
	p.ellipse(x, cy, 6.5, 5.5, grassDark)
	p.ellipse(x-3, cy-3, 1.8, 1.0, grass)
	drawFace(p, x, cy+0.5, happy)
}

// drawMushroom draws the Grumpy Mushroom, a red spotted cap on a cream stem
// with its face on the stem. puffed swells the cap while it gathers spores.
func drawMushroom(p *pen, x, y float64, puffed, happy bool) {
	k := 1.0
	if puffed {
		k = 1.25
	}
	p.rect(x-4.5, y-7, 9, 7, taupe)
	p.rect(x-3.8, y-7, 7.6, 6.3, cream)
	capY := y - 9
	p.ellipse(x, capY, 8*k+1, 4.5*k+1, berryRed)
	p.ellipse(x, capY, 8*k, 4.5*k, poppyRed)
	p.circle(x-3.5*k, capY-1*k, 1.2*k, white)
	p.circle(x+2.5*k, capY-2*k, 1.0*k, white)
	p.circle(x+4.5*k, capY+0.8*k, 0.8*k, white)
	drawFace(p, x, y-3.5, happy)
}

// drawFace draws a small face centred on x, y: big dark eyes with a white
// highlight, and either a grumpy look (eyebrows tilted inward and a little
// pout) or a happy smile.
func drawFace(p *pen, x, y float64, happy bool) {
	for _, side := range []float64{-1, 1} {
		ex := x + side*2.2
		p.ellipse(ex, y-1, 0.9, 1.2, ink)
		p.circle(ex-0.3, y-1.5, 0.35, white)
		if !happy {
			// The inner end of each eyebrow sits lower: grumpy, not angry.
			inner, outer := x+side*1.0, x+side*3.4
			p.polygon(cocoa, outer, y-3.4, inner, y-2.6, inner, y-2.0, outer, y-2.8)
		}
	}
	if happy {
		drawArc(p, x, y+0.6, 1.6, true, cocoa)
	} else {
		drawArc(p, x, y+2.0, 1.0, false, cocoa)
	}
}

// drawArc draws a half-world-pixel-thick half ring of radius r around x, y:
// the lower half, a smile, when smile is true, else the upper half, a pout.
func drawArc(p *pen, x, y, r float64, smile bool, clr color.NRGBA) {
	const (
		steps = 8
		thick = 0.5
	)
	dir := 1.0
	if !smile {
		dir = -1
	}
	xy := make([]float64, 0, 4*(steps+1))
	for i := 0; i <= steps; i++ {
		a := math.Pi * float64(i) / steps
		xy = append(xy, x+r*math.Cos(a), y+dir*r*math.Sin(a))
	}
	for i := steps; i >= 0; i-- {
		a := math.Pi * float64(i) / steps
		xy = append(xy, x+(r-thick)*math.Cos(a), y+dir*(r-thick)*math.Sin(a))
	}
	p.polygon(clr, xy...)
}

// drawPoofFlowers draws the ring of flowers a defeated enemy poofs into,
// around x, y.
func drawPoofFlowers(p *pen, x, y float64) {
	for i := 0; i < 6; i++ {
		a := math.Pi * float64(i) / 3
		petal := petalWhite
		if i%2 != 0 {
			petal = petalPink
		}
		drawFlower(p, x+9*math.Cos(a), y+7*math.Sin(a), petal)
	}
}

// drawSpore draws a puffed spore filling the box r: a big soft cloud.
func drawSpore(p *pen, r geom.Rect) {
	c := r.Min.Add(r.Max).Scale(0.5)
	w := (r.Max.X - r.Min.X) / 2
	soft := fade(stoneLight, 0.85)
	p.circle(c.X-w*0.4, c.Y+w*0.15, w*0.6, soft)
	p.circle(c.X+w*0.4, c.Y+w*0.15, w*0.6, soft)
	p.circle(c.X, c.Y-w*0.2, w*0.7, soft)
	p.circle(c.X-w*0.15, c.Y-w*0.35, w*0.35, fade(white, 0.9))
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
