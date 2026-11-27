package assets_test

import (
	"testing"

	"github.com/Elpulgo/critterkingdom/internal/assets"
	"github.com/Elpulgo/critterkingdom/internal/inventory"
)

func TestLoadItems_ShippedCatalogNamesEveryItemInBothLanguagesWithAnIcon(t *testing.T) {
	c, err := assets.LoadItems()
	if err != nil {
		t.Fatalf("LoadItems: %v", err)
	}

	var ids []string
	for _, it := range c.Items() {
		ids = append(ids, it.ID)
		if it.Name.EN == "" || it.Name.SV == "" || it.Icon == "" {
			t.Errorf("item %q = %+v; want en and sv names and an icon", it.ID, it)
		}
	}
	want := []string{"berry", "cookie", "grandmas_hat"}
	if len(ids) != len(want) {
		t.Fatalf("item ids = %v; want %v", ids, want)
	}
	for i := range want {
		if ids[i] != want[i] {
			t.Fatalf("item ids = %v; want %v", ids, want)
		}
	}

	berry, ok := c.Item("berry")
	wantBerry := inventory.Item{ID: "berry", Name: inventory.Name{EN: "Berry", SV: "Bär"}, Icon: "berry"}
	if !ok || berry != wantBerry {
		t.Errorf("Item(berry) = %+v, %v; want %+v, true", berry, ok, wantBerry)
	}

	hat, ok := c.Item("grandmas_hat")
	wantHat := inventory.Item{ID: "grandmas_hat", Name: inventory.Name{EN: "Grandma's Hat", SV: "Mormors hatt"}, Icon: "hat", Key: true}
	if !ok || hat != wantHat {
		t.Errorf("Item(grandmas_hat) = %+v, %v; want %+v, true", hat, ok, wantHat)
	}

	if it, ok := c.Item("nope"); ok {
		t.Errorf("Item(nope) = %+v, true; want not found", it)
	}
}
