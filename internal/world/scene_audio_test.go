package world

import (
	"errors"
	"fmt"
	"slices"
	"testing"

	"github.com/Elpulgo/critterkingdom/internal/geom"
)

// audioTick is one Tick call heard by recordingAudio.
type audioTick struct {
	dt      float64
	walking bool
}

// recordingAudio stands in for the game's sound manager, which ends at the
// audio device: it records every track asked for and every tick.
type recordingAudio struct {
	tracks []string
	ticks  []audioTick
}

func (a *recordingAudio) SetMusic(track string) { a.tracks = append(a.tracks, track) }

func (a *recordingAudio) Tick(dt float64, walking bool) {
	a.ticks = append(a.ticks, audioTick{dt, walking})
}

// musicRoomMap returns a cols by rows Tiled map like roomMap whose music
// property is music.
func musicRoomMap(cols, rows int, music, objects, walls string) []byte {
	return fmt.Appendf(nil, `{
		"orientation": "orthogonal", "infinite": false,
		"width": %d, "height": %d, "tilewidth": 128, "tileheight": 128,
		"properties": [{"name": "music", "type": "string", "value": %q}],
		"layers": [
			{"name": "collision", "type": "objectgroup", "objects": [%s]},
			{"name": "objects", "type": "objectgroup", "objects": [%s]}
		]
	}`, cols, rows, music, walls, objects)
}

// newAudioScene starts a scene with a on the map start of maps, each loaded
// through LoadMap.
func newAudioScene(t *testing.T, maps map[string][]byte, start string, a Audio) *Scene {
	t.Helper()
	load := func(name string) (*Map, error) {
		data, ok := maps[name]
		if !ok {
			return nil, errors.New("not found")
		}
		return LoadMap(data)
	}
	s, err := NewScene(nil, load, start, WithAudio(a))
	if err != nil {
		t.Fatalf("NewScene: %v", err)
	}
	return s
}

func TestScene_EnteringTheStartMapAndFollowingAWarpAskForThatMapsMusic(t *testing.T) {
	a := &recordingAudio{}
	s := newAudioScene(t, map[string][]byte{
		"village": musicRoomMap(6, 4, "meadow_song", villageStart+`,`+warpObject(2, 4, 2, "house", "from_village"), ``),
		"house":   musicRoomMap(5, 4, "fireside", houseObjects, houseWall),
	}, "village", a)

	if want := []string{"meadow_song"}; !slices.Equal(a.tracks, want) {
		t.Fatalf("after NewScene: tracks asked for = %q; want %q", a.tracks, want)
	}

	updateUntilMap(t, s, "house", walkFrame(geom.Vec{X: 1}))

	if want := []string{"meadow_song", "fireside"}; !slices.Equal(a.tracks, want) {
		t.Errorf("after the warp: tracks asked for = %q; want %q", a.tracks, want)
	}
}

func TestScene_ReportsWalkingOnlyOnTicksTheCritterMoved(t *testing.T) {
	a := &recordingAudio{}
	// The garden's wall stops the critter walking east from its start at
	// x 24 at x 42: 18 px at 1.2 px a tick, so after 15 of the 60 ticks it
	// walks.
	s := newAudioScene(t, map[string][]byte{
		"garden": musicRoomMap(6, 4, "meadow_song", gardenStart, gardenWall),
	}, "garden", a)
	stand := walkFrame(geom.Vec{})
	east := walkFrame(geom.Vec{X: 1})

	update(t, s, stand, 1)
	update(t, s, east, 60)

	if len(a.ticks) != 61 {
		t.Fatalf("Tick called %d times in 61 updates; want once per update", len(a.ticks))
	}
	for i, tk := range a.ticks {
		if tk.dt != 1.0/60 {
			t.Errorf("tick %d: dt = %v; want the frame's DT, 1/60", i, tk.dt)
		}
	}
	if s.player.Pos.X != 42 {
		t.Fatalf("critter at %v; want stopped on the wall at x 42", s.player.Pos)
	}
	want := map[int]bool{
		0:  false, // standing still
		1:  true,  // first step east
		10: true,  // still walking, 12 px along
		60: false, // pressed against the wall
	}
	for i, walking := range want {
		if a.ticks[i].walking != walking {
			t.Errorf("tick %d: walking = %v; want %v", i, a.ticks[i].walking, walking)
		}
	}
}
