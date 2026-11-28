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

// openArea is a 4.5 by 4.5 tile wander area with no walls around it, centred
// on openCentre, where a villager starts.
var (
	openArea   = geom.Rect{Min: geom.Vec{X: 64, Y: 64}, Max: geom.Vec{X: 136, Y: 136}}
	openCentre = geom.Vec{X: 100, Y: 100}
)

func TestNPC_StopsAndFacesANearbyCritterOnEachSide(t *testing.T) {
	tests := []struct {
		name    string
		critter geom.Vec
		want    entity.Facing
	}{
		{"critter above", geom.Vec{X: 100, Y: 70}, entity.FacingUp},
		{"critter below", geom.Vec{X: 100, Y: 130}, entity.FacingDown},
		{"critter to the left", geom.Vec{X: 70, Y: 100}, entity.FacingLeft},
		{"critter to the right", geom.Vec{X: 130, Y: 100}, entity.FacingRight},
		{"critter diagonally left and up, more up", geom.Vec{X: 90, Y: 70}, entity.FacingUp},
		{"critter exactly diagonal faces sideways", geom.Vec{X: 120, Y: 80}, entity.FacingRight},
		{"critter at the edge of the range", geom.Vec{X: 100, Y: 60}, entity.FacingUp},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			n := entity.NewNPC("grandma_hedgehog", openCentre, openArea, rand.New(rand.NewPCG(1, 2)))
			// Left alone a villager walks off within its 1 to 3 s rest; ten
			// seconds near the critter must not move it at all.
			for range 10 * 60 {
				n.Update(tick, tc.critter)
			}
			if n.Pos != openCentre {
				t.Errorf("villager moved to %v with the critter at %v; want it to stand at %v", n.Pos, tc.critter, openCentre)
			}
			if n.Facing != tc.want {
				t.Errorf("Facing = %v with the critter at %v; want %v", n.Facing, tc.critter, tc.want)
			}
		})
	}
}

func TestNPC_IgnoresAFarCritterAndWandersAgainOnceTheCritterLeaves(t *testing.T) {
	n := entity.NewNPC("baker_badger", openCentre, openArea, rand.New(rand.NewPCG(1, 2)))

	// 41 px below is just out of range: the villager sets off as if alone.
	// Stop at its first step, so the critter below arrives mid-walk.
	far := geom.Vec{X: 100, Y: 141}
	walking := false
	for range 5 * 60 {
		before := n.Pos
		n.Update(tick, far)
		if n.Pos != before {
			walking = true
			break
		}
	}
	if !walking {
		t.Fatalf("villager stood at %v for 5 s with the critter 41 px away; want it to wander", n.Pos)
	}

	// The critter comes up beside the walking villager, which stops at once.
	if movesWithin(n, 5, n.Pos.Add(geom.Vec{X: 20})) {
		t.Fatalf("villager moved to %v with the critter next to it; want it to stand still", n.Pos)
	}

	// The critter walks away: after a rest the villager wanders on.
	if !movesWithin(n, 5, geom.Vec{X: 1000, Y: 1000}) {
		t.Errorf("villager stood at %v for 5 s after the critter left; want it to wander again", n.Pos)
	}
}

// movesWithin updates n for the given seconds with the critter at critter and
// reports whether the villager's feet moved.
func movesWithin(n *entity.NPC, seconds int, critter geom.Vec) bool {
	start := n.Pos
	for range seconds * 60 {
		n.Update(tick, critter)
	}
	return n.Pos != start
}
