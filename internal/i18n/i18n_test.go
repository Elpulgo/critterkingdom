package i18n

import "testing"

// testTables returns fresh tables for one test, so no test shares state.
func testTables() map[Lang]map[string]string {
	return map[Lang]map[string]string{
		Swedish: {
			"ui.greeting": "Hej!",
			"ui.only_sv":  "Bara svenska",
			"ui.empty_sv": "",
			"ui.empty":    "",
		},
		English: {
			"ui.greeting": "Hi!",
			"ui.only_en":  "Only English",
			"ui.empty_sv": "Empty in Swedish",
			"ui.empty":    "",
		},
	}
}

func newTestCatalog(t *testing.T, current Lang) *Catalog {
	t.Helper()
	c, err := NewCatalog(testTables(), current)
	if err != nil {
		t.Fatalf("NewCatalog: %v", err)
	}
	return c
}

func TestCatalogT_MissingOrEmptyTextFallsBackToTheOtherLanguageAndThenTheKey(t *testing.T) {
	tests := []struct {
		name    string
		current Lang
		key     string
		want    string
	}{
		{"text in the current language is shown", Swedish, "ui.greeting", "Hej!"},
		{"text in the current language is shown in English too", English, "ui.greeting", "Hi!"},
		{"a key missing in Swedish shows the English text", Swedish, "ui.only_en", "Only English"},
		{"a key missing in English shows the Swedish text", English, "ui.only_sv", "Bara svenska"},
		{"an empty Swedish text shows the English text", Swedish, "ui.empty_sv", "Empty in Swedish"},
		{"a key in no language shows the key", Swedish, "ui.typo", "ui.typo"},
		{"a key empty in every language shows the key", English, "ui.empty", "ui.empty"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			c := newTestCatalog(t, tc.current)
			if got := c.T(tc.key); got != tc.want {
				t.Errorf("T(%q) in %q = %q, want %q", tc.key, tc.current, got, tc.want)
			}
		})
	}
}

func TestCatalogSetLanguage_SwitchingLanguageChangesTheText(t *testing.T) {
	c := newTestCatalog(t, Swedish)
	if got := c.T("ui.greeting"); got != "Hej!" {
		t.Fatalf("T before switching = %q, want %q", got, "Hej!")
	}

	if err := c.SetLanguage(English); err != nil {
		t.Fatalf("SetLanguage(English): %v", err)
	}
	if got := c.Language(); got != English {
		t.Errorf("Language() = %q, want %q", got, English)
	}
	if got := c.T("ui.greeting"); got != "Hi!" {
		t.Errorf("T after switching = %q, want %q", got, "Hi!")
	}
}

func TestCatalogSetLanguage_ALanguageWithoutATableIsRefusedAndTheCurrentOneStays(t *testing.T) {
	c := newTestCatalog(t, English)

	if err := c.SetLanguage("de"); err == nil {
		t.Fatal(`SetLanguage("de") returned no error`)
	}
	if got := c.Language(); got != English {
		t.Errorf("Language() after a refused switch = %q, want %q", got, English)
	}
	if got := c.T("ui.greeting"); got != "Hi!" {
		t.Errorf("T after a refused switch = %q, want %q", got, "Hi!")
	}
}

func TestNewCatalog_StartingInALanguageWithoutATableIsAnError(t *testing.T) {
	tables := map[Lang]map[string]string{Swedish: {"ui.greeting": "Hej!"}}
	if _, err := NewCatalog(tables, English); err == nil {
		t.Fatal("NewCatalog starting in English without an English table returned no error")
	}
}

func TestParseTable_ATableThatIsNotAFlatObjectOfStringsIsRejected(t *testing.T) {
	tests := []struct {
		name string
		data string
	}{
		{"nested object", `{"ui": {"greeting": "Hej!"}}`},
		{"number value", `{"ui.count": 3}`},
		{"null value", `{"ui.greeting": null}`},
		{"list value", `{"ui.greeting": ["Hej!"]}`},
		{"top-level array", `["Hej!"]`},
		{"top-level string", `"Hej!"`},
		{"top-level null", `null`},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if table, err := ParseTable([]byte(tc.data)); err == nil {
				t.Errorf("ParseTable(%s) = %v, want an error", tc.data, table)
			}
		})
	}
}

func TestPackageT_ShowsTheKeyBeforeUseAndTheInstalledCatalogsTextAfter(t *testing.T) {
	Use(nil)
	t.Cleanup(func() { Use(nil) })

	if got := T("ui.greeting"); got != "ui.greeting" {
		t.Errorf("T before Use = %q, want the key %q", got, "ui.greeting")
	}

	Use(newTestCatalog(t, Swedish))
	if got := T("ui.greeting"); got != "Hej!" {
		t.Errorf("T after Use = %q, want %q", got, "Hej!")
	}
	if err := SetLanguage(English); err != nil {
		t.Fatalf("SetLanguage(English): %v", err)
	}
	if got := Language(); got != English {
		t.Errorf("Language() = %q, want %q", got, English)
	}
	if got := T("ui.greeting"); got != "Hi!" {
		t.Errorf("T after SetLanguage(English) = %q, want %q", got, "Hi!")
	}
}
