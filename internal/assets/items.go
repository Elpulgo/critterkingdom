package assets

import (
	"fmt"
	"io/fs"

	embedded "github.com/Elpulgo/critterkingdom/assets"
	"github.com/Elpulgo/critterkingdom/internal/inventory"
)

// LoadItems loads the embedded item catalog at data/items.json.
func LoadItems() (*inventory.Catalog, error) {
	b, err := fs.ReadFile(embedded.FS, "data/items.json")
	if err != nil {
		return nil, fmt.Errorf("read item catalog: %w", err)
	}
	c, err := inventory.ParseCatalog(b)
	if err != nil {
		return nil, fmt.Errorf("load item catalog: %w", err)
	}
	return c, nil
}
