// Package i18n looks up the game's texts by key in Swedish and English string
// tables, with a fallback for missing keys and switching language at run time.
//
// A Catalog holds the tables and the current language. Game code calls the
// package-level T, SetLanguage and Language, which work on the catalog
// installed with Use; tests build their own Catalog instead.
package i18n

import (
	"encoding/json"
	"fmt"
)

// Lang names a language. Its value matches the string table's file name,
// strings/<lang>.json.
type Lang string

const (
	// Swedish is the game's first language.
	Swedish Lang = "sv"
	// English is the game's second language.
	English Lang = "en"
)

// fallbackOrder is the fixed order in which other languages are tried when a
// key is missing or empty in the current language.
var fallbackOrder = []Lang{Swedish, English}

// ParseTable parses a string table: a flat JSON object of key to text.
// Anything else, such as nested objects, numbers, null or a top-level array,
// is an error.
func ParseTable(data []byte) (map[string]string, error) {
	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("parse string table: %w", err)
	}
	if raw == nil {
		return nil, fmt.Errorf("parse string table: want a JSON object of key to text")
	}
	table := make(map[string]string, len(raw))
	for key, value := range raw {
		text, ok := value.(string)
		if !ok {
			return nil, fmt.Errorf("parse string table: key %q has a %T value, want a string", key, value)
		}
		table[key] = text
	}
	return table, nil
}

// Catalog holds the string table of each language and the current language.
// It is not safe for concurrent use.
type Catalog struct {
	tables  map[Lang]map[string]string
	current Lang
}

// NewCatalog returns a catalog with the given tables, starting in current.
// It is an error if current has no table.
func NewCatalog(tables map[Lang]map[string]string, current Lang) (*Catalog, error) {
	if _, ok := tables[current]; !ok {
		return nil, fmt.Errorf("no string table for language %q", current)
	}
	copied := make(map[Lang]map[string]string, len(tables))
	for lang, table := range tables {
		copied[lang] = table
	}
	return &Catalog{tables: copied, current: current}, nil
}

// T returns the text for key in the current language. When the key is missing
// or empty there, it returns the first non-empty text from the other languages
// in the order Swedish, English, and when no language has it, the key itself.
func (c *Catalog) T(key string) string {
	if text := c.tables[c.current][key]; text != "" {
		return text
	}
	for _, lang := range fallbackOrder {
		if lang == c.current {
			continue
		}
		if text := c.tables[lang][key]; text != "" {
			return text
		}
	}
	return key
}

// SetLanguage makes lang the current language. It returns an error and keeps
// the current language when lang has no table.
func (c *Catalog) SetLanguage(lang Lang) error {
	if _, ok := c.tables[lang]; !ok {
		return fmt.Errorf("no string table for language %q", lang)
	}
	c.current = lang
	return nil
}

// Language returns the current language.
func (c *Catalog) Language() Lang {
	return c.current
}

// active is the catalog the package-level functions use; nil until Use.
var active *Catalog

// Use installs c as the catalog behind the package-level T, SetLanguage and
// Language. Use(nil) removes it again.
func Use(c *Catalog) {
	active = c
}

// T returns the text for key from the catalog installed with Use, following
// Catalog.T's fallback. Before Use it returns the key.
func T(key string) string {
	if active == nil {
		return key
	}
	return active.T(key)
}

// SetLanguage switches the installed catalog to lang. It returns an error when
// no catalog is installed or lang has no table.
func SetLanguage(lang Lang) error {
	if active == nil {
		return fmt.Errorf("no string catalog installed")
	}
	return active.SetLanguage(lang)
}

// Language returns the installed catalog's current language, or "" before Use.
func Language() Lang {
	if active == nil {
		return ""
	}
	return active.Language()
}
