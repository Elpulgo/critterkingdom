package world

import (
	"strings"
	"testing"
	"time"

	"github.com/Elpulgo/critterkingdom/internal/geom"
)

func TestLoadMap_LoadsLayersObjectsPropertiesAndScalesPositionsToWorldPixels(t *testing.T) {
	const tmj = `{
		"orientation": "orthogonal",
		"infinite": false,
		"width": 3,
		"height": 2,
		"tilewidth": 128,
		"tileheight": 128,
		"properties": [
			{"name": "biome", "type": "string", "value": "meadow"},
			{"name": "dark", "type": "bool", "value": false},
			{"name": "wind", "type": "float", "value": 1.5}
		],
		"tilesets": [
			{
				"firstgid": 1,
				"name": "terrain",
				"tilecount": 4,
				"tiles": [
					{"id": 0, "type": "grass"},
					{"id": 1, "class": "path"}
				]
			}
		],
		"layers": [
			{
				"name": "ground",
				"type": "tilelayer",
				"width": 3,
				"height": 2,
				"data": [1, 2, 1, 0, 1, 2147483650]
			},
			{
				"name": "objects",
				"type": "objectgroup",
				"objects": [
					{
						"id": 7,
						"name": "start",
						"type": "player_spawn",
						"x": 256,
						"y": 384,
						"width": 0,
						"height": 0,
						"point": true,
						"properties": [
							{"name": "facing", "type": "string", "value": "down"}
						]
					},
					{
						"id": 8,
						"name": "pond_zone",
						"class": "spawn_zone",
						"x": 256,
						"y": 384,
						"width": 128,
						"height": 64,
						"properties": [
							{"name": "habitat", "type": "string", "value": "pond"},
							{"name": "max", "type": "int", "value": 3},
							{"name": "night", "type": "bool", "value": true}
						]
					}
				]
			},
			{
				"name": "props",
				"type": "objectgroup",
				"objects": [
					{
						"id": 9,
						"name": "oak",
						"type": "tree",
						"gid": 2147483651,
						"x": 128,
						"y": 256,
						"width": 128,
						"height": 128
					}
				]
			}
		]
	}`

	m, err := LoadMap([]byte(tmj))
	if err != nil {
		t.Fatalf("LoadMap: %v", err)
	}

	if m.Cols != 3 || m.Rows != 2 || m.TileSize != 16 {
		t.Fatalf("map dimensions = %dx%d tileSize %v; want 3x2 tileSize 16", m.Cols, m.Rows, m.TileSize)
	}
	if got, want := m.Size(), (geom.Vec{X: 48, Y: 32}); got != want {
		t.Errorf("Size() = %v; want %v", got, want)
	}
	if got := m.Properties.String("biome"); got != "meadow" {
		t.Errorf("Properties.String(biome) = %q; want meadow", got)
	}
	if got := m.Properties.Bool("dark"); got != false {
		t.Errorf("Properties.Bool(dark) = %v; want false", got)
	}
	if got := m.Properties.Float("wind"); got != 1.5 {
		t.Errorf("Properties.Float(wind) = %v; want 1.5", got)
	}

	ground := m.TileLayer("ground")
	if ground == nil {
		t.Fatal("TileLayer(ground) = nil")
	}
	wantGIDs := []uint32{1, 2, 1, 0, 1, 2}
	if len(ground.GIDs) != len(wantGIDs) {
		t.Fatalf("ground GIDs = %v; want %v", ground.GIDs, wantGIDs)
	}
	for i, want := range wantGIDs {
		if ground.GIDs[i] != want {
			t.Errorf("ground.GIDs[%d] = %d; want %d", i, ground.GIDs[i], want)
		}
	}
	if got := ground.At(2, 1); got != 2 {
		t.Errorf("ground.At(2, 1) = %d; want 2", got)
	}
	if m.TileLayer("decor") != nil {
		t.Errorf("TileLayer(decor) should be nil when absent")
	}
	if m.Collision == nil {
		t.Fatal("Map.Collision = nil when collision layer is absent; want empty Grid")
	}
	if m.Collision.Solid(0, 0) {
		t.Errorf("Collision.Solid(0, 0) = true on map without collision layer; want false")
	}

	objs := m.ObjectLayer("objects")
	if objs == nil || len(objs.Objects) != 2 {
		t.Fatalf("ObjectLayer(objects) = %v; want 2 objects", objs)
	}
	spawn := objs.Objects[0]
	if spawn.ID != 7 || spawn.Name != "start" || spawn.Type != "player_spawn" || !spawn.Point || spawn.X != 32 || spawn.Y != 48 {
		t.Errorf("spawn = %+v; want ID 7 start player_spawn point at 32,48", spawn)
	}
	if got := spawn.Properties.String("facing"); got != "down" {
		t.Errorf("spawn facing = %q; want down", got)
	}

	zone := objs.Objects[1]
	if zone.ID != 8 || zone.Name != "pond_zone" || zone.Type != "spawn_zone" || zone.Point {
		t.Errorf("zone = %+v; want ID 8 pond_zone spawn_zone", zone)
	}
	if zone.X != 32 || zone.Y != 48 || zone.W != 16 || zone.H != 8 {
		t.Errorf("zone geometry = (%v,%v %vx%v); want (32,48 16x8)", zone.X, zone.Y, zone.W, zone.H)
	}
	if zone.Properties.String("habitat") != "pond" || zone.Properties.Float("max") != 3 || !zone.Properties.Bool("night") {
		t.Errorf("zone properties = %v; want habitat=pond max=3 night=true", zone.Properties)
	}

	props := m.ObjectLayer("props")
	if props == nil || len(props.Objects) != 1 {
		t.Fatalf("ObjectLayer(props) = %v; want 1 object", props)
	}
	oak := props.Objects[0]
	if oak.ID != 9 || oak.Name != "oak" || oak.Type != "tree" || oak.GID != 3 || oak.X != 16 || oak.Y != 32 || oak.W != 16 || oak.H != 16 {
		t.Errorf("oak = %+v; want ID 9 oak tree GID 3 at 16,32 16x16", oak)
	}
}

