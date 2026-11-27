package world

import (
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"time"

	"github.com/Elpulgo/critterkingdom/internal/geom"
)

const (
	// tiledTileSize is the size of a tile in Tiled pixels that every map is
	// authored with.
	tiledTileSize = 128
	// tiledScale converts Tiled pixels (authored at 8 px per world pixel)
	// into world pixels.
	tiledScale = 8
	// gidFlipMask covers the top four bits of a Tiled global tile ID:
	// horizontal flip, vertical flip, diagonal flip and hexagonal 120° rotate.
	gidFlipMask uint32 = 0xF0000000
)

// Map is a Tiled map loaded into world pixels.
type Map struct {
	Cols       int
	Rows       int
	TileSize   float64
	Properties Properties
	Tilesets   []Tileset
	Collision  *Grid

	tileLayers   map[string]*TileLayer
	objectLayers map[string]*ObjectLayer
}

// Size returns the map's dimensions in world pixels.
func (m *Map) Size() geom.Vec {
	return geom.Vec{
		X: float64(m.Cols) * m.TileSize,
		Y: float64(m.Rows) * m.TileSize,
	}
}

// TileLayer returns the tile layer with the given name, or nil when the map
// does not have one.
func (m *Map) TileLayer(name string) *TileLayer {
	return m.tileLayers[name]
}

// ObjectLayer returns the object layer with the given name, or nil when the
// map does not have one.
func (m *Map) ObjectLayer(name string) *ObjectLayer {
	return m.objectLayers[name]
}

// Tile looks up the tileset tile for gid. Flip bits on gid are ignored. It
// returns false when gid is 0 or no tileset in the map covers it.
func (m *Map) Tile(gid uint32) (Tile, bool) {
	ts, local, ok := m.findTileset(gid)
	if !ok {
		return Tile{}, false
	}
	if t, exists := ts.tiles[local]; exists {
		return t, true
	}
	return Tile{}, true
}

// FrameGID returns the global tile ID of the animation frame showing for gid
// at elapsed. It loops after the last frame and returns gid (with flip bits
// cleared) when the tile has no animation or gid is not in any tileset.
func (m *Map) FrameGID(gid uint32, elapsed time.Duration) uint32 {
	gid &^= gidFlipMask
	ts, local, ok := m.findTileset(gid)
	if !ok {
		return gid
	}
	t, exists := ts.tiles[local]
	if !exists || len(t.Animation) == 0 {
		return gid
	}
	var total time.Duration
	for _, f := range t.Animation {
		if f.Duration > 0 {
			total += f.Duration
		}
	}
	if total <= 0 {
		return ts.FirstGID + uint32(t.Animation[0].TileID)
	}
	rem := elapsed % total
	if rem < 0 {
		rem += total
	}
	var acc time.Duration
	for _, f := range t.Animation {
		if f.Duration <= 0 {
			continue
		}
		acc += f.Duration
		if rem < acc {
			return ts.FirstGID + uint32(f.TileID)
		}
	}
	return ts.FirstGID + uint32(t.Animation[len(t.Animation)-1].TileID)
}

// VisibleCells returns the half-open cell range [c0, c1) and [r0, r1) of map
// cells that overlap view with positive area, clamped to the map's dimensions.
// When view does not overlap the map with positive area, it returns an empty
// range (c0 == c1 or r0 == r1).
func (m *Map) VisibleCells(view geom.Rect) (c0, r0, c1, r1 int) {
	if m == nil || m.TileSize <= 0 || m.Cols <= 0 || m.Rows <= 0 {
		return 0, 0, 0, 0
	}
	if view.Max.X <= view.Min.X || view.Max.Y <= view.Min.Y {
		return 0, 0, 0, 0
	}
	c0 = clampCell(int(math.Floor(view.Min.X/m.TileSize)), 0, m.Cols)
	c1 = clampCell(int(math.Ceil(view.Max.X/m.TileSize)), 0, m.Cols)
	r0 = clampCell(int(math.Floor(view.Min.Y/m.TileSize)), 0, m.Rows)
	r1 = clampCell(int(math.Ceil(view.Max.Y/m.TileSize)), 0, m.Rows)
	if c1 <= c0 || r1 <= r0 {
		return 0, 0, 0, 0
	}
	return c0, r0, c1, r1
}

