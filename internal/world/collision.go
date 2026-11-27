package world

import (
	"math"

	"github.com/Elpulgo/critterkingdom/internal/geom"
)

// Grid is a tile-aligned grid of solid and open cells in world pixels.
type Grid struct {
	Cols     int
	Rows     int
	TileSize float64
	solid    []bool
}

// buildCollisionGrid creates a Grid from the map's "collision" layer, which
// may be a tile layer (every non-zero GID is solid), an object layer (every
// cell an object's rectangle overlaps with positive area is solid), or absent
// (all cells are open).
func buildCollisionGrid(m *Map) *Grid {
	g := &Grid{
		Cols:     m.Cols,
		Rows:     m.Rows,
		TileSize: m.TileSize,
		solid:    make([]bool, m.Cols*m.Rows),
	}

	if tl := m.TileLayer("collision"); tl != nil {
		for i, gid := range tl.GIDs {
			if gid != 0 {
				g.solid[i] = true
			}
		}
	}

	if ol := m.ObjectLayer("collision"); ol != nil {
		for _, obj := range ol.Objects {
			r := obj.Rect()
			if r.Max.X <= r.Min.X || r.Max.Y <= r.Min.Y {
				continue
			}
			c0, r0, c1, r1 := g.cellRange(r)
			for row := r0; row <= r1; row++ {
				for col := c0; col <= c1; col++ {
					if col >= 0 && col < g.Cols && row >= 0 && row < g.Rows {
						g.solid[row*g.Cols+col] = true
					}
				}
			}
		}
	}

	return g
}

// Solid reports whether cell (col, row) is solid. Cells outside the grid are
// not solid.
func (g *Grid) Solid(col, row int) bool {
	if g == nil || col < 0 || col >= g.Cols || row < 0 || row >= g.Rows {
		return false
	}
	return g.solid[row*g.Cols+col]
}

// Slide moves box by delta one axis at a time (X then Y), shortening each
// component so the box stops flush against any solid cell it would overlap.
// Touching a solid cell's edge without positive-area overlap does not block
// movement along that edge. It returns the part of delta the box can move.
func (g *Grid) Slide(box geom.Rect, delta geom.Vec) geom.Vec {
	if g == nil || g.TileSize <= 0 || box.Max.X <= box.Min.X || box.Max.Y <= box.Min.Y {
		return delta
	}

	dx := g.slideX(box, delta.X)
	box.Min.X += dx
	box.Max.X += dx

	dy := g.slideY(box, delta.Y)
	return geom.Vec{X: dx, Y: dy}
}

func (g *Grid) slideX(box geom.Rect, dx float64) float64 {
	if dx == 0 {
		return 0
	}
	shifted := geom.Rect{
		Min: geom.Vec{X: box.Min.X + dx, Y: box.Min.Y},
		Max: geom.Vec{X: box.Max.X + dx, Y: box.Max.Y},
	}
	c0, r0, c1, r1 := g.cellRange(shifted)
	for row := r0; row <= r1; row++ {
		for col := c0; col <= c1; col++ {
			if !g.Solid(col, row) {
				continue
			}
			if dx > 0 {
				limit := float64(col)*g.TileSize - box.Max.X
				if limit < 0 {
					limit = 0
				}
				if limit < dx {
					dx = limit
				}
			} else {
				limit := float64(col+1)*g.TileSize - box.Min.X
				if limit > 0 {
					limit = 0
				}
				if limit > dx {
					dx = limit
				}
			}
		}
	}
	return dx
}

func (g *Grid) slideY(box geom.Rect, dy float64) float64 {
	if dy == 0 {
		return 0
	}
	shifted := geom.Rect{
		Min: geom.Vec{X: box.Min.X, Y: box.Min.Y + dy},
		Max: geom.Vec{X: box.Max.X, Y: box.Max.Y + dy},
	}
	c0, r0, c1, r1 := g.cellRange(shifted)
	for row := r0; row <= r1; row++ {
		for col := c0; col <= c1; col++ {
			if !g.Solid(col, row) {
				continue
			}
			if dy > 0 {
				limit := float64(row)*g.TileSize - box.Max.Y
				if limit < 0 {
					limit = 0
				}
				if limit < dy {
					dy = limit
				}
			} else {
				limit := float64(row+1)*g.TileSize - box.Min.Y
				if limit > 0 {
					limit = 0
				}
				if limit > dy {
					dy = limit
				}
			}
		}
	}
	return dy
}

// cellRange returns the inclusive range of cell columns [c0..c1] and rows
// [r0..r1] that r overlaps with positive area.
func (g *Grid) cellRange(r geom.Rect) (c0, r0, c1, r1 int) {
	c0 = int(math.Floor(r.Min.X / g.TileSize))
	c1 = int(math.Ceil(r.Max.X/g.TileSize)) - 1
	r0 = int(math.Floor(r.Min.Y / g.TileSize))
	r1 = int(math.Ceil(r.Max.Y/g.TileSize)) - 1
	return c0, r0, c1, r1
}
