package inventory_test

import (
	"errors"
	"fmt"
	"slices"
	"testing"

	"github.com/Elpulgo/critterkingdom/internal/inventory"
)

const testCatalog = `[
	{"id": "berry", "name": {"en": "Berry", "sv": "Bär"}, "icon": "berry"},
	{"id": "cookie", "name": {"en": "Cookie", "sv": "Kaka"}, "icon": "cookie"},
	{"id": "grandmas_hat", "name": {"en": "Grandma's Hat", "sv": "Mormors hatt"}, "icon": "hat", "key": true}
]`

func newCatalog(t *testing.T) *inventory.Catalog {
	t.Helper()
	c, err := inventory.ParseCatalog([]byte(testCatalog))
	if err != nil {
		t.Fatalf("ParseCatalog: %v", err)
	}
	return c
}

func newInventory(t *testing.T) *inventory.Inventory {
	t.Helper()
	return inventory.New(newCatalog(t))
}

func mustAdd(t *testing.T, inv *inventory.Inventory, id string, n int) {
	t.Helper()
	if err := inv.Add(id, n); err != nil {
		t.Fatalf("Add(%q, %d): %v", id, n, err)
	}
}

// held renders stacks as "id×count" so they compare against literals.
func held(stacks []inventory.Stack) []string {
	var out []string
	for _, s := range stacks {
		out = append(out, fmt.Sprintf("%s×%d", s.Item.ID, s.Count))
	}
	return out
}

func TestInventory_DeliveringMoreThanYouHaveIsRefused(t *testing.T) {
	inv := newInventory(t)
	mustAdd(t, inv, "berry", 2)

	err := inv.Remove("berry", 3)

	if !errors.Is(err, inventory.ErrNotEnough) {
		t.Fatalf("Remove(berry, 3) holding 2 = %v; want ErrNotEnough", err)
	}
	if got := inv.Count("berry"); got != 2 {
		t.Fatalf("Count(berry) after refused delivery = %d; want 2", got)
	}
}

func TestInventory_CollectedItemsAddUpAndDeliveryTakesExactlyWhatWasDelivered(t *testing.T) {
	inv := newInventory(t)
	mustAdd(t, inv, "cookie", 1)
	mustAdd(t, inv, "berry", 1)
	mustAdd(t, inv, "berry", 2)

	if err := inv.Remove("berry", 2); err != nil {
		t.Fatalf("Remove(berry, 2): %v", err)
	}
	if got := inv.Count("berry"); got != 1 {
		t.Fatalf("Count(berry) = %d; want 1", got)
	}
	if !inv.Has("berry", 1) || inv.Has("berry", 2) {
		t.Fatalf("Has(berry, 1), Has(berry, 2) = %v, %v; want true, false", inv.Has("berry", 1), inv.Has("berry", 2))
	}
	if got, want := held(inv.Items()), []string{"berry×1", "cookie×1"}; !slices.Equal(got, want) {
		t.Fatalf("Items() = %v; want %v in catalog order", got, want)
	}

	if err := inv.Remove("cookie", 1); err != nil {
		t.Fatalf("Remove(cookie, 1): %v", err)
	}
	if got, want := held(inv.Items()), []string{"berry×1"}; !slices.Equal(got, want) {
		t.Fatalf("Items() after delivering the last cookie = %v; want %v", got, want)
	}
}

func TestInventory_UnknownItemsAndNonPositiveCountsAreRefused(t *testing.T) {
	tests := []struct {
		name string
		op   func(*inventory.Inventory) error
		want error
	}{
		{"add unknown item", func(inv *inventory.Inventory) error { return inv.Add("nope", 1) }, inventory.ErrUnknownItem},
		{"remove unknown item", func(inv *inventory.Inventory) error { return inv.Remove("nope", 1) }, inventory.ErrUnknownItem},
		{"add zero", func(inv *inventory.Inventory) error { return inv.Add("berry", 0) }, inventory.ErrInvalidCount},
		{"add negative", func(inv *inventory.Inventory) error { return inv.Add("berry", -1) }, inventory.ErrInvalidCount},
		{"remove zero", func(inv *inventory.Inventory) error { return inv.Remove("berry", 0) }, inventory.ErrInvalidCount},
		{"remove negative", func(inv *inventory.Inventory) error { return inv.Remove("berry", -1) }, inventory.ErrInvalidCount},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			inv := newInventory(t)
			mustAdd(t, inv, "berry", 2)

			if err := tc.op(inv); !errors.Is(err, tc.want) {
				t.Fatalf("err = %v; want %v", err, tc.want)
			}
			if got, want := held(inv.Items()), []string{"berry×2"}; !slices.Equal(got, want) {
				t.Fatalf("Items() after refusal = %v; want %v", got, want)
			}
		})
	}
}

func TestInventory_KeyItemsAreListedApartOnlyWhileHeldAndCanBeDelivered(t *testing.T) {
	inv := newInventory(t)
	mustAdd(t, inv, "berry", 1)
	if got := inv.KeyItems(); len(got) != 0 {
		t.Fatalf("KeyItems() holding only a berry = %v; want none", held(got))
	}

	mustAdd(t, inv, "grandmas_hat", 2)
	if got, want := held(inv.KeyItems()), []string{"grandmas_hat×2"}; !slices.Equal(got, want) {
		t.Fatalf("KeyItems() = %v; want %v", got, want)
	}

	if err := inv.Remove("grandmas_hat", 2); err != nil {
		t.Fatalf("Remove(grandmas_hat, 2): %v", err)
	}
	if got := inv.KeyItems(); len(got) != 0 {
		t.Fatalf("KeyItems() after delivering the hat = %v; want none", held(got))
	}
}
