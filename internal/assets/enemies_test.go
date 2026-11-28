package assets_test

import (
	"testing"

	"github.com/Elpulgo/critterkingdom/internal/assets"
	"github.com/Elpulgo/critterkingdom/internal/combat"
)

func TestLoadEnemies_ShipsTheSlimeTheThistleBeastAndTheMushroom(t *testing.T) {
	defs, err := assets.LoadEnemies()
	if err != nil {
		t.Fatalf("LoadEnemies: %v", err)
	}

	type want struct {
		id, behaviour string
		hp            int
	}
	wants := []want{
		{"sulky_slime", "hop_toward", 1},
		{"thistle_beast", "charge", 2},
		{"grumpy_mushroom", "puff", 1},
	}
	if len(defs) != len(wants) {
		t.Fatalf("LoadEnemies returned %d enemies; want %d", len(defs), len(wants))
	}
	for i, w := range wants {
		d := defs[i]
		if d.ID != w.id || d.Behaviour != w.behaviour || d.HP != w.hp {
			t.Errorf("enemy %d = %q %q hp %d; want %q %q hp %d", i, d.ID, d.Behaviour, d.HP, w.id, w.behaviour, w.hp)
		}
		if d.Damage != 1 {
			t.Errorf("%s damage = %d; want 1", d.ID, d.Damage)
		}
		if d.Speed <= 0 {
			t.Errorf("%s speed = %v; want a speed", d.ID, d.Speed)
		}
		if d.TelegraphSeconds < 0.6 {
			t.Errorf("%s telegraphSeconds = %v; want at least 0.6", d.ID, d.TelegraphSeconds)
		}
		if len(d.Drops) == 0 {
			t.Errorf("%s has no drops", d.ID)
		}
		if d.LittleExplorer == nil {
			t.Errorf("%s has no littleExplorer override", d.ID)
		}
	}
}

func TestLoadEnemies_OnLittleExplorerEveryEnemyHasOneHPDealsOnePetalAndIsSlower(t *testing.T) {
	defs, err := assets.LoadEnemies()
	if err != nil {
		t.Fatalf("LoadEnemies: %v", err)
	}
	for _, d := range defs {
		le := d.ForPreset(combat.PresetLittleExplorer)
		if le.HP != 1 {
			t.Errorf("%s hp on Little Explorer = %d; want 1", d.ID, le.HP)
		}
		if le.Damage != 1 {
			t.Errorf("%s damage on Little Explorer = %d; want 1", d.ID, le.Damage)
		}
		if le.Speed >= d.Speed {
			t.Errorf("%s speed on Little Explorer = %v; want below its base speed %v", d.ID, le.Speed, d.Speed)
		}
	}
}