func clampCell(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

// findTileset returns the tileset that owns gid and its 0-based local tile ID.
func (m *Map) findTileset(gid uint32) (*Tileset, int, bool) {
	gid &^= gidFlipMask
	if gid == 0 {
		return nil, 0, false
	}
	var chosen *Tileset
	for i := range m.Tilesets {
		ts := &m.Tilesets[i]
		if ts.FirstGID <= gid && (chosen == nil || ts.FirstGID > chosen.FirstGID) {
			chosen = ts
		}
	}
	if chosen == nil {
		return nil, 0, false
	}
	local := int(gid - chosen.FirstGID)
	if local >= chosen.TileCount {
		return nil, 0, false
	}
	return chosen, local, true
}

// TileLayer is a grid of global tile IDs, stored row by row from the top-left.
// A GID of 0 means the cell is empty.
type TileLayer struct {
	Name string
	Cols int
	Rows int
	GIDs []uint32
}

// At returns the global tile ID at col, row, or 0 when col, row is outside the
// layer.
func (l *TileLayer) At(col, row int) uint32 {
	if col < 0 || col >= l.Cols || row < 0 || row >= l.Rows {
		return 0
	}
	return l.GIDs[row*l.Cols+col]
}

// ObjectLayer is a named list of map objects.
type ObjectLayer struct {
	Name    string
	Objects []Object
}

// Object is a rectangle, point or tile object placed on an object layer, with
// its coordinates and size scaled to world pixels.
type Object struct {
	ID         int
	Name       string
	Type       string
	X, Y, W, H float64
	GID        uint32
	Point      bool
	Properties Properties
}

// Rect returns the world-pixel rectangle the object covers. In Tiled, a tile
// object (GID != 0) has its origin at the bottom-left corner, whereas a plain
// rectangle has its origin at the top-left corner.
func (o Object) Rect() geom.Rect {
	if o.GID != 0 {
		return geom.Rect{
			Min: geom.Vec{X: o.X, Y: o.Y - o.H},
			Max: geom.Vec{X: o.X + o.W, Y: o.Y},
		}
	}
	return geom.Rect{
		Min: geom.Vec{X: o.X, Y: o.Y},
		Max: geom.Vec{X: o.X + o.W, Y: o.Y + o.H},
	}
}

// Tileset is an embedded tileset.
type Tileset struct {
	FirstGID  uint32
	Name      string
	TileCount int
	tiles     map[int]Tile
}

// Tile holds the per-tile metadata from a tileset.
type Tile struct {
	Type       string
	Properties Properties
	Animation  []Frame
}

// Frame is one step of a tile animation.
type Frame struct {
	TileID   int
	Duration time.Duration
}

// Properties holds custom Tiled properties decoded from JSON (`string`,
// `float64`, `bool`).
type Properties map[string]any

// String returns the string property name, or "" when absent or not a string.
func (p Properties) String(name string) string {
	v, _ := p[name].(string)
	return v
}

// Float returns the numeric property name, or 0 when absent or not a number.
func (p Properties) Float(name string) float64 {
	v, _ := p[name].(float64)
	return v
}

// Bool returns the boolean property name, or false when absent or not a bool.
func (p Properties) Bool(name string) bool {
	v, _ := p[name].(bool)
	return v
}

type tmjMap struct {
	Orientation string        `json:"orientation"`
	Infinite    bool          `json:"infinite"`
	Width       int           `json:"width"`
	Height      int           `json:"height"`
	TileWidth   int           `json:"tilewidth"`
	TileHeight  int           `json:"tileheight"`
	Layers      []tmjLayer    `json:"layers"`
	Tilesets    []tmjTileset  `json:"tilesets"`
	Properties  []tmjProperty `json:"properties"`
}

type tmjLayer struct {
	Name        string          `json:"name"`
	Type        string          `json:"type"`
	Width       int             `json:"width"`
	Height      int             `json:"height"`
	Data        json.RawMessage `json:"data"`
	Encoding    string          `json:"encoding"`
	Compression string          `json:"compression"`
	Objects     []tmjObject     `json:"objects"`
	Properties  []tmjProperty   `json:"properties"`
}

type tmjObject struct {
	ID         int           `json:"id"`
	Name       string        `json:"name"`
	Type       string        `json:"type"`
	Class      string        `json:"class"`
	X          float64       `json:"x"`
	Y          float64       `json:"y"`
	Width      float64       `json:"width"`
	Height     float64       `json:"height"`
	GID        uint32        `json:"gid"`
	Point      bool          `json:"point"`
	Properties []tmjProperty `json:"properties"`
}

type tmjTileset struct {
	FirstGID   uint32        `json:"firstgid"`
	Name       string        `json:"name"`
	Source     string        `json:"source"`
	TileCount  int           `json:"tilecount"`
	Tiles      []tmjTile     `json:"tiles"`
	Properties []tmjProperty `json:"properties"`
}

type tmjTile struct {
	ID         int           `json:"id"`
	Type       string        `json:"type"`
	Class      string        `json:"class"`
	Properties []tmjProperty `json:"properties"`
	Animation  []tmjFrame    `json:"animation"`
}

type tmjFrame struct {
	TileID   int `json:"tileid"`
	Duration int `json:"duration"`
}

type tmjProperty struct {
	Name  string `json:"name"`
	Type  string `json:"type"`
	Value any    `json:"value"`
}

// LoadMap parses a Tiled JSON (.tmj) map and converts all pixel positions and
// sizes into world pixels.
func LoadMap(data []byte) (*Map, error) {
	var raw tmjMap
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("parse map json: %w", err)
	}
	if raw.Orientation != "orthogonal" {
		return nil, fmt.Errorf("unsupported map orientation %q: want orthogonal", raw.Orientation)
	}
	if raw.Infinite {
		return nil, fmt.Errorf("infinite maps are not supported")
	}
	if raw.TileWidth != tiledTileSize || raw.TileHeight != tiledTileSize {
		return nil, fmt.Errorf("unsupported tile size %dx%d: want %dx%d", raw.TileWidth, raw.TileHeight, tiledTileSize, tiledTileSize)
	}
	if raw.Width <= 0 || raw.Height <= 0 {
		return nil, fmt.Errorf("invalid map dimensions %dx%d", raw.Width, raw.Height)
	}

	m := &Map{
		Cols:         raw.Width,
		Rows:         raw.Height,
		TileSize:     tiledTileSize / tiledScale,
		Properties:   decodeProperties(raw.Properties),
		tileLayers:   make(map[string]*TileLayer),
		objectLayers: make(map[string]*ObjectLayer),
	}

	for _, rts := range raw.Tilesets {
		if rts.Source != "" {
			return nil, fmt.Errorf("tileset %q is external (%s): turn on Embed Tilesets when exporting", rts.Name, rts.Source)
		}
		if rts.FirstGID == 0 || rts.TileCount <= 0 {
			return nil, fmt.Errorf("tileset %q has invalid firstgid %d or tilecount %d", rts.Name, rts.FirstGID, rts.TileCount)
		}
		ts := Tileset{
			FirstGID:  rts.FirstGID,
			Name:      rts.Name,
			TileCount: rts.TileCount,
			tiles:     make(map[int]Tile, len(rts.Tiles)),
		}
		for _, rt := range rts.Tiles {
			var anim []Frame
			if len(rt.Animation) > 0 {
				anim = make([]Frame, len(rt.Animation))
				for i, f := range rt.Animation {
					anim[i] = Frame{
						TileID:   f.TileID,
						Duration: time.Duration(f.Duration) * time.Millisecond,
					}
				}
			}
			ts.tiles[rt.ID] = Tile{
				Type:       firstNonEmpty(rt.Type, rt.Class),
				Properties: decodeProperties(rt.Properties),
				Animation:  anim,
			}
		}
		m.Tilesets = append(m.Tilesets, ts)
	}
	sort.Slice(m.Tilesets, func(i, j int) bool {
		return m.Tilesets[i].FirstGID < m.Tilesets[j].FirstGID
	})

	seenLayers := make(map[string]bool, len(raw.Layers))
	for _, rl := range raw.Layers {
		if !allowedLayerType(rl.Name, rl.Type) {
			return nil, fmt.Errorf("layer %q of type %q is not allowed", rl.Name, rl.Type)
		}
		if seenLayers[rl.Name] {
			return nil, fmt.Errorf("duplicate layer %q", rl.Name)
		}
		seenLayers[rl.Name] = true

		switch rl.Type {
		case "tilelayer":
			if rl.Encoding != "" || rl.Compression != "" {
				return nil, fmt.Errorf("layer %q uses encoding %q / compression %q: use CSV (plain JSON array)", rl.Name, rl.Encoding, rl.Compression)
			}
			var rawData []uint32
			if err := json.Unmarshal(rl.Data, &rawData); err != nil {
				return nil, fmt.Errorf("parse layer %q data: %w", rl.Name, err)
			}
			w, h := rl.Width, rl.Height
			if w != m.Cols || h != m.Rows || len(rawData) != w*h {
				return nil, fmt.Errorf("layer %q has %d gids for %dx%d (map is %dx%d)", rl.Name, len(rawData), w, h, m.Cols, m.Rows)
			}
			gids := make([]uint32, len(rawData))
			for i, rawGID := range rawData {
				gid := rawGID &^ gidFlipMask
				if gid != 0 {
					if _, _, ok := m.findTileset(gid); !ok {
						return nil, fmt.Errorf("layer %q cell %d has gid %d not in any tileset", rl.Name, i, gid)
					}
				}
				gids[i] = gid
			}
			m.tileLayers[rl.Name] = &TileLayer{
				Name: rl.Name,
				Cols: w,
				Rows: h,
				GIDs: gids,
			}

		case "objectgroup":
			objs := make([]Object, len(rl.Objects))
			for i, ro := range rl.Objects {
				gid := ro.GID &^ gidFlipMask
				if gid != 0 {
					if _, _, ok := m.findTileset(gid); !ok {
						return nil, fmt.Errorf("layer %q object %d has gid %d not in any tileset", rl.Name, ro.ID, gid)
					}
				}
				objs[i] = Object{
					ID:         ro.ID,
					Name:       ro.Name,
					Type:       firstNonEmpty(ro.Type, ro.Class),
					X:          ro.X / tiledScale,
					Y:          ro.Y / tiledScale,
					W:          ro.Width / tiledScale,
					H:          ro.Height / tiledScale,
					GID:        gid,
					Point:      ro.Point,
					Properties: decodeProperties(ro.Properties),
				}
			}
			m.objectLayers[rl.Name] = &ObjectLayer{
				Name:    rl.Name,
				Objects: objs,
			}
		}
	}

	m.Collision = buildCollisionGrid(m)

	return m, nil
}

// allowedLayerType reports whether a layer with the given name and Tiled layer
// type is permitted by the map format.
func allowedLayerType(name, kind string) bool {
	switch name {
	case "ground", "decor", "above":
		return kind == "tilelayer"
	case "props", "objects":
		return kind == "objectgroup"
	case "collision":
		return kind == "tilelayer" || kind == "objectgroup"
	default:
		return false
	}
}

func decodeProperties(raw []tmjProperty) Properties {
	if len(raw) == 0 {
		return nil
	}
	p := make(Properties, len(raw))
	for _, prop := range raw {
		p[prop.Name] = prop.Value
	}
	return p
}

func firstNonEmpty(a, b string) string {
	if a != "" {
		return a
	}
	return b
}
