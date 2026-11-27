package assets_test

import (
	"testing"

	"github.com/Elpulgo/critterkingdom/internal/assets"
	"github.com/Elpulgo/critterkingdom/internal/i18n"
)

func TestStringTables_EveryKeyHasNonEmptyTextInBothLanguages(t *testing.T) {
	tables, err := assets.LoadStringTables()
	if err != nil {
		t.Fatalf("LoadStringTables: %v", err)
	}
	sv, en := tables[i18n.Swedish], tables[i18n.English]
	if len(sv) == 0 || len(en) == 0 {
		t.Fatalf("got %d Swedish and %d English texts; want both tables non-empty", len(sv), len(en))
	}

	pairs := []struct {
		name, otherName string
		table, other    map[string]string
	}{
		{"sv", "en", sv, en},
		{"en", "sv", en, sv},
	}
	for _, p := range pairs {
		for key, text := range p.table {
			if text == "" {
				t.Errorf("%s key %q has empty text", p.name, key)
			}
			if _, ok := p.other[key]; !ok {
				t.Errorf("%s key %q is missing from %s", p.name, key, p.otherName)
			}
		}
	}
}

func TestLoadStrings_GreetsInSwedishFirstAndInEnglishAfterSwitching(t *testing.T) {
	c, err := assets.LoadStrings()
	if err != nil {
		t.Fatalf("LoadStrings: %v", err)
	}
	if got := c.Language(); got != i18n.Swedish {
		t.Fatalf("Language() = %q; want %q", got, i18n.Swedish)
	}
	if got := c.T("ui.greeting"); got != "Hej!" {
		t.Errorf("Swedish T(ui.greeting) = %q; want %q", got, "Hej!")
	}

	if err := c.SetLanguage(i18n.English); err != nil {
		t.Fatalf("SetLanguage(English): %v", err)
	}
	if got := c.T("ui.greeting"); got != "Hi!" {
		t.Errorf("English T(ui.greeting) = %q; want %q", got, "Hi!")
	}
}
