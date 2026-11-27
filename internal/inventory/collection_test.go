package inventory_test

import (
	"slices"
	"testing"
)

func TestInventory_NewestWeaponWinsAndRegrantingAnOwnedWeaponKeepsIt(t *testing.T) {
	inv := newInventory(t)
	if id, ok := inv.Weapon(); id != "" || ok {
		t.Fatalf("Weapon() with no weapons = %q, %v; want \"\", false", id, ok)
	}

	inv.GrantWeapon("stick")
	inv.GrantWeapon("sturdy_stick")
	if id, ok := inv.Weapon(); id != "sturdy_stick" || !ok {
		t.Fatalf("Weapon() after stick, sturdy_stick = %q, %v; want \"sturdy_stick\", true", id, ok)
	}

	inv.GrantWeapon("stick")
	if id, _ := inv.Weapon(); id != "sturdy_stick" {
		t.Fatalf("Weapon() after stick is given again = %q; want \"sturdy_stick\"", id)
	}
	if got, want := inv.Weapons(), []string{"stick", "sturdy_stick"}; !slices.Equal(got, want) {
		t.Fatalf("Weapons() = %v; want %v, oldest first", got, want)
	}
	if !inv.HasWeapon("stick") || inv.HasWeapon("sparkle_wand") {
		t.Fatalf("HasWeapon(stick), HasWeapon(sparkle_wand) = %v, %v; want true, false", inv.HasWeapon("stick"), inv.HasWeapon("sparkle_wand"))
	}
}

func TestInventory_StickerIsNewOnlyTheFirstTimeItIsUnlocked(t *testing.T) {
	inv := newInventory(t)
	if inv.HasSticker("hat_hero") {
		t.Fatal("HasSticker(hat_hero) on a new inventory = true; want false")
	}

	if !inv.UnlockSticker("hat_hero") {
		t.Fatal("first UnlockSticker(hat_hero) = false; want true")
	}
	if !inv.UnlockSticker("frog") {
		t.Fatal("first UnlockSticker(frog) = false; want true")
	}
	if inv.UnlockSticker("hat_hero") {
		t.Fatal("second UnlockSticker(hat_hero) = true; want false")
	}

	if !inv.HasSticker("hat_hero") {
		t.Fatal("HasSticker(hat_hero) after unlocking = false; want true")
	}
	if got, want := inv.Stickers(), []string{"hat_hero", "frog"}; !slices.Equal(got, want) {
		t.Fatalf("Stickers() = %v; want %v in unlock order", got, want)
	}
}

func TestInventory_ChoosingANewCompanionReplacesThePreviousOne(t *testing.T) {
	inv := newInventory(t)
	if id, ok := inv.Companion(); id != "" || ok {
		t.Fatalf("Companion() on a new inventory = %q, %v; want \"\", false", id, ok)
	}

	inv.SetCompanion("firefly")
	if id, ok := inv.Companion(); id != "firefly" || !ok {
		t.Fatalf("Companion() = %q, %v; want \"firefly\", true", id, ok)
	}

	inv.SetCompanion("frog")
	if id, ok := inv.Companion(); id != "frog" || !ok {
		t.Fatalf("Companion() after choosing the frog = %q, %v; want \"frog\", true", id, ok)
	}

	inv.ClearCompanion()
	if id, ok := inv.Companion(); id != "" || ok {
		t.Fatalf("Companion() after ClearCompanion = %q, %v; want \"\", false", id, ok)
	}
}
