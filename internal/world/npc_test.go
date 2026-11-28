package world

import (
	"strings"
	"testing"

	"github.com/Elpulgo/critterkingdom/internal/geom"
)

func TestMap_NPCs_ARectangleIsAWanderAreaWithTheVillagerAtItsCentreAndAPointStandsStill(t *testing.T) {
	// The baker's rectangle covers Tiled (128,256)..(512,640), so world
	// (16,32)..(64,80) with its centre at (40,56). The point npc is at Tiled
	// (640,384), so world (80,48).
	m, err := LoadMap(markerMap(``, `
		{"id": 7, "type": "npc", "x": 128, "y": 256, "width": 384, "height": 384,
			"properties": [{"name": "id", "type": "string", "value": "baker_badger"}]},
		{"id": 9, "type": "npc", "x": 640, "y": 384, "point": true,
			"properties": [{"name": "id", "type": "string", "value": "grandma_hedgehog"}]}`))
	if err != nil {
		t.Fatalf("LoadMap: %v", err)
	}

	want := []NPCSpawn{
		{
			ObjectID: 7,
			ID:       "baker_badger",
			Pos:      geom.Vec{X: 40, Y: 56},
			Area:     geom.Rect{Min: geom.Vec{X: 16, Y: 32}, Max: geom.Vec{X: 64, Y: 80}},
		},
		{
			ObjectID: 9,
			ID:       "grandma_hedgehog",
			Pos:      geom.Vec{X: 80, Y: 48},
			Area:     geom.Rect{Min: geom.Vec{X: 80, Y: 48}, Max: geom.Vec{X: 80, Y: 48}},
		},
	}
	got := m.NPCs()
	if len(got) != len(want) {
		t.Fatalf("NPCs() = %+v; want %+v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("NPCs()[%d] = %+v; want %+v", i, got[i], want[i])
		}
	}
}

func TestLoadMap_RejectsNPCsWithoutAnIDOrWithAnIDAlreadyUsed(t *testing.T) {
	tests := []struct {
		name    string
		objects string
		wantErr string
	}{
		{
			name:    "no id property",
			objects: `{"id": 21, "type": "npc", "x": 128, "y": 128, "point": true}`,
			wantErr: "object 21",
		},
		{
			name: "empty id",
			objects: `{"id": 22, "type": "npc", "x": 128, "y": 128, "point": true,
				"properties": [{"name": "id", "type": "string", "value": ""}]}`,
			wantErr: "object 22",
		},
		{
			name: "id used twice",
			objects: `{"id": 23, "type": "npc", "x": 128, "y": 128, "point": true,
					"properties": [{"name": "id", "type": "string", "value": "mayor_owl"}]},
				{"id": 24, "type": "npc", "x": 256, "y": 128, "point": true,
					"properties": [{"name": "id", "type": "string", "value": "mayor_owl"}]}`,
			wantErr: "object 24",
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
