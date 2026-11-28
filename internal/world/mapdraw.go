package world

import (
	"time"

	"github.com/Elpulgo/critterkingdom/internal/geom"
)

// drawTileLayer draws the visible cells [c0, c1) by [r0, r1) of layer, picking
// animated frames for elapsed.
func drawTileLayer(p *pen, m *Map, layer *TileLayer, c0, r0, c1, r1 int, elapsed time.Duration) {
	if layer == nil {
		return
	}
	ts := m.TileSize
	for row := r0; row < r1; row++ {
		for col := c0; col < c1; col++ {
			gid := layer.At(col, row)
			if gid == 0 {
				continue
			}
			frameGID := m.FrameGID(gid, elapsed)
			tile, ok := m.Tile(frameGID)
			var kind string
			if ok {
				kind = tile.Type
			}
			drawTileArt(p, kind, float64(col)*ts, float64(row)*ts, ts, ts, col, row)
		}
	}
}

// drawProp draws the prop object obj, picking its animated frame for elapsed.
func drawProp(p *pen, m *Map, obj Object, elapsed time.Duration) {
	r := obj.Rect()
	kind := propKind(m, obj, elapsed)
	w, h := r.Max.X-r.Min.X, r.Max.Y-r.Min.Y
	col := int(r.Min.X / m.TileSize)
	row := int(r.Min.Y / m.TileSize)
	drawTileArt(p, kind, r.Min.X, r.Min.Y, w, h, col, row)
}

// propKind returns the type of the prop object obj at elapsed: the type of
// its tile's animated frame when it is a tile object whose tile has a type,
// and otherwise the object's own type.
func propKind(m *Map, obj Object, elapsed time.Duration) string {
	if obj.GID == 0 {
		return obj.Type
	}
	if tile, ok := m.Tile(m.FrameGID(obj.GID, elapsed)); ok && tile.Type != "" {
		return tile.Type
	}
	return obj.Type
}

// flatProp reports whether a prop of the given type lies flat on the floor,
// like a rug, so everything standing on it is drawn over it.
func flatProp(kind string) bool {
	return kind == "rug"
}

// rectsOverlap reports whether a and b share an area larger than zero.
func rectsOverlap(a, b geom.Rect) bool {
	return a.Min.X < b.Max.X && a.Max.X > b.Min.X && a.Min.Y < b.Max.Y && a.Max.Y > b.Min.Y
}
