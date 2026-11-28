package world

import (
	"strings"
	"testing"
	"time"

	"github.com/Elpulgo/critterkingdom/internal/geom"
)

func TestMap_SpawnZones_CarryRectHabitatIDAndDefaultOrGivenLimits(t *testing.T) {
	// The first zone covers Tiled (128,256)..(640,512), so world (16,32)..(80,64).
	m, err := LoadMap(markerMap(``, `
		{"id": 7, "name": "lily_shore", "type": "spawn_zone",
			"x": 128, "y": 256, "width": 512, "height": 256,
			"properties": [{"name": "habitat", "type": "string", "value": "pond"}]},
		{"id": 9, "name": "", "type": "spawn_zone",
			"x": 0, "y": 0, "width": 256, "height": 128,
			"properties": [
				{"name": "habitat", "type": "string", "value": "meadow"},
				{"name": "max", "type": "int", "value": 5},
				{"name": "respawn_s", "type": "float", "value": 12.5}
			]}`))
	if err != nil {
		t.Fatalf("LoadMap: %v", err)
	}

	want := []SpawnZone{
		{
			ID:           "lily_shore",
			Rect:         geom.Rect{Min: geom.Vec{X: 16, Y: 32}, Max: geom.Vec{X: 80, Y: 64}},
			Habitat:      "pond",
			Max:          3,
			RespawnAfter: 60 * time.Second,
		},
		{
			ID:           "zone-9",
			Rect:         geom.Rect{Min: geom.Vec{X: 0, Y: 0}, Max: geom.Vec{X: 32, Y: 16}},
			Habitat:      "meadow",
			Max:          5,
			RespawnAfter: 12500 * time.Millisecond,
		},
	}
	got := m.SpawnZones()
	if len(got) != len(want) {
		t.Fatalf("SpawnZones() = %+v; want %+v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("SpawnZones()[%d] = %+v; want %+v", i, got[i], want[i])
		}
	}
}

func TestLoadMap_RejectsSpawnZonesThatCannotBeUsed(t *testing.T) {
	const pond = `{"name": "habitat", "type": "string", "value": "pond"}`
	tests := []struct {
		name    string
		objects string
		wantErr string
	}{
		{
			name:    "zero width",
			objects: `{"id": 21, "type": "spawn_zone", "x": 0, "y": 0, "width": 0, "height": 128, "properties": [` + pond + `]}`,
			wantErr: "object 21",
		},
		{
			name:    "zero height",
			objects: `{"id": 22, "type": "spawn_zone", "x": 0, "y": 0, "width": 128, "height": 0, "properties": [` + pond + `]}`,
			wantErr: "object 22",
		},
		{
			name:    "no habitat",
			objects: `{"id": 23, "type": "spawn_zone", "x": 0, "y": 0, "width": 128, "height": 128}`,
			wantErr: "object 23",
		},
		{
			name: "max 0",
			objects: `{"id": 24, "type": "spawn_zone", "x": 0, "y": 0, "width": 128, "height": 128, "properties": [` + pond + `,
				{"name": "max", "type": "int", "value": 0}]}`,
			wantErr: "object 24",
		},
		{
			name: "negative respawn",
			objects: `{"id": 25, "type": "spawn_zone", "x": 0, "y": 0, "width": 128, "height": 128, "properties": [` + pond + `,
				{"name": "respawn_s", "type": "float", "value": -1}]}`,
			wantErr: "object 25",
		},
		{
			name: "duplicate name",
			objects: `{"id": 26, "name": "shore", "type": "spawn_zone", "x": 0, "y": 0, "width": 128, "height": 128, "properties": [` + pond + `]},
				{"id": 27, "name": "shore", "type": "spawn_zone", "x": 256, "y": 0, "width": 128, "height": 128, "properties": [` + pond + `]}`,
			wantErr: "object 27",
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
