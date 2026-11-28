package world

import (
	"fmt"
	"slices"
	"testing"

	"github.com/Elpulgo/critterkingdom/internal/geom"
)

// propRoomMap returns a cols by rows Tiled map like roomMap, without walls,
// with the given objects and props object layers, both as JSON array bodies
// in Tiled pixels (128 to a tile, 8 to a world pixel).
func propRoomMap(cols, rows int, objects, props string) []byte {
	return fmt.Appendf(nil, `{
		"orientation": "orthogonal", "infinite": false,
		"width": %d, "height": %d, "tilewidth": 128, "tileheight": 128,
		"properties": [{"name": "biome", "type": "string", "value": "meadow"}],
		"layers": [
			{"name": "props", "type": "objectgroup", "objects": [%s]},
			{"name": "collision", "type": "objectgroup", "objects": []},
			{"name": "objects", "type": "objectgroup", "objects": [%s]}
		]
	}`, cols, rows, props, objects)
}

// treeProp returns a tree prop called name covering the world rectangle
// (x,y)-(x+16,y+32), so its feet are at world Y y+32.
func treeProp(id int, name string, x, y int) string {
	return fmt.Sprintf(`{"id": %d, "name": %q, "type": "tree", "x": %d, "y": %d, "width": 128, "height": 256}`,
		id, name, x*tiledScale, y*tiledScale)
}

// ysortView is the 96x64 world-pixel view at the top-left of the map.
var ysortView = geom.Rect{Max: geom.Vec{X: 96, Y: 64}}

// newPropScene starts a scene on a 10x6-tile (160x96 world pixel) meadow
// with the critter's spawn at (24,40) and the given extra objects and props.
func newPropScene(t *testing.T, objects, props string) *Scene {
	t.Helper()
	load := mapLoader(map[string][]byte{
		"meadow": propRoomMap(10, 6, villageStart+objects, props),
	})
	s, err := NewScene(nil, load, "meadow")
	if err != nil {
		t.Fatalf("NewScene: %v", err)
	}
	return s
}

// drawOrderOf names what s draws in view, back to front: a prop by its Tiled
// name, a villager by its id and the critter as "critter".
func drawOrderOf(s *Scene, view geom.Rect) []string {
	props := s.m.ObjectLayer("props")
	var got []string
	for _, a := range s.drawOrder(view) {
		switch a.kind {
		case actorProp:
			got = append(got, props.Objects[a.i].Name)
		case actorVillager:
			got = append(got, s.npcs[a.i].ID)
		case actorPlayer:
			got = append(got, "critter")
		default:
			got = append(got, fmt.Sprintf("kind %d #%d", a.kind, a.i))
		}
	}
	return got
}

func TestScene_TheCritterIsDrawnBehindATreeAboveItsBaseAndInFrontOfItBelow(t *testing.T) {
	// The oak covers world (32,16)-(48,48): its base is at Y 48.
	oak := treeProp(1, "oak", 32, 16)
	cases := []struct {
		name  string
		feetY float64
		want  []string
	}{
		{"feet above the base, behind the trunk", 40, []string{"critter", "oak"}},
		{"feet below the base, in front of the trunk", 56, []string{"oak", "critter"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s := newPropScene(t, ``, oak)
			s.player.Pos = geom.Vec{X: 40, Y: tc.feetY}

			if got := drawOrderOf(s, ysortView); !slices.Equal(got, tc.want) {
				t.Errorf("draw order with the critter's feet at Y %v = %q; want %q", tc.feetY, got, tc.want)
			}
		})
	}
}

func TestScene_OverlappingPropsAndCharactersAreDrawnBackToFrontWithCharactersOverAPropAndTheCritterOverAVillagerAtTheSameY(t *testing.T) {
	// The oak's base is at Y 48. Grandma stands just behind it at world
	// (36,44); the baker stands on its base line at world (44,48), and the
	// critter does too, at (40,48).
	villagers := `,` + pointNPC(2, "grandma_hedgehog", 36*tiledScale, 44*tiledScale) +
		`,` + pointNPC(3, "baker_badger", 44*tiledScale, 48*tiledScale)
	s := newPropScene(t, villagers, treeProp(1, "oak", 32, 16))
	s.player.Pos = geom.Vec{X: 40, Y: 48}

	want := []string{"grandma_hedgehog", "oak", "baker_badger", "critter"}
	if got := drawOrderOf(s, ysortView); !slices.Equal(got, want) {
		t.Errorf("draw order = %q; want %q", got, want)
	}
}

func TestScene_APropOutsideTheViewIsLeftOutOfTheDrawOrder(t *testing.T) {
	// The view ends at X 96 and a prop shape reaches 6 px past its
	// rectangle. The oak at X 32 is in view and the birch at X 100 is just
	// within reach of it; the pine at X 104 is beyond reach. The critter at
	// Y 40 stands behind the oak and the birch, whose bases are at Y 48.
	props := treeProp(1, "oak", 32, 16) + `,` + treeProp(2, "birch", 100, 16) + `,` + treeProp(3, "pine", 104, 16)
	s := newPropScene(t, ``, props)
	s.player.Pos = geom.Vec{X: 24, Y: 40}

	want := []string{"critter", "oak", "birch"}
	if got := drawOrderOf(s, ysortView); !slices.Equal(got, want) {
		t.Errorf("draw order = %q; want %q", got, want)
	}
}
