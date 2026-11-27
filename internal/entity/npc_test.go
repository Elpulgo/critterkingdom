package entity_test

import (
	"math"
	"math/rand/v2"
	"testing"

	"github.com/Elpulgo/critterkingdom/internal/entity"
	"github.com/Elpulgo/critterkingdom/internal/geom"
	"github.com/Elpulgo/critterkingdom/internal/world"
)

// pondTestTMJ is a 5x5 map of 16 px world tiles whose middle cell (2,2),
// covering x=32..48, y=32..48, is solid and every other cell is open.
const pondTestTMJ = `{
	"orientation": "orthogonal",
	"infinite": false,
	"width": 5,
	"height": 5,
	"tilewidth": 128,
	"tileheight": 128,
	"tilesets": [
		{"firstgid": 1, "name": "solid", "tilecount": 1}
	],
	"layers": [
		{
			"name": "collision",
			"type": "tilelayer",
			"width": 5,
			"height": 5,
			"data": [
				0, 0, 0, 0, 0,
				0, 0, 0, 0, 0,
				0, 0, 1, 0, 0,
				0, 0, 0, 0, 0,
				0, 0, 0, 0, 0
			]
		}
	]
}`

// wanderSeconds is how long the wander tests run the villager.
const wanderSeconds = 60

func TestNPC_WanderingStaysInsideItsAreaAndNeverWalksIntoSolidTiles(t *testing.T) {
	m, err := world.LoadMap([]byte(pondTestTMJ))
	if err != nil {
		t.Fatalf("LoadMap: %v", err)
	}
	// The 3x3 tile area x=16..64, y=16..64 has the solid cell in its middle.
	area := geom.Rect{Min: geom.Vec{X: 16, Y: 16}, Max: geom.Vec{X: 64, Y: 64}}
	start := geom.Vec{X: 24, Y: 28} // box x=18..30, y=18..28: clear of the solid cell
	n := entity.NewNPC("grandma_hedgehog", start, area, rand.New(rand.NewPCG(1, 2)))
	n.Walls = m.Collision
	// The critter stays far outside the map so it plays no part.
	far := geom.Vec{X: 1000, Y: 1000}

	moved := 0.0
	for i := range wanderSeconds * 60 {
		before := n.Pos
		n.Update(tick, far)
		moved += n.Pos.Sub(before).Len()

		if !area.Contains(n.Pos) {
			t.Fatalf("tick %d: feet at %v left the wander area %v", i, n.Pos, area)
		}
		if col, row, ok := solidUnder(m.Collision, n.Box()); ok {
			t.Fatalf("tick %d: box %v overlaps solid cell (%d,%d)", i, n.Box(), col, row)
		}
	}
	// At 24 px/s with 1 to 3 s rests a minute of wandering covers far more
	// than one tile; a villager stuck in place would cover nothing.
	if moved < 16 {
		t.Errorf("walked %.1f px in %d s; want at least a tile's 16 px", moved, wanderSeconds)
	}
}

func TestNPC_WithNoWanderAreaStaysWhereItStands(t *testing.T) {
	pos := geom.Vec{X: 40, Y: 40}
	n := entity.NewNPC("baker_badger", pos, geom.Rect{Min: pos, Max: pos}, rand.New(rand.NewPCG(1, 2)))

	for range wanderSeconds * 60 {
		n.Update(tick, geom.Vec{X: 1000, Y: 1000})
	}
	if n.Pos != pos {
		t.Errorf("after %d s at %v; want still at %v", wanderSeconds, n.Pos, pos)
	}
}

// solidUnder returns a solid cell of g that box overlaps with positive area.
func solidUnder(g *world.Grid, box geom.Rect) (col, row int, ok bool) {
	for row := int(math.Floor(box.Min.Y / g.TileSize)); float64(row)*g.TileSize < box.Max.Y; row++ {
		for col := int(math.Floor(box.Min.X / g.TileSize)); float64(col)*g.TileSize < box.Max.X; col++ {
			if g.Solid(col, row) {
				return col, row, true
			}
		}
	}
	return 0, 0, false
}