func TestMap_Tile_ResolvesAcrossMultipleTilesetsAndIgnoresFlipBits(t *testing.T) {
	const tmj = `{
		"orientation": "orthogonal",
		"infinite": false,
		"width": 1,
		"height": 1,
		"tilewidth": 128,
		"tileheight": 128,
		"tilesets": [
			{
				"firstgid": 1,
				"name": "ground",
				"tilecount": 4,
				"tiles": [
					{"id": 0, "type": "grass"}
				]
			},
			{
				"firstgid": 5,
				"name": "water",
				"tilecount": 3,
				"tiles": [
					{
						"id": 1,
						"type": "water_1",
						"properties": [
							{"name": "swimmable", "type": "bool", "value": true}
						],
						"animation": [
							{"tileid": 1, "duration": 250},
							{"tileid": 2, "duration": 300}
						]
					}
				]
			}
		],
		"layers": []
	}`

	m, err := LoadMap([]byte(tmj))
	if err != nil {
		t.Fatalf("LoadMap: %v", err)
	}

	// GID 6 is local tile 1 in the second tileset (firstgid 5).
	tile, ok := m.Tile(6)
	if !ok {
		t.Fatal("Tile(6) not found")
	}
	if tile.Type != "water_1" || !tile.Properties.Bool("swimmable") {
		t.Errorf("Tile(6) = %+v; want type water_1 swimmable=true", tile)
	}
	wantAnim := []Frame{
		{TileID: 1, Duration: 250 * time.Millisecond},
		{TileID: 2, Duration: 300 * time.Millisecond},
	}
	if len(tile.Animation) != len(wantAnim) || tile.Animation[0] != wantAnim[0] || tile.Animation[1] != wantAnim[1] {
		t.Errorf("Tile(6).Animation = %v; want %v", tile.Animation, wantAnim)
	}

	// With horizontal + vertical + diagonal flip bits set on GID 6, it resolves to the same tile.
	flipped, ok := m.Tile(6 | 0xE0000000)
	if !ok || flipped.Type != "water_1" || !flipped.Properties.Bool("swimmable") {
		t.Errorf("Tile(6|flips) = %+v, %v; want water_1, true", flipped, ok)
	}

	// GID 2 is within the first tileset's tilecount (4) even though it has no entry in tiles.
	unlisted, ok := m.Tile(2)
	if !ok || unlisted.Type != "" {
		t.Errorf("Tile(2) = %+v, %v; want empty tile, true", unlisted, ok)
	}

	if _, ok := m.Tile(0); ok {
		t.Errorf("Tile(0) should return false")
	}
	if _, ok := m.Tile(8); ok {
		t.Errorf("Tile(8) outside tilesets should return false")
	}
}

