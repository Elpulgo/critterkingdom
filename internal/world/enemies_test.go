package world

import (
	"strings"
	"testing"

	"github.com/Elpulgo/critterkingdom/internal/geom"
)

func TestMap_Enemies_GivesEachEnemyObjectsIDAndWorldPositionInLayerOrder(t *testing.T) {
	// A spawn sits between the two enemies to show that only enemy objects
	// are returned, and that layer order is kept.
	m, err := LoadMap(markerMap(``, `
		{"id": 7, "type": "enemy", "x": 320, "y": 200, "point": true,
			"properties": [{"name": "enemy", "type": "string", "value": "thistle_beast"}]},
		{"id": 8, "name": "start", "type": "player_spawn", "x": 128, "y": 128, "point": true},
		{"id": 9, "type": "enemy", "x": 64, "y": 448, "point": true,
			"properties": [{"name": "enemy", "type": "string", "value": "sulky_slime"}]}`))
	if err != nil {
		t.Fatalf("LoadMap: %v", err)
	}

	// Tiled (320,200) / 8 = world (40,25); Tiled (64,448) / 8 = world (8,56).
	want := []EnemyMarker{
		{ObjectID: 7, EnemyID: "thistle_beast", Pos: geom.Vec{X: 40, Y: 25}},
		{ObjectID: 9, EnemyID: "sulky_slime", Pos: geom.Vec{X: 8, Y: 56}},
	}
	got := m.Enemies()
	if len(got) != len(want) {
		t.Fatalf("Enemies() = %+v; want %+v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("Enemies()[%d] = %+v; want %+v", i, got[i], want[i])
		}
	}
}

func TestLoadMap_RejectsAnEnemyObjectWithoutAnEnemyID(t *testing.T) {
	tests := []struct {
		name       string
		properties string
	}{
		{"no enemy property", ``},
		{"empty enemy property", `{"name": "enemy", "type": "string", "value": ""}`},
		{"enemy property is a number", `{"name": "enemy", "type": "int", "value": 3}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := LoadMap(markerMap(``, `{"id": 21, "type": "enemy", "x": 128, "y": 128, "point": true,
				"properties": [`+tt.properties+`]}`))
			if err == nil {
				t.Fatal("LoadMap succeeded; want error naming object 21")
			}
			if !strings.Contains(err.Error(), "object 21") {
				t.Errorf("LoadMap error = %q; want it to contain %q", err.Error(), "object 21")
			}
		})
	}
}
