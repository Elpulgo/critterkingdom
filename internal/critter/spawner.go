package critter

import (
	"math/rand/v2"
	"slices"
	"time"
)

// Clock returns the game time elapsed since the adventure began. It stops
// while the game is paused, so respawn timers pause with it.
type Clock func() time.Duration

// Spawn weights per rarity: a species is drawn in proportion to its weight
// among the active species of a habitat.
const (
	CommonWeight   = 6
	UncommonWeight = 3
	RareWeight     = 1
)

// weight returns the spawn weight for the rarity, or 0 for an unknown one.
func (r Rarity) weight() int {
	switch r {
	case Common:
		return CommonWeight
	case Uncommon:
		return UncommonWeight
	case Rare:
		return RareWeight
	}
	return 0
}

// Zone is an area of the map where critters of one habitat appear.
type Zone struct {
	// ID names the zone; it must be unique among the spawner's zones.
	ID string
	// Habitat is the habitat of the species that appear here.
	Habitat string
	// Max is how many critters the zone holds at once.
	Max int
	// RespawnAfter is how long a slot freed by Remove waits before it is
	// filled again.
	RespawnAfter time.Duration
}

// Spawned is one critter present in a zone.
type Spawned struct {
	// ID identifies this critter for Remove; it is unique per spawner.
	ID      int
	Species *Species
	// Rare is true when the species has rarity rare, so the caller can make
	// it sparkle.
	Rare bool
}

// Spawner keeps a few critters present in each zone, drawn by rarity from
// the zone's habitat among the species active in the current World.
type Spawner struct {
	species []Species
	zones   []zoneState
	rng     *rand.Rand
	clock   Clock
	nextID  int
}

type zoneState struct {
	Zone
	present []Spawned
	// respawnAt holds the game time at which each slot freed by Remove may
	// be filled again.
	respawnAt []time.Duration
}

// NewSpawner returns a spawner for the zones that draws from species with
// rng and reads game time from clock. No zone holds critters until the
// first Update.
func NewSpawner(species []Species, zones []Zone, rng *rand.Rand, clock Clock) *Spawner {
	s := &Spawner{species: species, rng: rng, clock: clock}
	for _, z := range zones {
		s.zones = append(s.zones, zoneState{Zone: z})
	}
	return s
}

// Update brings every zone up to date with w: critters whose conditions no
// longer hold leave at once and free their slot, slots whose respawn time
// has come are freed, and every free slot gets an independent weighted draw
// from the zone's active species. A zone with no active species stays short.
func (s *Spawner) Update(w World) {
	now := s.clock()
	for i := range s.zones {
		z := &s.zones[i]
		z.present = slices.DeleteFunc(z.present, func(c Spawned) bool {
			return !c.Species.Active(w)
		})
		z.respawnAt = slices.DeleteFunc(z.respawnAt, func(at time.Duration) bool {
			return at <= now
		})
		s.fill(z, w)
	}
}

func (s *Spawner) fill(z *zoneState, w World) {
	free := z.Max - len(z.present) - len(z.respawnAt)
	if free <= 0 {
		return
	}
	var candidates []*Species
	total := 0
	for i := range s.species {
		sp := &s.species[i]
		if sp.Habitat == z.Habitat && sp.Active(w) && sp.Rarity.weight() > 0 {
			candidates = append(candidates, sp)
			total += sp.Rarity.weight()
		}
	}
	if total == 0 {
		return
	}
	for range free {
		sp := draw(s.rng, candidates, total)
		s.nextID++
		z.present = append(z.present, Spawned{ID: s.nextID, Species: sp, Rare: sp.Rarity == Rare})
	}
}

// draw picks one of candidates with probability weight/total.
func draw(rng *rand.Rand, candidates []*Species, total int) *Species {
	r := rng.IntN(total)
	for _, sp := range candidates {
		r -= sp.Rarity.weight()
		if r < 0 {
			return sp
		}
	}
	return candidates[len(candidates)-1]
}

// Present returns the critters in the zone in the order they appeared, or
// nil for an unknown zone. The slice is a copy the caller may keep.
func (s *Spawner) Present(zoneID string) []Spawned {
	for i := range s.zones {
		if s.zones[i].ID == zoneID {
			return slices.Clone(s.zones[i].present)
		}
	}
	return nil
}

// Remove takes the critter with the given id away, for example when it hops
// off after Snack Time. Its slot stays empty until the zone's RespawnAfter
// of game time has passed. An unknown id is ignored.
func (s *Spawner) Remove(id int) {
	for i := range s.zones {
		z := &s.zones[i]
		j := slices.IndexFunc(z.present, func(c Spawned) bool { return c.ID == id })
		if j < 0 {
			continue
		}
		z.present = slices.Delete(z.present, j, j+1)
		z.respawnAt = append(z.respawnAt, s.clock()+z.RespawnAfter)
		return
	}
}
