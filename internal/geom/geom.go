// Package geom has the small 2D maths types the game packages share.
package geom

import "math"

// Vec is a point or a direction in pixels. Y grows downwards, as on screen.
type Vec struct {
	X, Y float64
}

// Add returns v + o.
func (v Vec) Add(o Vec) Vec { return Vec{v.X + o.X, v.Y + o.Y} }

// Sub returns v - o.
func (v Vec) Sub(o Vec) Vec { return Vec{v.X - o.X, v.Y - o.Y} }

// Scale returns v multiplied by s.
func (v Vec) Scale(s float64) Vec { return Vec{v.X * s, v.Y * s} }

// Len returns the length of v.
func (v Vec) Len() float64 { return math.Hypot(v.X, v.Y) }

// IsZero reports whether v is the zero vector.
func (v Vec) IsZero() bool { return v.X == 0 && v.Y == 0 }

// Normalized returns v scaled to length 1, or the zero vector when v is zero.
func (v Vec) Normalized() Vec {
	l := v.Len()
	if l == 0 {
		return Vec{}
	}
	return Vec{v.X / l, v.Y / l}
}

// ClampLen returns v shortened to at most length max.
func (v Vec) ClampLen(max float64) Vec {
	if v.Len() <= max {
		return v
	}
	return v.Normalized().Scale(max)
}

// Rect is an axis-aligned rectangle from Min to Max.
type Rect struct {
	Min, Max Vec
}

// Inset returns r shrunk by d on every side.
func (r Rect) Inset(d float64) Rect {
	return Rect{Min: Vec{r.Min.X + d, r.Min.Y + d}, Max: Vec{r.Max.X - d, r.Max.Y - d}}
}

// Clamp returns the point inside r, edges included, that is closest to p.
func (r Rect) Clamp(p Vec) Vec {
	return Vec{
		X: math.Max(r.Min.X, math.Min(p.X, r.Max.X)),
		Y: math.Max(r.Min.Y, math.Min(p.Y, r.Max.Y)),
	}
}
