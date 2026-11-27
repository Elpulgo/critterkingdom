package assets

import (
	"fmt"
	"io/fs"

	embedded "github.com/Elpulgo/critterkingdom/assets"
	"github.com/Elpulgo/critterkingdom/internal/critter"
)

// LoadCritters loads every critter species from the embedded critters.json.
func LoadCritters() ([]critter.Species, error) {
	b, err := fs.ReadFile(embedded.FS, "critters.json")
	if err != nil {
		return nil, fmt.Errorf("read critters: %w", err)
	}
	species, err := critter.LoadSpecies(b)
	if err != nil {
		return nil, fmt.Errorf("load critters: %w", err)
	}
	return species, nil
}
