package world

import (
	"fmt"
	"math"
	"time"

	"github.com/Elpulgo/critterkingdom/internal/geom"
)

const (
	// objectsLayer is the object layer that holds spawns, warps and other
	// gameplay markers.
	objectsLayer = "objects"
	// spawnType is the Tiled object type of a named arrival point.
	spawnType = "player_spawn"
	// warpType is the Tiled object type of an area that moves the critter to
	// another map.
	warpType = "warp"
	// spawnZoneType is the Tiled object type of an area where wild critters
	// of one habitat appear.
	spawnZoneType = "spawn_zone"

	// defaultZoneMax is how many critters a spawn_zone holds when it has no
	// max property.
	defaultZoneMax = 3
	// defaultZoneRespawn is how long a freed slot of a spawn_zone waits when
	// the zone has no respawn_s property.
	defaultZoneRespawn = 60 * time.Second
)

// SpawnZone is an area of a map where wild critters of one habitat appear.
type SpawnZone struct {
	// ID names the zone: the object's name, or "zone-<object id>" when the
	// object has none. It is unique within the map.
	ID string
	// Rect is the area critters appear and move in, in world pixels.
	Rect geom.Rect
	// Habitat is the habitat of the species that appear here.
	Habitat string
	// Max is how many critters the zone holds at once.
	Max int
	// RespawnAfter is how long a freed slot waits before it is filled again.
	RespawnAfter time.Duration
}

// Warp is an area of a map that moves the critter to the spawn TargetSpawn of
// the map TargetMap when it walks into it.
type Warp struct {
	// Rect is the area that triggers the warp, in world pixels.
	Rect        geom.Rect
	TargetMap   string
	TargetSpawn string
}

// contains reports whether pos lies inside the warp's area. The min edges are
// inside, the max edges are not, so two warps that share an edge never both
// contain the same point.
func (w Warp) contains(pos geom.Vec) bool {
	return pos.X >= w.Rect.Min.X && pos.X < w.Rect.Max.X &&
		pos.Y >= w.Rect.Min.Y && pos.Y < w.Rect.Max.Y
}

// spawn is a named player_spawn object.
type spawn struct {
	name string
	pos  geom.Vec
}

// Spawn returns the world-pixel position of the player_spawn named name in
// the objects layer, and false when the map has no spawn with that name.
func (m *Map) Spawn(name string) (geom.Vec, bool) {
	for _, s := range m.spawns {
		if s.name == name {
			return s.pos, true
		}
	}
	return geom.Vec{}, false
}

// DefaultSpawn returns the world-pixel position of the first player_spawn in
// the objects layer, and false when the map has none.
func (m *Map) DefaultSpawn() (geom.Vec, bool) {
	if len(m.spawns) == 0 {
		return geom.Vec{}, false
	}
	return m.spawns[0].pos, true
}

// Warps returns every warp object of the objects layer, in layer order. The
// slice is a copy; changing it does not change the map.
func (m *Map) Warps() []Warp {
	return append([]Warp(nil), m.warps...)
}

// SpawnZones returns every spawn_zone object of the objects layer, in layer
// order. The slice is a copy; changing it does not change the map.
func (m *Map) SpawnZones() []SpawnZone {
	return append([]SpawnZone(nil), m.zones...)
}

// WarpAt returns the first warp whose area contains pos, and false when pos
// is in no warp.
func (m *Map) WarpAt(pos geom.Vec) (Warp, bool) {
	for _, w := range m.warps {
		if w.contains(pos) {
			return w, true
		}
	}
	return Warp{}, false
}

