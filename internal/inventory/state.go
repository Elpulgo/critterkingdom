package inventory

import (
	"fmt"
	"maps"
	"slices"
)

// State is everything in an inventory that a save file keeps. It is a plain
// value that the save layer can embed, marshal and version.
type State struct {
	// Items maps held item ids to their counts.
	Items map[string]int `json:"items"`
	// Stickers lists unlocked sticker ids in unlock order.
	Stickers []string `json:"stickers"`
	// Companion is the chosen companion's id; "" means none.
	Companion string `json:"companion"`
	// Weapons lists owned weapon ids, oldest first; the last is in use.
	Weapons []string `json:"weapons"`
}

// Snapshot returns the inventory's state as a copy: later changes to the
// inventory do not show in it.
func (inv *Inventory) Snapshot() State {
	return State{
		Items:     maps.Clone(inv.counts),
		Stickers:  slices.Clone(inv.stickers),
		Companion: inv.companion,
		Weapons:   slices.Clone(inv.weapons),
	}
}

// Restore returns an inventory for catalog holding what state describes. It
// refuses an item id not in the catalog or a negative count, so corrupt or
// outdated saves reach the save layer as errors. Zero counts are dropped, and
// repeated sticker or weapon ids keep only their first occurrence. The
// inventory copies state, so later changes to state do not show in it.
func Restore(catalog *Catalog, state State) (*Inventory, error) {
	inv := New(catalog)
	for _, id := range slices.Sorted(maps.Keys(state.Items)) {
		n := state.Items[id]
		if _, ok := catalog.Item(id); !ok {
			return nil, fmt.Errorf("restore %d %q: %w", n, id, ErrUnknownItem)
		}
		if n < 0 {
			return nil, fmt.Errorf("restore %d %q: negative count: %w", n, id, ErrInvalidCount)
		}
		if n > 0 {
			inv.counts[id] = n
		}
	}
	for _, id := range state.Stickers {
		inv.UnlockSticker(id)
	}
	inv.companion = state.Companion
	for _, id := range state.Weapons {
		inv.GrantWeapon(id)
	}
	return inv, nil
}
