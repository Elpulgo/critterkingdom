package assets

import (
	"fmt"
	"io/fs"

	embedded "github.com/Elpulgo/critterkingdom/assets"
	"github.com/Elpulgo/critterkingdom/internal/combat"
)

// LoadEnemies loads every enemy definition from the embedded
// data/enemies.json.
func LoadEnemies() ([]combat.EnemyDef, error) {
	b, err := fs.ReadFile(embedded.FS, "data/enemies.json")
	if err != nil {
		return nil, fmt.Errorf("read enemies: %w", err)
	}
	defs, err := combat.ParseEnemies(b)
	if err != nil {
		return nil, fmt.Errorf("load enemies: %w", err)
	}
	return defs, nil
}
