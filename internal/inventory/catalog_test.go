package inventory_test

import (
	"testing"

	"github.com/Elpulgo/critterkingdom/internal/inventory"
)

func TestParseCatalog_RejectsItemDataMistakes(t *testing.T) {
	tests := []struct {
		name string
		data string
	}{
		{"duplicate id", `[
			{"id": "berry", "name": {"en": "Berry", "sv": "Bär"}, "icon": "berry"},
			{"id": "berry", "name": {"en": "Berry", "sv": "Bär"}, "icon": "berry"}]`},
		{"empty id", `[{"id": "", "name": {"en": "Berry", "sv": "Bär"}, "icon": "berry"}]`},
		{"missing en name", `[{"id": "berry", "name": {"sv": "Bär"}, "icon": "berry"}]`},
		{"missing sv name", `[{"id": "berry", "name": {"en": "Berry"}, "icon": "berry"}]`},
		{"missing icon", `[{"id": "berry", "name": {"en": "Berry", "sv": "Bär"}}]`},
		{"malformed json", `[{"id": "berry",`},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if c, err := inventory.ParseCatalog([]byte(tc.data)); err == nil {
				t.Fatalf("ParseCatalog = %+v, nil; want error", c.Items())
			}
		})
	}
}
