package world

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/Elpulgo/critterkingdom/internal/combat"
	"github.com/Elpulgo/critterkingdom/internal/engine"
	"github.com/Elpulgo/critterkingdom/internal/geom"
)

// slimeDef is a Sulky Slime as in enemies.json: it hops 30 px a second for
// HopDuration after a 0.6 s telegraph, and each hit takes one petal.
var slimeDef = combat.EnemyDef{
	ID:               "sulky_slime",
	Behaviour:        combat.BehaviourHopToward,
	HP:               1,
	Damage:           1,
	Speed:            30,
	TelegraphSeconds: 0.6,
}

// enemyObject returns an enemy object of kind enemyID standing at the world
// pixel x, y.
func enemyObject(id int, x, y float64, enemyID string) string {
	return fmt.Sprintf(`{"id": %d, "type": "enemy", "x": %v, "y": %v, "point": true,
		"properties": [{"name": "enemy", "type": "string", "value": %q}]}`,
		id, x*8, y*8, enemyID)
}

// mapsLoader loads the inline maps through the real LoadMap. Any other name
// fails.
func mapsLoader(maps map[string][]byte) MapLoader {
	return func(name string) (*Map, error) {
		data, ok := maps[name]
		if !ok {
			return nil, errors.New("not found")
		}
		return LoadMap(data)
	}
}

// standFrame is one tick at 60 ticks per second with no input.
var standFrame = engine.Frame{View: villageView, DT: 1.0 / 60}

func TestScene_ASlimeTakesNoPetalDuringItsTelegraphAndOneWhenItsHopLands(t *testing.T) {
	// The critter stands on the spawn at (24,40); the slime at (33,40) is
	// close enough that their hurt-boxes, 9.6 wide each, already touch.
	load := mapsLoader(map[string][]byte{
		"arena": roomMap(6, 4, "arena", villageStart+`,`+enemyObject(2, 33, 40, "sulky_slime"), ``),
	})
	s, err := NewScene(nil, load, "arena", WithEnemies([]combat.EnemyDef{slimeDef}))
	if err != nil {
		t.Fatalf("NewScene: %v", err)
	}

	// The slime notices the critter on tick 1, at 1/60 s, so its telegraph
	// ends at 37/60 s, tick 37. Tick 36 is the last step short of it.
	for tick := 1; tick <= 36; tick++ {
		if err := s.Update(standFrame); err != nil {
			t.Fatalf("Update: %v", err)
		}
		if got := s.health.Petals(); got != 3 {
			t.Fatalf("tick %d, %v into the game: %d petals; want 3, no hit before the 0.6 s telegraph has run", tick, s.now(), got)
		}
	}
	// The hop runs from tick 37 to tick 67 and the slime then rests 0.6 s,
	// so by tick 100 the hop has landed and no second attack has begun.
	for tick := 37; tick <= 100; tick++ {
		if err := s.Update(standFrame); err != nil {
			t.Fatalf("Update: %v", err)
		}
	}
	if got := s.health.Petals(); got != 2 {
		t.Errorf("after the hop: %d petals; want 2, exactly one petal lost", got)
	}
}

