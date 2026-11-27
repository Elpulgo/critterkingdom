package inventory

import (
	"errors"
	"fmt"
)

// Errors returned by Add and Remove. Check them with errors.Is.
var (
	// ErrUnknownItem means the item id is not in the catalog.
	ErrUnknownItem = errors.New("unknown item")
	// ErrInvalidCount means the count to add or remove is below 1.
	ErrInvalidCount = errors.New("count must be at least 1")
	// ErrNotEnough means a removal asked for more than the player holds.
	ErrNotEnough = errors.New("not enough items")
)

// Stack is a held item and how many of it the player carries.
type Stack struct {
	Item  Item
	Count int
}

// Inventory is what the player carries. It is not safe for concurrent use.
type Inventory struct {
	catalog *Catalog
	counts  map[string]int
	// stickers holds unlocked sticker ids in unlock order.
	stickers []string
	// companion is the chosen companion's id; "" means none.
	companion string
	// weapons holds owned weapon ids, oldest first; the last is in use.
	weapons []string
}

// New returns an empty inventory for the items in catalog.
func New(catalog *Catalog) *Inventory {
	return &Inventory{catalog: catalog, counts: make(map[string]int)}
}

// Add puts n of the item id into the inventory. It refuses an id not in the
// catalog or n below 1 and then changes nothing.
func (inv *Inventory) Add(id string, n int) error {
	if err := inv.check(id, n); err != nil {
		return fmt.Errorf("add %d %q: %w", n, id, err)
	}
	inv.counts[id] += n
	return nil
}

// Remove takes n of the item id out of the inventory, for example to deliver
// it to a quest. It refuses an id not in the catalog, n below 1, or n above
// what is held, and then changes nothing: there is no partial removal.
func (inv *Inventory) Remove(id string, n int) error {
	if err := inv.check(id, n); err != nil {
		return fmt.Errorf("remove %d %q: %w", n, id, err)
	}
	have := inv.counts[id]
	if n > have {
		return fmt.Errorf("remove %d %q, holding %d: %w", n, id, have, ErrNotEnough)
	}
	if have == n {
		delete(inv.counts, id)
	} else {
		inv.counts[id] = have - n
	}
	return nil
}

func (inv *Inventory) check(id string, n int) error {
	if _, ok := inv.catalog.Item(id); !ok {
		return ErrUnknownItem
	}
	if n < 1 {
		return ErrInvalidCount
	}
	return nil
}

// Count returns how many of the item id the player holds; 0 for an item not
// held or not in the catalog.
func (inv *Inventory) Count(id string) int {
	return inv.counts[id]
}

// Has reports whether the player holds at least n of the item id.
func (inv *Inventory) Has(id string, n int) bool {
	return inv.counts[id] >= n
}

// Items returns every held item in catalog order, without items the player
// holds none of.
func (inv *Inventory) Items() []Stack {
	return inv.stacks(func(Item) bool { return true })
}

// KeyItems returns the held items whose catalog entry is a key item, in
// catalog order.
func (inv *Inventory) KeyItems() []Stack {
	return inv.stacks(func(it Item) bool { return it.Key })
}

func (inv *Inventory) stacks(keep func(Item) bool) []Stack {
	var out []Stack
	for _, it := range inv.catalog.items {
		if n := inv.counts[it.ID]; n > 0 && keep(it) {
			out = append(out, Stack{Item: it, Count: n})
		}
	}
	return out
}