func TestObject_Rect_PlacesTileObjectAboveItsBottomLeftAnchor(t *testing.T) {
	rectObj := Object{X: 32, Y: 48, W: 16, H: 8}
	if got, want := rectObj.Rect(), (geom.Rect{Min: geom.Vec{X: 32, Y: 48}, Max: geom.Vec{X: 48, Y: 56}}); got != want {
		t.Errorf("rectangle Object.Rect() = %v; want %v", got, want)
	}

	tileObj := Object{X: 32, Y: 48, W: 16, H: 24, GID: 5}
	if got, want := tileObj.Rect(), (geom.Rect{Min: geom.Vec{X: 32, Y: 24}, Max: geom.Vec{X: 48, Y: 48}}); got != want {
		t.Errorf("tile Object.Rect() = %v; want %v", got, want)
	}
}

func TestLoadMap_RejectsUnsupportedOrMalformedMaps(t *testing.T) {
	tests := []struct {
		name    string
		json    string
		wantErr string
	}{
		{
			name:    "invalid JSON",
			json:    `{not valid json`,
			wantErr: "parse map json",
		},
		{
			name:    "non-orthogonal map",
			json:    `{"orientation":"isometric","infinite":false,"width":2,"height":2,"tilewidth":128,"tileheight":128}`,
			wantErr: "orthogonal",
		},
		{
			name:    "infinite map",
			json:    `{"orientation":"orthogonal","infinite":true,"width":2,"height":2,"tilewidth":128,"tileheight":128}`,
			wantErr: "infinite",
		},
		{
			name:    "tile size other than 128",
			json:    `{"orientation":"orthogonal","infinite":false,"width":2,"height":2,"tilewidth":16,"tileheight":16}`,
			wantErr: "128x128",
		},
		{
			name: "external tileset with source",
			json: `{
				"orientation":"orthogonal","infinite":false,"width":1,"height":1,"tilewidth":128,"tileheight":128,
				"tilesets":[{"firstgid":1,"source":"outdoor.tsj","tilecount":1}]
			}`,
			wantErr: "Embed Tilesets",
		},
		{
			name: "encoded layer data",
			json: `{
				"orientation":"orthogonal","infinite":false,"width":1,"height":1,"tilewidth":128,"tileheight":128,
				"tilesets":[{"firstgid":1,"name":"t","tilecount":1}],
				"layers":[{"name":"ground","type":"tilelayer","width":1,"height":1,"encoding":"base64","data":"AQAAAA=="}]
			}`,
			wantErr: "CSV",
		},
		{
			name: "unknown layer name",
			json: `{
				"orientation":"orthogonal","infinite":false,"width":1,"height":1,"tilewidth":128,"tileheight":128,
				"tilesets":[{"firstgid":1,"name":"t","tilecount":1}],
				"layers":[{"name":"sky","type":"tilelayer","width":1,"height":1,"data":[1]}]
			}`,
			wantErr: `"sky"`,
		},
		{
			name: "known layer of wrong kind",
			json: `{
				"orientation":"orthogonal","infinite":false,"width":1,"height":1,"tilewidth":128,"tileheight":128,
				"tilesets":[{"firstgid":1,"name":"t","tilecount":1}],
				"layers":[{"name":"ground","type":"objectgroup","objects":[]}]
			}`,
			wantErr: `"ground"`,
		},
		{
			name: "duplicate layer name",
			json: `{
				"orientation":"orthogonal","infinite":false,"width":1,"height":1,"tilewidth":128,"tileheight":128,
				"tilesets":[{"firstgid":1,"name":"t","tilecount":1}],
				"layers":[
					{"name":"ground","type":"tilelayer","width":1,"height":1,"data":[1]},
					{"name":"ground","type":"tilelayer","width":1,"height":1,"data":[1]}
				]
			}`,
			wantErr: "duplicate",
		},
		{
			name: "layer data length not width times height",
			json: `{
				"orientation":"orthogonal","infinite":false,"width":2,"height":2,"tilewidth":128,"tileheight":128,
				"tilesets":[{"firstgid":1,"name":"t","tilecount":2}],
				"layers":[{"name":"ground","type":"tilelayer","width":2,"height":2,"data":[1,1,1]}]
			}`,
			wantErr: "3 gids for 2x2",
		},
		{
			name: "gid not covered by any tileset",
			json: `{
				"orientation":"orthogonal","infinite":false,"width":1,"height":1,"tilewidth":128,"tileheight":128,
				"tilesets":[{"firstgid":1,"name":"t","tilecount":2}],
				"layers":[{"name":"ground","type":"tilelayer","width":1,"height":1,"data":[5]}]
			}`,
			wantErr: "gid 5",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := LoadMap([]byte(tt.json))
			if err == nil {
				t.Fatalf("LoadMap succeeded for %s; want error containing %q", tt.name, tt.wantErr)
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("LoadMap error = %q; want it to contain %q", err.Error(), tt.wantErr)
			}
		})
	}
}
