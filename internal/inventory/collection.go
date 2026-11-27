package inventory

import "slices"

// Stickers, companion and weapons are opaque ids: the inventory does not check
// them against a catalog. Sticker ids come from critter, enemy and quest data,
// and weapon stats belong to combat.

// UnlockSticker adds the sticker id to the sticker book. It returns true only
// when the sticker is new, so callers can fire the fanfare and autosave once.
func (inv *Inventory) UnlockSticker(id string) bool {
	if inv.HasSticker(id) {
		return false
	}
	inv.stickers = append(inv.stickers, id)
	return true
}

// HasSticker reports whether the sticker id is unlocked.
func (inv *Inventory) HasSticker(id string) bool {
	return slices.Contains(inv.stickers, id)
}

// Stickers returns the unlocked sticker ids in unlock order.
func (inv *Inventory) Stickers() []string {
	return slices.Clone(inv.stickers)
}

// SetCompanion makes the critter id the companion, replacing any earlier one.
// Whether the critter is befriended is for the caller to decide.
func (inv *Inventory) SetCompanion(id string) {
	inv.companion = id
}

// ClearCompanion leaves the player without a companion.
func (inv *Inventory) ClearCompanion() {
	inv.companion = ""
}

// Companion returns the chosen companion's id, and false when there is none.
func (inv *Inventory) Companion() (string, bool) {
	return inv.companion, inv.companion != ""
}

// GrantWeapon gives the player the weapon id. A newly owned weapon becomes
// the one in use; a weapon already owned changes nothing, so a repeated reward
// never swaps back to an older weapon.
func (inv *Inventory) GrantWeapon(id string) {
	if inv.HasWeapon(id) {
		return
	}
	inv.weapons = append(inv.weapons, id)
}

// HasWeapon reports whether the player owns the weapon id.
func (inv *Inventory) HasWeapon(id string) bool {
	return slices.Contains(inv.weapons, id)
}

// Weapons returns the owned weapon ids, oldest first.
func (inv *Inventory) Weapons() []string {
	return slices.Clone(inv.weapons)
}

// Weapon returns the weapon in use, the newest one owned, and false when the
// player owns none.
func (inv *Inventory) Weapon() (string, bool) {
	if len(inv.weapons) == 0 {
		return "", false
	}
	return inv.weapons[len(inv.weapons)-1], true
}
