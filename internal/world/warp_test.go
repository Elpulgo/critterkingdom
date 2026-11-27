package world

import (
	"strings"
	"testing"

	"github.com/Elpulgo/critterkingdom/internal/geom"
)

// markerMap returns a 6x4 Tiled map with the given map properties and objects
// layer objects, both as JSON array bodies.
func markerMap(properties, objects string) []byte {
	return []byte(`{
		"orientation": "orthogonal", "infinite": false,
		"width": 6, "height": 4, "tilewidth": 128, "tileheight": 128,
		"properties": [` + properties + `],
		"layers": [{"name": "objects", "type": "objectgroup", "objects": [` + objects + `]}]
	}`)
}

func TestLoadMap_KeepsMusicBiomeAndDarkOnTheMap(t *testing.T) {
	tests := []struct {
		name       string
		properties string
		wantMusic  string
		wantBiome  string
		wantDark   bool
	}{
		{
			name: "all three set",
			properties: `{"name": "music", "type": "string", "value": "cave_drips"},
				{"name": "biome", "type": "string", "value": "cave"},
				{"name": "dark", "type": "bool", "value": true}`,
			wantMusic: "cave_drips",
			wantBiome: "cave",
			wantDark:  true,
		},
		{
			name:       "none set",
			properties: ``,
			wantMusic:  "",
			wantBiome:  "",
			wantDark:   false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m, err := LoadMap(markerMap(tt.properties, ``))
			if err != nil {
				t.Fatalf("LoadMap: %v", err)
			}
			if m.Music != tt.wantMusic || m.Biome != tt.wantBiome || m.Dark != tt.wantDark {
				t.Errorf("Music, Biome, Dark = %q, %q, %v; want %q, %q, %v",
					m.Music, m.Biome, m.Dark, tt.wantMusic, tt.wantBiome, tt.wantDark)
			}
		})
	}
}

func TestLoadMap_RejectsMistypedMusicBiomeOrDark(t *testing.T) {
	tests := []struct {
		name       string
		properties string
		wantErr    string
	}{
		{"music is a number", `{"name": "music", "type": "int", "value": 3}`, `"music"`},
		{"biome is a bool", `{"name": "biome", "type": "bool", "value": true}`, `"biome"`},
		{"dark is a string", `{"name": "dark", "type": "string", "value": "yes"}`, `"dark"`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := LoadMap(markerMap(tt.properties, ``))
			if err == nil {
				t.Fatalf("LoadMap succeeded; want error containing %s", tt.wantErr)
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("LoadMap error = %q; want it to contain %s", err.Error(), tt.wantErr)
			}
		})
	}
}

func TestMap_WarpAt_FindsTheWarpWhoseAreaHoldsThePointWithMaxEdgesOutside(t *testing.T) {
	// The door warp covers Tiled (128,256)..(384,384), so world (16,32)..(48,48).
	m, err := LoadMap(markerMap(``, `{
		"id": 4, "name": "door", "type": "warp",
		"x": 128, "y": 256, "width": 256, "height": 128,
		"properties": [
			{"name": "target_map", "type": "string", "value": "house"},
			{"name": "target_spawn", "type": "string", "value": "from_village"}
		]
	}`))
	if err != nil {
		t.Fatalf("LoadMap: %v", err)
	}

	want := Warp{
		Rect:        geom.Rect{Min: geom.Vec{X: 16, Y: 32}, Max: geom.Vec{X: 48, Y: 48}},
		TargetMap:   "house",
		TargetSpawn: "from_village",
	}
	warps := m.Warps()
	if len(warps) != 1 || warps[0] != want {
		t.Fatalf("Warps() = %+v; want [%+v]", warps, want)
	}

	tests := []struct {
		name   string
		pos    geom.Vec
		wantOK bool
	}{
		{"on the min corner", geom.Vec{X: 16, Y: 32}, true},
		{"just inside the max corner", geom.Vec{X: 47.9, Y: 47.9}, true},
		{"on the right edge", geom.Vec{X: 48, Y: 40}, false},
		{"on the bottom edge", geom.Vec{X: 30, Y: 48}, false},
		{"left of the area", geom.Vec{X: 15.9, Y: 40}, false},
		{"far away", geom.Vec{X: 90, Y: 10}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := m.WarpAt(tt.pos)
			if ok != tt.wantOK {
				t.Fatalf("WarpAt(%v) ok = %v; want %v", tt.pos, ok, tt.wantOK)
			}
			if ok && got != want {
				t.Errorf("WarpAt(%v) = %+v; want %+v", tt.pos, got, want)
			}
		})
	}
}