// loadMarkers reads the player_spawn, warp and spawn_zone objects of the
// objects layer into m. It returns an error naming the object id for a spawn
// without a name or with a name already used, for a warp without an area or
// without a target map or spawn, and for a spawn_zone that parseSpawnZone
// rejects or whose ID another zone already uses.
func (m *Map) loadMarkers() error {
	layer := m.ObjectLayer(objectsLayer)
	if layer == nil {
		return nil
	}
	seen := make(map[string]int)
	seenZones := make(map[string]int)
	seenNPC := make(map[string]int)
	for _, o := range layer.Objects {
		switch o.Type {
		case spawnType:
			if o.Name == "" {
				return fmt.Errorf("%s object %d: player_spawn has no name", objectsLayer, o.ID)
			}
			if first, dup := seen[o.Name]; dup {
				return fmt.Errorf("%s object %d: player_spawn name %q is already used by object %d", objectsLayer, o.ID, o.Name, first)
			}
			seen[o.Name] = o.ID
			m.spawns = append(m.spawns, spawn{name: o.Name, pos: geom.Vec{X: o.X, Y: o.Y}})

		case warpType:
			if o.W <= 0 || o.H <= 0 {
				return fmt.Errorf("%s object %d: warp has size %vx%v; want a rectangle with positive width and height", objectsLayer, o.ID, o.W, o.H)
			}
			w := Warp{
				Rect:        o.Rect(),
				TargetMap:   o.Properties.String("target_map"),
				TargetSpawn: o.Properties.String("target_spawn"),
			}
			if w.TargetMap == "" {
				return fmt.Errorf("%s object %d: warp has no string property target_map", objectsLayer, o.ID)
			}
			if w.TargetSpawn == "" {
				return fmt.Errorf("%s object %d: warp has no string property target_spawn", objectsLayer, o.ID)
			}
			m.warps = append(m.warps, w)

		case spawnZoneType:
			z, err := parseSpawnZone(o)
			if err != nil {
				return err
			}
			if first, dup := seenZones[z.ID]; dup {
				return fmt.Errorf("%s object %d: spawn_zone id %q is already used by object %d", objectsLayer, o.ID, z.ID, first)
			}
			seenZones[z.ID] = o.ID
			m.zones = append(m.zones, z)
      
		case npcType:
			if err := m.loadNPC(o, seenNPC); err != nil {
				return err
			}
		}
	}
	return nil
}

// parseSpawnZone turns the spawn_zone object o into a SpawnZone. It returns an
// error naming the object id when o has no area, no string property habitat,
// a max that is not a whole number of at least 1, or a respawn_s that is not
// a number of at least 0. An absent max or respawn_s takes its default.
func parseSpawnZone(o Object) (SpawnZone, error) {
	if o.W <= 0 || o.H <= 0 {
		return SpawnZone{}, fmt.Errorf("%s object %d: spawn_zone has size %vx%v; want a rectangle with positive width and height", objectsLayer, o.ID, o.W, o.H)
	}
	z := SpawnZone{
		ID:           o.Name,
		Rect:         o.Rect(),
		Habitat:      o.Properties.String("habitat"),
		Max:          defaultZoneMax,
		RespawnAfter: defaultZoneRespawn,
	}
	if z.ID == "" {
		z.ID = fmt.Sprintf("zone-%d", o.ID)
	}
	if z.Habitat == "" {
		return SpawnZone{}, fmt.Errorf("%s object %d: spawn_zone has no string property habitat", objectsLayer, o.ID)
	}
	if raw, ok := o.Properties["max"]; ok {
		v, isNum := raw.(float64)
		if !isNum || v < 1 || v != math.Trunc(v) {
			return SpawnZone{}, fmt.Errorf("%s object %d: spawn_zone property max is %v; want a whole number of at least 1", objectsLayer, o.ID, raw)
		}
		z.Max = int(v)
	}
	if raw, ok := o.Properties["respawn_s"]; ok {
		v, isNum := raw.(float64)
		if !isNum || v < 0 {
			return SpawnZone{}, fmt.Errorf("%s object %d: spawn_zone property respawn_s is %v; want a number of seconds of at least 0", objectsLayer, o.ID, raw)
		}
		z.RespawnAfter = time.Duration(v * float64(time.Second))
	}
	return z, nil
}

// MapLoader loads the map called name, for example "meadow". The game passes
// assets.LoadMap; world cannot import assets itself.
type MapLoader func(name string) (*Map, error)

// ResolveWarp loads the target map of w with load and returns it together with
// the world-pixel position of its spawn w.TargetSpawn. It returns an error
// naming the target map when the map cannot be loaded, and naming both the map
// and the spawn when the map has no player_spawn of that name.
func ResolveWarp(load MapLoader, w Warp) (*Map, geom.Vec, error) {
	m, err := load(w.TargetMap)
	if err != nil {
		return nil, geom.Vec{}, fmt.Errorf("warp to map %q: %w", w.TargetMap, err)
	}
	pos, ok := m.Spawn(w.TargetSpawn)
	if !ok {
		return nil, geom.Vec{}, fmt.Errorf("warp to map %q: no player_spawn named %q", w.TargetMap, w.TargetSpawn)
	}
	return m, pos, nil
}
