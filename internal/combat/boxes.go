// Package combat has the fight rules for the critter's adventures: hitboxes,
// hurt-boxes, petals, invulnerability, knockback, aim assist and the nap.
package combat

import "github.com/Elpulgo/critterkingdom/internal/geom"

const (
	// HurtBoxShrink is the fraction of a sprite's width and height trimmed off
	// each side to form its hurt-box, leaving the middle 60%.
	HurtBoxShrink = 0.2
	// SwingMargin is how many world pixels wider than the stick's reach the
	// swing hitbox is on each side.
	SwingMargin = 4.0
)

// HurtBox returns the middle of sprite that counts as being hit, smaller than
// the drawn sprite on every side by HurtBoxShrink of that dimension.
func HurtBox(sprite geom.Rect) geom.Rect {
	dx := (sprite.Max.X - sprite.Min.X) * HurtBoxShrink
	dy := (sprite.Max.Y - sprite.Min.Y) * HurtBoxShrink
	return geom.Rect{
		Min: geom.Vec{X: sprite.Min.X + dx, Y: sprite.Min.Y + dy},
		Max: geom.Vec{X: sprite.Max.X - dx, Y: sprite.Max.Y - dy},
	}
}

// SwingBox returns the axis-aligned square hit by a swing starting at origin
// in direction dir with a stick of stickLength world pixels.
func SwingBox(origin, dir geom.Vec, stickLength float64) geom.Rect {
	d := dir.Normalized()
	c := origin.Add(d.Scale(stickLength / 2))
	h := stickLength/2 + SwingMargin
	return geom.Rect{
		Min: geom.Vec{X: c.X - h, Y: c.Y - h},
		Max: geom.Vec{X: c.X + h, Y: c.Y + h},
	}
}

// Overlap reports whether a and b share any point, edges and corners included.
func Overlap(a, b geom.Rect) bool {
	return a.Min.X <= b.Max.X && b.Min.X <= a.Max.X && a.Min.Y <= b.Max.Y && b.Min.Y <= a.Max.Y
}
