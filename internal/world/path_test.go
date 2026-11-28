package world

import (
	"math"
	"slices"
	"testing"
)

// gridOf builds a Grid of 16 px cells from rows of text, '#' solid and any
// other character open.
func gridOf(t *testing.T, rows ...string) *Grid {
	t.Helper()
	g := &Grid{Cols: len(rows[0]), Rows: len(rows), TileSize: 16}
	g.solid = make([]bool, g.Cols*g.Rows)
	for r, line := range rows {
		if len(line) != g.Cols {
			t.Fatalf("row %d has %d cells; want %d", r, len(line), g.Cols)
		}
		for c, ch := range line {
			g.solid[r*g.Cols+c] = ch == '#'
		}
	}
	return g
}

// checkWalkable fails the test unless path starts at from, ends at to, stays
// on open in-grid cells, moves to one of the eight neighbours each step and
// never steps diagonally past a solid or off-grid orthogonal neighbour. It
// returns the path's cost: 1 per straight step and √2 per diagonal step.
func checkWalkable(t *testing.T, g *Grid, path []Cell, from, to Cell) float64 {
	t.Helper()
	if len(path) == 0 || path[0] != from || path[len(path)-1] != to {
		t.Fatalf("path %v does not run from %v to %v", path, from, to)
	}
	standable := func(c Cell) bool {
		return c.Col >= 0 && c.Col < g.Cols && c.Row >= 0 && c.Row < g.Rows && !g.Solid(c.Col, c.Row)
	}
	cost := 0.0
	for i, c := range path {
		if !standable(c) {
			t.Fatalf("path %v: cell %v is solid or outside the grid", path, c)
		}
		if i == 0 {
			continue
		}
		p := path[i-1]
		dc, dr := c.Col-p.Col, c.Row-p.Row
		if dc < -1 || dc > 1 || dr < -1 || dr > 1 || (dc == 0 && dr == 0) {
			t.Fatalf("path %v: step %v -> %v is not to a neighbour", path, p, c)
		}
		if dc != 0 && dr != 0 {
			if !standable(Cell{p.Col + dc, p.Row}) || !standable(Cell{p.Col, p.Row + dr}) {
				t.Fatalf("path %v: diagonal step %v -> %v cuts a solid corner", path, p, c)
			}
			cost += math.Sqrt2
		} else {
			cost++
		}
	}
	return cost
}

func TestGrid_FindPath_GoesThroughTheOnlyGapInAWallByTheShortestWay(t *testing.T) {
	g := gridOf(t,
		".......",
		".......",
		"###.###",
		".......",
		".......",
	)
	from, to := Cell{0, 1}, Cell{6, 3}

	path, ok := g.FindPath(from, to)
	if !ok {
		t.Fatalf("FindPath(%v, %v) found no path", from, to)
	}
	checkWalkable(t, g, path, from, to)

	// The gap (3,2) can only be entered from (3,1) and left to (3,3), since
	// the diagonals into it cut the wall's corners. Straight along row 1, down
	// through the gap and straight along row 3 costs 8; any other route
	// swaps straight steps for dearer diagonals.
	want := []Cell{{0, 1}, {1, 1}, {2, 1}, {3, 1}, {3, 2}, {3, 3}, {4, 3}, {5, 3}, {6, 3}}
	if !slices.Equal(path, want) {
		t.Errorf("FindPath(%v, %v) = %v; want %v", from, to, path, want)
	}
}

func TestGrid_FindPath_DoesNotSqueezeDiagonallyBetweenTwoSolidCorners(t *testing.T) {
	g := gridOf(t,
		"....",
		"..#.",
		".#..",
		"....",
	)
	from, to := Cell{1, 1}, Cell{2, 2}

	path, ok := g.FindPath(from, to)
	if !ok {
		t.Fatalf("FindPath(%v, %v) found no path", from, to)
	}
	cost := checkWalkable(t, g, path, from, to)

	// Going round either solid cell takes six straight steps; every diagonal
	// that would shorten it cuts one of the two solid corners.
	if math.Abs(cost-6) > 1e-9 || len(path) != 7 {
		t.Errorf("path %v costs %.4f over %d cells; want 6 over 7", path, cost, len(path))
	}
}

func TestGrid_FindPath_ReportsNoPathToOrFromAnUnwalkableCell(t *testing.T) {
	g := gridOf(t,
		"..#..",
		"..#..",
		"..###",
		"...#.",
	)
	// (4,3) is open but enclosed by (3,2), (4,2) and (3,3); (4,0) is open but
	// the wall in column 2 and row 2 shuts its corner off from the left side.

	tests := []struct {
		name     string
		from, to Cell
	}{
		{"goal enclosed by solid cells", Cell{0, 0}, Cell{4, 3}},
		{"start enclosed by solid cells", Cell{4, 3}, Cell{0, 0}},
		{"goal behind a closed wall", Cell{0, 0}, Cell{4, 0}},
		{"goal solid", Cell{0, 0}, Cell{2, 0}},
		{"start solid", Cell{2, 1}, Cell{0, 0}},
		{"goal left of the grid", Cell{0, 0}, Cell{-1, 0}},
		{"goal below the grid", Cell{0, 0}, Cell{0, 4}},
		{"start right of the grid", Cell{5, 0}, Cell{0, 0}},
		{"start above the grid", Cell{1, -1}, Cell{0, 0}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if path, ok := g.FindPath(tc.from, tc.to); ok || path != nil {
				t.Errorf("FindPath(%v, %v) = %v, %v; want nil, false", tc.from, tc.to, path, ok)
			}
		})
	}
}

func TestGrid_FindPath_StartingOnTheGoalIsAOneCellPath(t *testing.T) {
	g := gridOf(t, "..", "..")

	path, ok := g.FindPath(Cell{1, 0}, Cell{1, 0})
	if !ok || len(path) != 1 || path[0] != (Cell{1, 0}) {
		t.Errorf("FindPath((1,0), (1,0)) = %v, %v; want [(1,0)], true", path, ok)
	}
}
