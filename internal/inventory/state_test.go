package inventory_test

import (
	"encoding/json"
	"errors"
	"maps"
	"reflect"
	"slices"
	"testing"

	"github.com/Elpulgo/critterkingdom/internal/inventory"
)

func TestInventory_SavedInventoryRestoredFromJSONHoldsTheSameThings(t *testing.T) {
	inv := newInventory(t)
	mustAdd(t, inv, "berry", 3)
	mustAdd(t, inv, "grandmas_hat", 1)
	inv.UnlockSticker("hat_hero")
	inv.UnlockSticker("frog")
	inv.SetCompanion("firefly")
	inv.GrantWeapon("stick")
	inv.GrantWeapon("sturdy_stick")

	data, err := json.Marshal(inv.Snapshot())
	if err != nil {
		t.Fatalf("json.Marshal: %v", err)
	}
	var keys map[string]json.RawMessage
	if err := json.Unmarshal(data, &keys); err != nil {
		t.Fatalf("json.Unmarshal into keys: %v", err)
	}
	if got, want := slices.Sorted(maps.Keys(keys)), []string{"companion", "items", "stickers", "weapons"}; !slices.Equal(got, want) {
		t.Fatalf("saved JSON keys = %v; want %v", got, want)
	}
	var state inventory.State
	if err := json.Unmarshal(data, &state); err != nil {
		t.Fatalf("json.Unmarshal: %v", err)
	}
	restored, err := inventory.Restore(newCatalog(t), state)
	if err != nil {
		t.Fatalf("Restore: %v", err)
	}

	if got, want := held(restored.Items()), []string{"berry×3", "grandmas_hat×1"}; !slices.Equal(got, want) {
		t.Fatalf("restored Items() = %v; want %v", got, want)
	}
	if got := restored.Count("berry"); got != 3 {
		t.Fatalf("restored Count(berry) = %d; want 3", got)
	}
	if got, want := held(restored.KeyItems()), []string{"grandmas_hat×1"}; !slices.Equal(got, want) {
		t.Fatalf("restored KeyItems() = %v; want %v", got, want)
	}
	if got, want := restored.Stickers(), []string{"hat_hero", "frog"}; !slices.Equal(got, want) {
		t.Fatalf("restored Stickers() = %v; want %v", got, want)
	}
	if id, ok := restored.Companion(); id != "firefly" || !ok {
		t.Fatalf("restored Companion() = %q, %v; want \"firefly\", true", id, ok)
	}
	if id, ok := restored.Weapon(); id != "sturdy_stick" || !ok {
		t.Fatalf("restored Weapon() = %q, %v; want \"sturdy_stick\", true", id, ok)
	}
	if got, want := restored.Weapons(), []string{"stick", "sturdy_stick"}; !slices.Equal(got, want) {
		t.Fatalf("restored Weapons() = %v; want %v", got, want)
	}
}

func TestInventory_SnapshotAndRestoredInventoryDoNotShareState(t *testing.T) {
	inv := newInventory(t)
	mustAdd(t, inv, "berry", 2)
	inv.UnlockSticker("frog")
	inv.GrantWeapon("stick")
	snap := inv.Snapshot()

	mustAdd(t, inv, "berry", 5)
	inv.UnlockSticker("hat_hero")
	inv.GrantWeapon("sturdy_stick")
	inv.SetCompanion("firefly")

	want := inventory.State{Items: map[string]int{"berry": 2}, Stickers: []string{"frog"}, Weapons: []string{"stick"}}
	if !reflect.DeepEqual(snap, want) {
		t.Fatalf("snapshot after later changes = %+v; want %+v", snap, want)
	}

	restored, err := inventory.Restore(newCatalog(t), snap)
	if err != nil {
		t.Fatalf("Restore: %v", err)
	}
	snap.Items["berry"] = 9
	snap.Stickers[0] = "owl"
	snap.Weapons[0] = "sparkle_wand"
	if got := restored.Count("berry"); got != 2 {
		t.Fatalf("restored Count(berry) after changing the state = %d; want 2", got)
	}
	if got, want := restored.Stickers(), []string{"frog"}; !slices.Equal(got, want) {
		t.Fatalf("restored Stickers() after changing the state = %v; want %v", got, want)
	}
	if id, _ := restored.Weapon(); id != "stick" {
		t.Fatalf("restored Weapon() after changing the state = %q; want \"stick\"", id)
	}
}

func TestRestore_SaveWithUnknownItemOrNegativeCountIsRefused(t *testing.T) {
	tests := []struct {
		name  string
		items map[string]int
		want  error
	}{
		{"unknown item", map[string]int{"berry": 1, "golden_acorn": 1}, inventory.ErrUnknownItem},
		{"negative count", map[string]int{"berry": 1, "cookie": -1}, inventory.ErrInvalidCount},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			inv, err := inventory.Restore(newCatalog(t), inventory.State{Items: tc.items})
			if !errors.Is(err, tc.want) {
				t.Fatalf("Restore(%v) error = %v; want %v", tc.items, err, tc.want)
			}
			if inv != nil {
				t.Fatal("Restore returned an inventory along with the error; want nil")
			}
		})
	}
}

func TestRestore_DropsZeroCountsAndRepeatedIds(t *testing.T) {
	state := inventory.State{
		Items:    map[string]int{"berry": 0, "cookie": 1},
		Stickers: []string{"frog", "hat_hero", "frog"},
		Weapons:  []string{"stick", "sturdy_stick", "stick"},
	}

	inv, err := inventory.Restore(newCatalog(t), state)
	if err != nil {
		t.Fatalf("Restore: %v", err)
	}

	want := inventory.State{
		Items:    map[string]int{"cookie": 1},
		Stickers: []string{"frog", "hat_hero"},
		Weapons:  []string{"stick", "sturdy_stick"},
	}
	if got := inv.Snapshot(); !reflect.DeepEqual(got, want) {
		t.Fatalf("Snapshot() of restored save = %+v; want %+v", got, want)
	}
	if id, _ := inv.Weapon(); id != "sturdy_stick" {
		t.Fatalf("Weapon() = %q; want \"sturdy_stick\"", id)
	}
}