func TestMap_Spawn_FindsSpawnsByNameAndTheFirstIsTheDefault(t *testing.T) {
	m, err := LoadMap(markerMap(``, `
		{"id": 2, "name": "start", "type": "player_spawn", "x": 256, "y": 384, "point": true},
		{"id": 3, "name": "from_house", "type": "player_spawn", "x": 512, "y": 128, "point": true}`))
	if err != nil {
		t.Fatalf("LoadMap: %v", err)
	}

	if got, ok := m.Spawn("from_house"); !ok || got != (geom.Vec{X: 64, Y: 16}) {
		t.Errorf("Spawn(from_house) = %v, %v; want (64,16), true", got, ok)
	}
	if got, ok := m.Spawn("start"); !ok || got != (geom.Vec{X: 32, Y: 48}) {
		t.Errorf("Spawn(start) = %v, %v; want (32,48), true", got, ok)
	}
	if _, ok := m.Spawn("from_meadow"); ok {
		t.Errorf("Spawn(from_meadow) ok = true; want false for an unknown name")
	}
	if got, ok := m.DefaultSpawn(); !ok || got != (geom.Vec{X: 32, Y: 48}) {
		t.Errorf("DefaultSpawn() = %v, %v; want the first spawn (32,48), true", got, ok)
	}

	empty, err := LoadMap(markerMap(``, ``))
	if err != nil {
		t.Fatalf("LoadMap without spawns: %v", err)
	}
	if _, ok := empty.DefaultSpawn(); ok {
		t.Errorf("DefaultSpawn() ok = true on a map without player_spawn; want false")
	}
}

func TestLoadMap_RejectsSpawnsAndWarpsThatCannotBeUsed(t *testing.T) {
	const targets = `"properties": [
		{"name": "target_map", "type": "string", "value": "meadow"},
		{"name": "target_spawn", "type": "string", "value": "from_village"}]`
	tests := []struct {
		name    string
		objects string
		wantErr string
	}{
		{
			name:    "unnamed spawn",
			objects: `{"id": 11, "name": "", "type": "player_spawn", "x": 128, "y": 128, "point": true}`,
			wantErr: "object 11",
		},
		{
			name: "duplicate spawn name",
			objects: `{"id": 11, "name": "start", "type": "player_spawn", "x": 128, "y": 128, "point": true},
				{"id": 12, "name": "start", "type": "player_spawn", "x": 256, "y": 128, "point": true}`,
			wantErr: "object 12",
		},
		{
			name: "warp without target_map",
			objects: `{"id": 13, "type": "warp", "x": 0, "y": 0, "width": 128, "height": 128,
				"properties": [{"name": "target_spawn", "type": "string", "value": "from_village"}]}`,
			wantErr: "object 13",
		},
		{
			name: "warp with empty target_map",
			objects: `{"id": 13, "type": "warp", "x": 0, "y": 0, "width": 128, "height": 128,
				"properties": [{"name": "target_map", "type": "string", "value": ""},
					{"name": "target_spawn", "type": "string", "value": "from_village"}]}`,
			wantErr: "object 13",
		},
		{
			name: "warp without target_spawn",
			objects: `{"id": 14, "type": "warp", "x": 0, "y": 0, "width": 128, "height": 128,
				"properties": [{"name": "target_map", "type": "string", "value": "meadow"}]}`,
			wantErr: "object 14",
		},
		{
			name:    "warp with zero width",
			objects: `{"id": 15, "type": "warp", "x": 0, "y": 0, "width": 0, "height": 128, ` + targets + `}`,
			wantErr: "object 15",
		},
		{
			name:    "warp with zero height",
			objects: `{"id": 16, "type": "warp", "x": 0, "y": 0, "width": 128, "height": 0, ` + targets + `}`,
			wantErr: "object 16",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := LoadMap(markerMap(``, tt.objects))
			if err == nil {
				t.Fatalf("LoadMap succeeded; want error containing %q", tt.wantErr)
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("LoadMap error = %q; want it to contain %q", err.Error(), tt.wantErr)
			}
		})
	}
}
