package world

import (
	"fmt"

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
)

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

// loadMarkers reads the player_spawn and warp objects of the objects layer
// into m. It returns an error naming the object id for a spawn without a name
// or with a name already used, and for a warp without an area or without a
// target map or spawn.
func (m *Map) loadMarkers() error {
	layer := m.ObjectLayer(objectsLayer)
	if layer == nil {
		return nil
	}
	seen := make(map[string]int)
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
		}
	}
	return nil
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
