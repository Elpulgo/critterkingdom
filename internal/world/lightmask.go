package world

import (
	"math"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/Elpulgo/critterkingdom/internal/engine"
	"github.com/Elpulgo/critterkingdom/internal/geom"
)

// maskMargin is how many texels of ink the mask image has around the view
// on every side. The linear upscale samples half a texel past the view edge,
// and DrawImage gives no guarantee what lies outside an image, so without the
// margin the edge of the darkness would let the world show through.
const maskMargin = 1

// lightSpriteSize is the width and height of the radial light sprite, in
// pixels. The sprite is scaled to each light's diameter when it is drawn.
const lightSpriteSize = 64

// lightMask covers a dark map with darkness except where lights shine. It is
// a render cache only: the mask image is one pixel per world pixel, the view
// plus maskMargin on every side, and is rebuilt when the view size changes;
// the light sprite is built once.
type lightMask struct {
	image  *ebiten.Image
	w, h   int // size of the view the image covers, in world pixels
	sprite *ebiten.Image
}

// draw covers the view on c with ink, except for the lights, which are given
// in world pixels with the view's top-left corner at camera. view is the size
// of the view in world pixels. It draws nothing when the view is empty.
func (lm *lightMask) draw(c engine.Canvas, camera, view geom.Vec, lights []Light) {
	w, h := int(math.Ceil(view.X)), int(math.Ceil(view.Y))
	if w < 1 || h < 1 {
		return
	}
	if lm.image == nil || lm.w != w || lm.h != h {
		if lm.image != nil {
			lm.image.Deallocate()
		}
		lm.image = ebiten.NewImage(w+2*maskMargin, h+2*maskMargin)
		lm.w, lm.h = w, h
	}
	if lm.sprite == nil {
		lm.sprite = newLightSprite()
	}

	lm.image.Fill(ink)
	for _, l := range lights {
		if !(l.Radius > 0) {
			continue
		}
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(-lightSpriteSize/2, -lightSpriteSize/2)
		op.GeoM.Scale(2*l.Radius/lightSpriteSize, 2*l.Radius/lightSpriteSize)
		op.GeoM.Translate(l.Pos.X-camera.X+maskMargin, l.Pos.Y-camera.Y+maskMargin)
		op.Filter = ebiten.FilterLinear
		op.Blend = ebiten.BlendDestinationOut
		lm.image.DrawImage(lm.sprite, op)
	}

	// The margin hangs outside the view, where the canvas clips it.
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(-maskMargin, -maskMargin)
	op.GeoM.Scale(c.Scale, c.Scale)
	op.GeoM.Translate(c.OffX, c.OffY)
	op.Filter = ebiten.FilterLinear
	c.Image.DrawImage(lm.image, op)
}

// newLightSprite returns a white sprite whose alpha at each pixel is the
// brightness of a light centred on the sprite with a radius of half its
// width, so the edge band and the core match Light.Brightness.
func newLightSprite() *ebiten.Image {
	const n = lightSpriteSize
	l := Light{Pos: geom.Vec{X: n / 2, Y: n / 2}, Radius: n / 2}
	pix := make([]byte, 4*n*n)
	for y := range n {
		for x := range n {
			b := l.Brightness(geom.Vec{X: float64(x) + 0.5, Y: float64(y) + 0.5})
			a := byte(math.Round(255 * b))
			// Premultiplied alpha: white at alpha a is a in every channel.
			i := 4 * (y*n + x)
			pix[i], pix[i+1], pix[i+2], pix[i+3] = a, a, a, a
		}
	}
	img := ebiten.NewImage(n, n)
	img.WritePixels(pix)
	return img
}
