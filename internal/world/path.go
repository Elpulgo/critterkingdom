package world

import (
	"container/heap"
	"math"
)

// Cell is a column and row on a Grid.
type Cell struct {
	Col, Row int
}

// neighbours lists the eight steps A* tries, in a fixed order so the same
// grid, start and goal always give the same path. Orthogonal steps come first.
var neighbours = [8]Cell{
	{0, -1}, {1, 0}, {0, 1}, {-1, 0},
	{1, -1}, {1, 1}, {-1, 1}, {-1, -1},
}

// open reports whether c is inside the grid and not solid. Cells outside the
// grid are not open, although Solid reports them as not solid: the critter
// can never stand there.
func (g *Grid) open(c Cell) bool {
	return g != nil && c.Col >= 0 && c.Col < g.Cols && c.Row >= 0 && c.Row < g.Rows &&
		!g.solid[c.Row*g.Cols+c.Col]
}

// canStep reports whether a step from open cell c by d (one of neighbours)
// lands on an open cell. A diagonal step is allowed only when both
// orthogonally adjacent cells are open, so it never cuts a solid corner.
func (g *Grid) canStep(c, d Cell) bool {
	if !g.open(Cell{c.Col + d.Col, c.Row + d.Row}) {
		return false
	}
	if d.Col != 0 && d.Row != 0 {
		return g.open(Cell{c.Col + d.Col, c.Row}) && g.open(Cell{c.Col, c.Row + d.Row})
	}
	return true
}

// octile is the cost of the cheapest eight-direction route from a to b on an
// empty grid: 1 per straight step and √2 per diagonal step.
func octile(a, b Cell) float64 {
	dx := math.Abs(float64(a.Col - b.Col))
	dy := math.Abs(float64(a.Row - b.Row))
	return dx + dy + (math.Sqrt2-2)*math.Min(dx, dy)
}

// FindPath returns the cheapest path of open cells from from to to, both
// included, moving in eight directions without cutting solid corners (cost 1
// straight, √2 diagonal). It returns false when either end is solid or
// outside the grid, or when no path exists. from == to gives a one-cell path.
func (g *Grid) FindPath(from, to Cell) ([]Cell, bool) {
	if !g.open(from) || !g.open(to) {
		return nil, false
	}
	if from == to {
		return []Cell{from}, true
	}

	n := g.Cols * g.Rows
	index := func(c Cell) int { return c.Row*g.Cols + c.Col }
	cost := make([]float64, n)
	for i := range cost {
		cost[i] = math.Inf(1)
	}
	came := make([]int, n)
	closed := make([]bool, n)

	start := index(from)
	cost[start] = 0
	came[start] = -1
	q := &pathQueue{}
	heap.Push(q, pathNode{cell: from, index: start, f: octile(from, to), h: octile(from, to)})

	for q.Len() > 0 {
		cur := heap.Pop(q).(pathNode)
		if closed[cur.index] {
			continue
		}
		closed[cur.index] = true
		if cur.cell == to {
			return g.walkBack(came, cur.index), true
		}
		for _, d := range neighbours {
			if !g.canStep(cur.cell, d) {
				continue
			}
			next := Cell{cur.cell.Col + d.Col, cur.cell.Row + d.Row}
			ni := index(next)
			if closed[ni] {
				continue
			}
			step := 1.0
			if d.Col != 0 && d.Row != 0 {
				step = math.Sqrt2
			}
			c := cost[cur.index] + step
			if c >= cost[ni] {
				continue
			}
			cost[ni] = c
			came[ni] = cur.index
			h := octile(next, to)
			heap.Push(q, pathNode{cell: next, index: ni, f: c + h, h: h})
		}
	}
	return nil, false
}

// walkBack follows came from cell index last back to the start and returns
// the cells in walking order.
func (g *Grid) walkBack(came []int, last int) []Cell {
	var path []Cell
	for i := last; i != -1; i = came[i] {
		path = append(path, Cell{Col: i % g.Cols, Row: i / g.Cols})
	}
	for a, b := 0, len(path)-1; a < b; a, b = a+1, b-1 {
		path[a], path[b] = path[b], path[a]
	}
	return path
}

// pathNode is an entry in the A* open set.
type pathNode struct {
	cell  Cell
	index int
	f, h  float64
}

// pathQueue is a min-heap of pathNode ordered by f, then h, then cell index,
// a total order that keeps A* deterministic.
type pathQueue []pathNode

func (q pathQueue) Len() int { return len(q) }

func (q pathQueue) Less(i, j int) bool {
	if q[i].f != q[j].f {
		return q[i].f < q[j].f
	}
	if q[i].h != q[j].h {
		return q[i].h < q[j].h
	}
	return q[i].index < q[j].index
}

func (q pathQueue) Swap(i, j int) { q[i], q[j] = q[j], q[i] }

func (q *pathQueue) Push(x any) { *q = append(*q, x.(pathNode)) }

func (q *pathQueue) Pop() any {
	old := *q
	n := old[len(old)-1]
	*q = old[:len(old)-1]
	return n
}

// NearestReachable returns the cell reachable from from (by the same moves as
// FindPath, from itself included) whose centre is closest to target's centre
// in a straight line. Ties go to the lowest row, then the lowest column. It
// returns false only when from is solid or outside the grid. target may be any
// cell, solid or outside the grid; an open target reachable from from is
// returned itself.
func (g *Grid) NearestReachable(from, target Cell) (Cell, bool) {
	if !g.open(from) {
		return Cell{}, false
	}

	// Flood fill every cell reachable from from.
	reached := make([]bool, g.Cols*g.Rows)
	reached[from.Row*g.Cols+from.Col] = true
	queue := []Cell{from}
	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		for _, d := range neighbours {
			if !g.canStep(cur, d) {
				continue
			}
			next := Cell{cur.Col + d.Col, cur.Row + d.Row}
			if i := next.Row*g.Cols + next.Col; !reached[i] {
				reached[i] = true
				queue = append(queue, next)
			}
		}
	}

	// All cells are the same size, so comparing squared cell offsets compares
	// centre distances exactly. Scanning in row-major order and keeping only a
	// strictly closer cell resolves ties to the lowest row, then column.
	best, bestDist := from, -1
	for i, ok := range reached {
		if !ok {
			continue
		}
		c := Cell{Col: i % g.Cols, Row: i / g.Cols}
		dc, dr := c.Col-target.Col, c.Row-target.Row
		if dist := dc*dc + dr*dr; bestDist < 0 || dist < bestDist {
			best, bestDist = c, dist
		}
	}
	return best, true
}
