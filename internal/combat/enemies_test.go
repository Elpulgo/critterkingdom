package combat_test

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"github.com/Elpulgo/critterkingdom/internal/combat"
)

// enemyJSON returns a valid thistle_beast definition with edit applied, as
// enemies.json bytes holding a valid sulky_slime first and the edited enemy
// second.
func enemyJSON(t *testing.T, edit func(e map[string]any)) []byte {
	t.Helper()
	slime := map[string]any{
		"id": "sulky_slime", "behaviour": "hop_toward",
		"hp": 1, "damage": 1, "speed": 30, "telegraphSeconds": 0.6,
	}
	beast := map[string]any{
		"id":               "thistle_beast",
		"name":             map[string]any{"en": "Thistle Beast", "sv": "Tistelbest"},
		"sprite":           "enemies/thistle_beast",
		"behaviour":        "charge",
		"hp":               2,
		"damage":           1,
		"speed":            60,
		"telegraphSeconds": 0.8,
		"drops":            []any{map[string]any{"item": "heart", "chance": 0.5}},
		"littleExplorer":   map[string]any{"hp": 1, "damage": 1, "speed": 40},
	}
	edit(beast)
	b, err := json.Marshal([]any{slime, beast})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	return b
}

func TestParseEnemies_RejectsABrokenDefinitionNamingTheEnemy(t *testing.T) {
	override := func(key string, v any) func(map[string]any) {
		return func(e map[string]any) { e["littleExplorer"].(map[string]any)[key] = v }
	}
	tests := []struct {
		name string
		edit func(e map[string]any)
		// wantInErr are the words the error must hold: who and what.
		wantInErr []string
	}{
		{"missing id is named by its position", func(e map[string]any) { delete(e, "id") }, []string{"enemy 1", "missing id"}},
		{"duplicate id", func(e map[string]any) { e["id"] = "sulky_slime" }, []string{`"sulky_slime"`, "duplicate id"}},
		{"unknown behaviour", func(e map[string]any) { e["behaviour"] = "fly" }, []string{`"thistle_beast"`, "behaviour", `"fly"`}},
		{"missing behaviour", func(e map[string]any) { delete(e, "behaviour") }, []string{`"thistle_beast"`, "behaviour"}},
		{"zero hp", func(e map[string]any) { e["hp"] = 0 }, []string{`"thistle_beast"`, "hp"}},
		{"negative damage", func(e map[string]any) { e["damage"] = -1 }, []string{`"thistle_beast"`, "damage"}},
		{"zero speed for a charge", func(e map[string]any) { e["speed"] = 0 }, []string{`"thistle_beast"`, "speed"}},
		{"negative speed for a puff", func(e map[string]any) { e["behaviour"] = "puff"; e["speed"] = -5 }, []string{`"thistle_beast"`, "speed"}},
		{"zero spore speed for a puff", func(e map[string]any) { e["behaviour"] = "puff"; e["speed"] = 0 }, []string{`"thistle_beast"`, "speed"}},
		{"telegraph of 0.59 seconds", func(e map[string]any) { e["telegraphSeconds"] = 0.59 }, []string{`"thistle_beast"`, "telegraph"}},
		{"missing telegraph", func(e map[string]any) { delete(e, "telegraphSeconds") }, []string{`"thistle_beast"`, "telegraph"}},
		{"drop with an empty item", func(e map[string]any) { e["drops"] = []any{map[string]any{"item": "", "chance": 0.5}} }, []string{`"thistle_beast"`, "drop", "item"}},
		{"drop with a chance of 0", func(e map[string]any) { e["drops"] = []any{map[string]any{"item": "heart", "chance": 0}} }, []string{`"thistle_beast"`, "drop", "chance"}},
		{"drop with a chance above 1", func(e map[string]any) { e["drops"] = []any{map[string]any{"item": "heart", "chance": 1.01}} }, []string{`"thistle_beast"`, "drop", "chance"}},
		{"little explorer hp of 0", override("hp", 0), []string{`"thistle_beast"`, "littleExplorer", "hp"}},
		{"little explorer damage of -1", override("damage", -1), []string{`"thistle_beast"`, "littleExplorer", "damage"}},
		{"little explorer speed of 0 for a charge", override("speed", 0), []string{`"thistle_beast"`, "littleExplorer", "speed"}},
		{"little explorer spore speed of 0 for a puff", func(e map[string]any) {
			e["behaviour"] = "puff"
			e["littleExplorer"] = map[string]any{"speed": 0}
		}, []string{`"thistle_beast"`, "littleExplorer", "speed"}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			defs, err := combat.ParseEnemies(enemyJSON(t, tc.edit))
			if err == nil {
				t.Fatalf("ParseEnemies = %+v, nil; want an error", defs)
			}
			for _, w := range tc.wantInErr {
				if !strings.Contains(err.Error(), w) {
					t.Errorf("error %q does not contain %q", err, w)
				}
			}
		})
	}
}

