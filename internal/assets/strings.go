package assets

import (
	"fmt"
	"io/fs"

	embedded "github.com/Elpulgo/critterkingdom/assets"
	"github.com/Elpulgo/critterkingdom/internal/i18n"
)

// stringLangs are the languages that have a string table at
// data/strings/<lang>.json.
var stringLangs = []i18n.Lang{i18n.Swedish, i18n.English}

// LoadStringTables reads and parses the embedded string table of each
// language, data/strings/<lang>.json.
func LoadStringTables() (map[i18n.Lang]map[string]string, error) {
	tables := make(map[i18n.Lang]map[string]string, len(stringLangs))
	for _, lang := range stringLangs {
		path := "data/strings/" + string(lang) + ".json"
		b, err := fs.ReadFile(embedded.FS, path)
		if err != nil {
			return nil, fmt.Errorf("read string table %s: %w", lang, err)
		}
		table, err := i18n.ParseTable(b)
		if err != nil {
			return nil, fmt.Errorf("load string table %s: %w", lang, err)
		}
		tables[lang] = table
	}
	return tables, nil
}

// LoadStrings loads the embedded Swedish and English string tables into a
// catalog that starts in Swedish.
func LoadStrings() (*i18n.Catalog, error) {
	tables, err := LoadStringTables()
	if err != nil {
		return nil, err
	}
	return i18n.NewCatalog(tables, i18n.Swedish)
}