func TestScene_ACritterHitToZeroPetalsWakesWhereItCameInWithFullPetalsAndTheEnemiesReset(t *testing.T) {
	// The arena's first spawn is far from the warp's target from_village at
	// (24,40), so waking there shows the nap uses the way the critter came in.
	// The slime at (60,40) notices the critter on arrival; the second slime
	// at (110,70) stays too far away to notice it.
	arena := strings.Join([]string{
		`{"id": 1, "name": "start", "type": "player_spawn", "x": 800, "y": 480, "point": true}`,
		`{"id": 2, "name": "from_village", "type": "player_spawn", "x": 192, "y": 320, "point": true}`,
		enemyObject(3, 60, 40, "sulky_slime"),
		enemyObject(4, 110, 70, "sulky_slime"),
	}, ",")
	load := mapsLoader(map[string][]byte{
		"village": roomMap(6, 4, "village", villageStart+`,`+warpObject(2, 4, 2, "arena", "from_village"), ``),
		"arena":   roomMap(8, 5, "arena", arena, ``),
	})
	s, err := NewScene(nil, load, "village", WithEnemies([]combat.EnemyDef{slimeDef}))
	if err != nil {
		t.Fatalf("NewScene: %v", err)
	}
	updateUntilMap(t, s, "arena", walkFrame(geom.Vec{X: 1}))

	// Walk 18 px east, away from the entrance and toward the slime, and
	// defeat the far slime so the reset has to bring it back.
	for range 15 {
		if err := s.Update(walkFrame(geom.Vec{X: 1})); err != nil {
			t.Fatalf("Update: %v", err)
		}
	}
	if !s.enemies[1].Hit(1) || !s.enemies[1].Defeated() {
		t.Fatal("the far slime was not defeated by a hit of 1")
	}

	napped := false
	for range 1000 {
		before := s.health.Petals()
		if err := s.Update(standFrame); err != nil {
			t.Fatalf("Update: %v", err)
		}
		if before == 1 && s.health.Petals() != 1 {
			napped = true
			break
		}
	}
	if !napped {
		t.Fatalf("no nap after 1000 ticks: %d petals, critter at %v", s.health.Petals(), s.player.Pos)
	}

	if s.player.Pos != (geom.Vec{X: 24, Y: 40}) {
		t.Errorf("woke at %v; want at the from_village spawn the warp led to, 24,40", s.player.Pos)
	}
	if s.health.Petals() != 3 || s.health.Napping() {
		t.Errorf("woke with %d petals, napping %v; want 3 petals and awake", s.health.Petals(), s.health.Napping())
	}
	wantStarts := []geom.Vec{{X: 60, Y: 40}, {X: 110, Y: 70}}
	if len(s.enemies) != len(wantStarts) {
		t.Fatalf("%d enemies after the nap; want %d, the defeated slime back too", len(s.enemies), len(wantStarts))
	}
	for i, e := range s.enemies {
		if e.Pos != wantStarts[i] || e.HP() != 1 || e.Defeated() {
			t.Errorf("enemy %d after the nap: at %v, hp %d, defeated %v; want at %v, hp 1, not defeated",
				i, e.Pos, e.HP(), e.Defeated(), wantStarts[i])
		}
	}
}

func TestScene_StartingOnAMapWithAnEnemyWithoutADefinitionFailsNamingIt(t *testing.T) {
	load := mapsLoader(map[string][]byte{
		"arena": roomMap(6, 4, "arena", villageStart+`,`+enemyObject(5, 60, 40, "grumpy_badger"), ``),
	})

	s, err := NewScene(nil, load, "arena", WithEnemies([]combat.EnemyDef{slimeDef}))
	if err == nil {
		t.Fatalf("NewScene = %v; want an error for the enemy grumpy_badger", s)
	}
	for _, want := range []string{"object 5", `"grumpy_badger"`} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("NewScene error = %q; want it to contain %s", err, want)
		}
	}
}