func TestParseEnemies_AcceptsValuesOnTheEdgeOfTheRules(t *testing.T) {
	tests := []struct {
		name string
		edit func(e map[string]any)
	}{
		{"telegraph of exactly 0.6 seconds", func(e map[string]any) { e["telegraphSeconds"] = 0.6 }},
		{"a drop that always happens has chance 1", func(e map[string]any) { e["drops"] = []any{map[string]any{"item": "heart", "chance": 1}} }},
		{"no little explorer overrides", func(e map[string]any) { delete(e, "littleExplorer") }},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			defs, err := combat.ParseEnemies(enemyJSON(t, tc.edit))
			if err != nil {
				t.Fatalf("ParseEnemies: %v", err)
			}
			if len(defs) != 2 || defs[1].ID != "thistle_beast" {
				t.Errorf("ParseEnemies = %+v; want sulky_slime and thistle_beast", defs)
			}
		})
	}
}

func TestEnemyDef_ForPreset_LittleExplorerAppliesOnlyTheHPDamageAndSpeedOverrides(t *testing.T) {
	intp := func(v int) *int { return &v }
	floatp := func(v float64) *float64 { return &v }
	base := func(o *combat.EnemyOverrides) combat.EnemyDef {
		return combat.EnemyDef{
			ID:               "thistle_beast",
			Name:             combat.EnemyName{EN: "Thistle Beast", SV: "Tistelbest"},
			Sprite:           "enemies/thistle_beast",
			Behaviour:        combat.BehaviourCharge,
			HP:               3,
			Damage:           2,
			Speed:            60,
			TelegraphSeconds: 0.8,
			Drops:            []combat.Drop{{Item: "heart", Chance: 0.5}, {Item: "berry", Chance: 0.3}},
			LittleExplorer:   o,
		}
	}
	full := &combat.EnemyOverrides{HP: intp(1), Damage: intp(1), Speed: floatp(40)}
	tests := []struct {
		name      string
		overrides *combat.EnemyOverrides
		preset    combat.Preset
		// wantHP, wantDamage and wantSpeed are the stats the result must have.
		wantHP, wantDamage int
		wantSpeed          float64
	}{
		{"explorer keeps the base values even with overrides", full, combat.PresetExplorer, 3, 2, 60},
		{"explorer keeps the base values without overrides", nil, combat.PresetExplorer, 3, 2, 60},
		{"little explorer applies every override", full, combat.PresetLittleExplorer, 1, 1, 40},
		{"little explorer applies only the hp override", &combat.EnemyOverrides{HP: intp(1)}, combat.PresetLittleExplorer, 1, 2, 60},
		{"little explorer applies only the damage override", &combat.EnemyOverrides{Damage: intp(1)}, combat.PresetLittleExplorer, 3, 1, 60},
		{"little explorer applies only the speed override", &combat.EnemyOverrides{Speed: floatp(40)}, combat.PresetLittleExplorer, 3, 2, 40},
		{"little explorer with an empty override block keeps the base values", &combat.EnemyOverrides{}, combat.PresetLittleExplorer, 3, 2, 60},
		{"little explorer without an override block keeps the base values", nil, combat.PresetLittleExplorer, 3, 2, 60},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			d := base(tc.overrides)

			got := d.ForPreset(tc.preset)

			want := base(tc.overrides)
			want.HP, want.Damage, want.Speed = tc.wantHP, tc.wantDamage, tc.wantSpeed
			if !reflect.DeepEqual(got, want) {
				t.Errorf("ForPreset = %+v;\nwant %+v", got, want)
			}
			if !reflect.DeepEqual(d, base(tc.overrides)) {
				t.Errorf("ForPreset changed its receiver to %+v", d)
			}
		})
	}
}

func TestParsePreset_AcceptsTheTwoPresetNamesAndRejectsOthers(t *testing.T) {
	tests := []struct {
		name    string
		want    combat.Preset
		wantErr bool
	}{
		{"explorer", combat.PresetExplorer, false},
		{"little-explorer", combat.PresetLittleExplorer, false},
		{"Little Explorer", combat.PresetExplorer, true},
		{"", combat.PresetExplorer, true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := combat.ParsePreset(tc.name)
			if got != tc.want || (err != nil) != tc.wantErr {
				t.Errorf("ParsePreset(%q) = %v, %v; want %v, error %v", tc.name, got, err, tc.want, tc.wantErr)
			}
		})
	}
}
