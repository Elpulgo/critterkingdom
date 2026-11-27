// Package inventory keeps what the player carries: items, stickers, the
// chosen companion and weapons. It is pure Go and only parses bytes; reading
// the embedded data files lives in internal/assets.
package inventory

import (
	"encoding/json"
	"errors"
	"fmt"
)

// Name is an item's display name in each supported language.
type Name struct {
	EN string `json:"en"`
	SV string `json:"sv"`
}

// Item is one entry of the item catalog.
type Item struct {
	ID   string `json:"id"`
	Name Name   `json:"name"`
	// Icon names the image the UI draws for the item, such as "berry".
	Icon string `json:"icon"`
	// Key marks a key item, such as a quest item.
	Key bool `json:"key"`
}

// Catalog is the list of every item the game knows, in data file order.
type Catalog struct {
	items []Item
	index map[string]int
}

// ParseCatalog parses an items.json file: a JSON array of items. It rejects
// an item with an empty or duplicate id, a missing English or Swedish name,
// or a missing icon.
func ParseCatalog(data []byte) (*Catalog, error) {
	var items []Item
	if err := json.Unmarshal(data, &items); err != nil {
		return nil, fmt.Errorf("parse item catalog: %w", err)
	}
	c := &Catalog{items: items, index: make(map[string]int, len(items))}
	for i, it := range items {
		if err := validate(it); err != nil {
			return nil, fmt.Errorf("item %d %q: %w", i, it.ID, err)
		}
		if _, dup := c.index[it.ID]; dup {
			return nil, fmt.Errorf("item %d %q: duplicate id", i, it.ID)
		}
		c.index[it.ID] = i
	}
	return c, nil
}

func validate(it Item) error {
	switch {
	case it.ID == "":
		return errors.New("empty id")
	case it.Name.EN == "":
		return errors.New("missing en name")
	case it.Name.SV == "":
		return errors.New("missing sv name")
	case it.Icon == "":
		return errors.New("missing icon")
	}
	return nil
}

// Item returns the catalog entry with the given id, and false when the
// catalog has no such item.
func (c *Catalog) Item(id string) (Item, bool) {
	i, ok := c.index[id]
	if !ok {
		return Item{}, false
	}
	return c.items[i], true
}

// Items returns every item in catalog order. The slice is a copy.
func (c *Catalog) Items() []Item {
	return append([]Item(nil), c.items...)
}