func TestScene_OnLittleExplorerAnEnemyHitTakesItsOverriddenDamage(t *testing.T) {
	// A slime whose hit takes two petals, one on Little Explorer. Only the
	// damage is overridden, so its hop lands at the same time on both presets.
	one := 1
	heavySlime := slimeDef
	heavySlime.Damage = 2
	heavySlime.LittleExplorer = &combat.EnemyOverrides{Damage: &one}

	// The arena's only spawn, from_village at (24,40), is both where a start
	// on the arena and the warp from the village put the critter, right next
	// to the slime at (33,40).
	arena := `{"id": 1, "name": "from_village", "type": "player_spawn", "x": 192, "y": 320, "point": true},` +
		enemyObject(2, 33, 40, "sulky_slime")
	load := mapsLoader(map[string][]byte{
		"village": roomMap(6, 4, "village", villageStart+`,`+warpObject(2, 4, 2, "arena", "from_village"), ``),
		"arena":   roomMap(6, 4, "arena", arena, ``),
	})

	littleExplorer := []Option{WithPreset(combat.PresetLittleExplorer)}
	tests := []struct {
		name       string
		opts       []Option
		entry      string // "start" on the arena, "warp" into it from the village, or "nap" in it
		wantPetals int
	}{
		{name: "explorer by default, starting on the map", entry: "start", wantPetals: 1},
		{name: "explorer, starting on the map", opts: []Option{WithPreset(combat.PresetExplorer)}, entry: "start", wantPetals: 1},
		{name: "little explorer, starting on the map", opts: littleExplorer, entry: "start", wantPetals: 2},
		{name: "little explorer, after a warp into the map", opts: littleExplorer, entry: "warp", wantPetals: 2},
		{name: "little explorer, after a nap", opts: littleExplorer, entry: "nap", wantPetals: 2},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			start := "arena"
			if tc.entry == "warp" {
				start = "village"
			}
			opts := append([]Option{WithEnemies([]combat.EnemyDef{heavySlime})}, tc.opts...)
			s, err := NewScene(nil, load, start, opts...)
			if err != nil {
				t.Fatalf("NewScene: %v", err)
			}
			switch tc.entry {
			case "warp":
				updateUntilMap(t, s, "arena", walkFrame(geom.Vec{X: 1}))
			case "nap":
				standUntilNap(t, s)
			}

			// The slime's first hop lands well within 200 ticks; stop at the
			// first petal lost.
			for range 200 {
				if err := s.Update(standFrame); err != nil {
					t.Fatalf("Update: %v", err)
				}
				if s.health.Petals() < 3 {
					break
				}
			}
			if got := s.health.Petals(); got != tc.wantPetals {
				t.Errorf("after the first hit: %d petals; want %d", got, tc.wantPetals)
			}
		})
	}
}

// standUntilNap updates s with the critter standing still until it has napped
// and woken with full petals.
func standUntilNap(t *testing.T, s *Scene) {
	t.Helper()
	for range 1000 {
		before := s.health.Petals()
		if err := s.Update(standFrame); err != nil {
			t.Fatalf("Update: %v", err)
		}
		if before < 3 && s.health.Petals() == 3 {
			return
		}
	}
	t.Fatalf("no nap after 1000 ticks: %d petals, critter at %v", s.health.Petals(), s.player.Pos)
}

func TestScene_InPeacefulModeACritterStandingOnAnEnemyLosesNoPetal(t *testing.T) {
	// The arena's only spawn, from_village at (24,40), is where both a start
	// on the arena and the warp from the village put the critter, right next
	// to the slime at (33,40), whose awake hop lands within 100 ticks.
	arena := `{"id": 1, "name": "from_village", "type": "player_spawn", "x": 192, "y": 320, "point": true},` +
		enemyObject(2, 33, 40, "sulky_slime")
	load := mapsLoader(map[string][]byte{
		"village": roomMap(6, 4, "village", villageStart+`,`+warpObject(2, 4, 2, "arena", "from_village"), ``),
		"arena":   roomMap(6, 4, "arena", arena, ``),
	})

	tests := []struct {
		name     string
		peaceful bool
		start    string
		wantHurt bool
	}{
		{name: "awake, the slime's hop takes a petal", start: "arena", wantHurt: true},
		{name: "peaceful, starting on the map", peaceful: true, start: "arena"},
		{name: "peaceful, after a warp into the map", peaceful: true, start: "village"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			opts := []Option{WithEnemies([]combat.EnemyDef{slimeDef})}
			if tc.peaceful {
				opts = append(opts, WithPeaceful())
			}
			s, err := NewScene(nil, load, tc.start, opts...)
			if err != nil {
				t.Fatalf("NewScene: %v", err)
			}
			if tc.start == "village" {
				updateUntilMap(t, s, "arena", walkFrame(geom.Vec{X: 1}))
			}

			// Ten seconds of game time, many times the slime's telegraph,
			// hop and rest.
			hurt := false
			for range 600 {
				if err := s.Update(standFrame); err != nil {
					t.Fatalf("Update: %v", err)
				}
				hurt = hurt || s.health.Petals() < 3
			}
			if hurt != tc.wantHurt {
				t.Errorf("10 s on the slime: lost a petal %v; want %v", hurt, tc.wantHurt)
			}
		})
	}
}

