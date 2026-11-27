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

// drawProps draws every object in layer whose rectangle (grown by reach)
// overlaps view.
func drawProps(p *pen, m *Map, layer *ObjectLayer, view geom.Rect, elapsed time.Duration) {
	if layer == nil {
		return
	}
	near := view.Inset(-reach)
	for _, obj := range layer.Objects {
		r := obj.Rect()
		if !rectsOverlap(r, near) {
			continue
		}
		kind := obj.Type
		if obj.GID != 0 {
			frameGID := m.FrameGID(obj.GID, elapsed)
			if tile, ok := m.Tile(frameGID); ok && tile.Type != "" {
				kind = tile.Type
			}
		}
		w, h := r.Max.X-r.Min.X, r.Max.Y-r.Min.Y
		col := int(r.Min.X / m.TileSize)
		row := int(r.Min.Y / m.TileSize)
		drawTileArt(p, kind, r.Min.X, r.Min.Y, w, h, col, row)
	}
}

// rectsOverlap reports whether a and b share an area larger than zero.
func rectsOverlap(a, b geom.Rect) bool {
	return a.Min.X < b.Max.X && a.Max.X > b.Min.X && a.Min.Y < b.Max.Y && a.Max.Y > b.Min.Y
}
