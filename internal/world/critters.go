package world

import (
	"math/rand/v2"
	"time"

	"github.com/Elpulgo/critterkingdom/internal/critter"
	"github.com/Elpulgo/critterkingdom/internal/entity"
	"github.com/Elpulgo/critterkingdom/internal/geom"
)

// placeTries is how many random points of a zone are tried for a new wild
// critter before giving up until the next tick.
const placeTries = 20

// wildCritter is one wild critter in the scene: the spawner's critter with
// the entity that moves it.
type wildCritter struct {
	*entity.Wild
	id      int // the critter.Spawned ID
	species *critter.Species
	zone    string // the ID of the zone it belongs to
}

// AddWildCritters makes wild critters of species appear in the spawn zones of
// the scene's maps from the next Update on, the current map and every map a
// warp leads to. All their randomness comes from rng. Without this call the
// scene has no wild critters.
func (s *Scene) AddWildCritters(species []critter.Species, rng *rand.Rand) {
	s.species = species
	s.rng = rng
	s.resetWildCritters()
}

// resetWildCritters drops the wild critters and builds a new spawner for the
// current map's zones, if AddWildCritters was called.
func (s *Scene) resetWildCritters() {
	s.wild = nil
	s.spawner = nil
	if s.rng == nil {
		return
	}
	zones := s.m.SpawnZones()
	czones := make([]critter.Zone, len(zones))
	for i, z := range zones {
		czones[i] = critter.Zone{ID: z.ID, Habitat: z.Habitat, Max: z.Max, RespawnAfter: z.RespawnAfter}
	}
	s.spawner = critter.NewSpawner(s.species, czones, s.rng, s.gameTime)
}

// gameTime is the scene's elapsed game time, the spawner's clock.
func (s *Scene) gameTime() time.Duration {
	return time.Duration(s.elapsed * float64(time.Second))
}

// updateWildCritters brings the spawner up to date, gives every critter the
// spawner has present an entity and drops the entities of those it no longer
// has, then moves each critter on by dt.
func (s *Scene) updateWildCritters(dt float64) {
	if s.spawner == nil {
		return
	}
	s.spawner.Update(critter.World{TimeOfDay: s.timeOfDay})

	byID := make(map[int]*wildCritter, len(s.wild))
	for _, w := range s.wild {
		byID[w.id] = w
	}
	var wild []*wildCritter
	for _, z := range s.m.SpawnZones() {
		for _, sp := range s.spawner.Present(z.ID) {
			if w, ok := byID[sp.ID]; ok {
				wild = append(wild, w)
				continue
			}
			pos, ok := s.openPointIn(z.Rect)
			if !ok {
				continue // tried again next tick
			}
			e := entity.NewWild(pos, z.Rect, s.rng)
			e.Walls = s.m.Collision
			e.Rare = sp.Rare
			wild = append(wild, &wildCritter{Wild: e, id: sp.ID, species: sp.Species, zone: z.ID})
		}
	}
	s.wild = wild

	for _, w := range s.wild {
		w.Update(dt, s.player.Pos)
	}
}

// openPointIn draws up to placeTries random points inside r and returns the
// first where a critter's feet box stays inside the map on open cells, and
// false when none of them is.
func (s *Scene) openPointIn(r geom.Rect) (geom.Vec, bool) {
	size := r.Max.Sub(r.Min)
	for range placeTries {
		p := geom.Vec{
			X: r.Min.X + s.rng.Float64()*size.X,
			Y: r.Min.Y + s.rng.Float64()*size.Y,
		}
		if s.walkable.Contains(p) && s.m.Collision.open(critterFeet(p)) {
			return p, true
		}
	}
	return geom.Vec{}, false
}

// critterFeet returns the 12 by 10 feet box of a critter whose feet are at p.
func critterFeet(p geom.Vec) geom.Rect {
	return geom.Rect{
		Min: geom.Vec{X: p.X - feetMarginX, Y: p.Y - feetMarginTop},
		Max: p.Add(geom.Vec{X: feetMarginX}),
	}
}

// open reports whether box overlaps no solid cell with positive area. A nil
// grid is open everywhere.
func (g *Grid) open(box geom.Rect) bool {
	if g == nil || g.TileSize <= 0 {
		return true
	}
	c0, r0, c1, r1 := g.cellRange(box)
	for row := r0; row <= r1; row++ {
		for col := c0; col <= c1; col++ {
			if g.Solid(col, row) {
				return false
			}
		}
	}
	return true
}