func TestScene_OnLittleExplorerAnExplorerOnlyEnemyIsLeftOutAndAnUnmarkedOneStays(t *testing.T) {
	// The arena's only spawn, from_village at (24,40), is where a start on
	// the arena, the warp from the village and a nap put the critter, right
	// next to the explorer_only slime at (33,40). The unmarked slime at
	// (110,40) is 86 px away, too far to notice a critter on the spawn.
	marked := strings.Replace(enemyObject(2, 33, 40, "sulky_slime"), `"value": "sulky_slime"}`,
		`"value": "sulky_slime"}, {"name": "explorer_only", "type": "bool", "value": true}`, 1)
	arena := `{"id": 1, "name": "from_village", "type": "player_spawn", "x": 192, "y": 320, "point": true},` +
		marked + `,` + enemyObject(3, 110, 40, "sulky_slime")
	load := mapsLoader(map[string][]byte{
		"village": roomMap(6, 4, "village", villageStart+`,`+warpObject(2, 4, 2, "arena", "from_village"), ``),
		"arena":   roomMap(8, 5, "arena", arena, ``),
	})

	littleExplorer := []Option{WithPreset(combat.PresetLittleExplorer)}
	tests := []struct {
		name  string
		opts  []Option
		entry string // "start" on the arena, "warp" into it from the village, or "nap" in it
		// markedSpawns is whether the explorer_only slime is in the room.
		markedSpawns bool
	}{
		{name: "explorer by default, starting on the map", entry: "start", markedSpawns: true},
		{name: "explorer, starting on the map", opts: []Option{WithPreset(combat.PresetExplorer)}, entry: "start", markedSpawns: true},
		{name: "little explorer, starting on the map", opts: littleExplorer, entry: "start"},
		{name: "little explorer, after a warp into the map", opts: littleExplorer, entry: "warp"},
		{name: "little explorer, after a nap", opts: littleExplorer, entry: "nap"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			start := "arena"
			if tc.entry == "warp" {
				start = "village"
			}
			opts := append([]Option{WithEnemies([]combat.EnemyDef{slimeDef})}, tc.opts...)
			s, err := NewScene(nil, load, start, opts...)
			if err != nil {
				t.Fatalf("NewScene: %v", err)
			}
			switch tc.entry {
			case "warp":
				updateUntilMap(t, s, "arena", walkFrame(geom.Vec{X: 1}))
			case "nap":
				// Only the unmarked slime can hurt the critter here, so
				// napping shows it is in the room.
				walkEastTo(t, s, 70)
				standUntilNap(t, s)
			}
			if s.player.Pos != (geom.Vec{X: 24, Y: 40}) {
				t.Fatalf("critter at %v; want on the spawn at 24,40, next to the explorer_only slime", s.player.Pos)
			}

			// Ten seconds of game time on the explorer_only slime's spot, many
			// times its telegraph, hop and rest.
			hurt := false
			for range 600 {
				if err := s.Update(standFrame); err != nil {
					t.Fatalf("Update: %v", err)
				}
				hurt = hurt || s.health.Petals() < 3
			}
			if hurt != tc.markedSpawns {
				t.Errorf("10 s next to the explorer_only slime: lost a petal %v; want %v", hurt, tc.markedSpawns)
			}
			if tc.markedSpawns {
				return
			}

			// With the explorer_only slime left out, walking up to the
			// unmarked slime still costs a petal.
			walkEastTo(t, s, 70)
			for range 300 {
				if s.health.Petals() < 3 {
					break
				}
				if err := s.Update(standFrame); err != nil {
					t.Fatalf("Update: %v", err)
				}
			}
			if got := s.health.Petals(); got != 2 {
				t.Errorf("next to the unmarked slime: %d petals; want 2, it is in the room and hits", got)
			}
		})
	}
}

// walkEastTo updates s with the critter walking east until it is at x or
// further.
func walkEastTo(t *testing.T, s *Scene, x float64) {
	t.Helper()
	for range 300 {
		if s.player.Pos.X >= x {
			return
		}
		if err := s.Update(walkFrame(geom.Vec{X: 1})); err != nil {
			t.Fatalf("Update: %v", err)
		}
	}
	t.Fatalf("critter at %v after 300 ticks walking east; want at x %v or further", s.player.Pos, x)
}
