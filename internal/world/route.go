package world

import (
	"math"

	"github.com/Elpulgo/critterkingdom/internal/geom"
)

// router plans walks for the critter's feet box on a collision grid. Points
// are feet positions: the bottom centre of a box of size box. grid must not be
// nil and its TileSize must be positive.
type router struct {
	grid *Grid
	box  geom.Vec // width and height of the feet box
}

// cellAt returns the cell holding p, clamped into the grid so a point on the
// map's bottom or right edge counts as being in the last row or column.
func (g *Grid) cellAt(p geom.Vec) Cell {
	clamp := func(v float64, n int) int {
		return max(0, min(int(math.Floor(v/g.TileSize)), n-1))
	}
	return Cell{Col: clamp(p.X, g.Cols), Row: clamp(p.Y, g.Rows)}
}

// feetBox returns the box whose bottom centre is at feet.
func (r router) feetBox(feet geom.Vec) geom.Rect {
	return geom.Rect{
		Min: geom.Vec{X: feet.X - r.box.X/2, Y: feet.Y - r.box.Y},
		Max: geom.Vec{X: feet.X + r.box.X/2, Y: feet.Y},
	}
}

// standingCell returns the cell holding the centre of the feet box.
func (r router) standingCell(feet geom.Vec) Cell {
	return r.grid.cellAt(geom.Vec{X: feet.X, Y: feet.Y - r.box.Y/2})
}

// waypoint returns the feet position that centres the box in c.
func (r router) waypoint(c Cell) geom.Vec {
	t := r.grid.TileSize
	return geom.Vec{
		X: float64(c.Col)*t + t/2,
		Y: float64(c.Row)*t + (t+r.box.Y)/2,
	}
}

// clear reports whether the box can move in a straight line from feet a to
// feet b without overlapping a solid cell with positive area at any moment,
// the same overlap rule Grid.Slide uses.
func (r router) clear(a, b geom.Vec) bool {
	from, to := r.feetBox(a), r.feetBox(b)
	d := b.Sub(a)
	swept := geom.Rect{
		Min: geom.Vec{X: math.Min(from.Min.X, to.Min.X), Y: math.Min(from.Min.Y, to.Min.Y)},
		Max: geom.Vec{X: math.Max(from.Max.X, to.Max.X), Y: math.Max(from.Max.Y, to.Max.Y)},
	}
	t := r.grid.TileSize
	c0, r0, c1, r1 := r.grid.cellRange(swept)
	for row := r0; row <= r1; row++ {
		for col := c0; col <= c1; col++ {
			if !r.grid.Solid(col, row) {
				continue
			}
			// The open time spans in which the moving box overlaps the
			// cell on each axis; it hits the cell if they meet in [0, 1].
			xLo, xHi := overlapSpan(from.Min.X, from.Max.X, d.X, float64(col)*t, float64(col+1)*t)
			yLo, yHi := overlapSpan(from.Min.Y, from.Max.Y, d.Y, float64(row)*t, float64(row+1)*t)
			lo, hi := math.Max(xLo, yLo), math.Min(xHi, yHi)
			if lo < hi && lo < 1 && hi > 0 {
				return false
			}
		}
	}
	return true
}

// overlapSpan returns the open span of times t in which the interval
// [lo+t*d, hi+t*d] overlaps (cellLo, cellHi) with positive length. The span is
// empty (start after end) when they never overlap.
func overlapSpan(lo, hi, d, cellLo, cellHi float64) (float64, float64) {
	if d == 0 {
		if lo < cellHi && hi > cellLo {
			return math.Inf(-1), math.Inf(1)
		}
		return math.Inf(1), math.Inf(-1)
	}
	a, b := (cellLo-hi)/d, (cellHi-lo)/d
	if a > b {
		a, b = b, a
	}
	return a, b
}

// plan returns the route for the critter's feet at from to walk to tap, a
// point already inside the walkable area. With a clear straight line the
// route is just tap. Otherwise it follows A* to tap's cell, or to the nearest
// reachable cell when tap's cell is solid or cannot be reached, through cell
// waypoints smoothed by heading for the furthest one the box can move to in a
// straight line. toTap reports whether the route ends on tap; when it does
// not, it ends on the nearest reachable cell's waypoint.
func (r router) plan(from, tap geom.Vec) (route []geom.Vec, toTap bool) {
	if r.clear(from, tap) {
		return []geom.Vec{tap}, true
	}
	start, goal := r.standingCell(from), r.grid.cellAt(tap)
	path, toTap := r.grid.FindPath(start, goal)
	if !toTap {
		near, ok := r.grid.NearestReachable(start, goal)
		if !ok {
			// The critter is not in an open cell, which Slide never allows;
			// walk straight as before rather than not at all.
			return []geom.Vec{tap}, true
		}
		path, _ = r.grid.FindPath(start, near)
	}

	points := make([]geom.Vec, 0, len(path)+1)
	for _, c := range path {
		points = append(points, r.waypoint(c))
	}
	if toTap {
		points = append(points, tap)
	}

	pos := from
	for i := 0; i < len(points); {
		k := len(points) - 1
		for k > i && !r.clear(pos, points[k]) {
			k--
		}
		if points[k] != pos {
			route = append(route, points[k])
			pos = points[k]
		}
		i = k + 1
	}
	if len(route) == 0 {
		route = []geom.Vec{pos}
	}
	return route, toTap
}
