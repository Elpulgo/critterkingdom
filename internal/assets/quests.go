package assets

import (
	"fmt"

	embedded "github.com/Elpulgo/critterkingdom/assets"
	"github.com/Elpulgo/critterkingdom/internal/quest"
)

// LoadQuests loads every quest definition from the embedded
// data/quests/*.json files, sorted by file name.
func LoadQuests() ([]quest.Definition, error) {
	defs, err := quest.LoadDir(embedded.FS, "data/quests")
	if err != nil {
		return nil, fmt.Errorf("load quests: %w", err)
	}
	return defs, nil
}
