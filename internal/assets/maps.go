package assets

import (
	"fmt"
	"io/fs"

	embedded "github.com/Elpulgo/critterkingdom/assets"
	"github.com/Elpulgo/critterkingdom/internal/world"
)

// LoadMap loads the embedded Tiled JSON map at maps/<name>.tmj.
func LoadMap(name string) (*world.Map, error) {
	b, err := fs.ReadFile(embedded.FS, "maps/"+name+".tmj")
	if err != nil {
		return nil, fmt.Errorf("read map %s: %w", name, err)
	}
	m, err := world.LoadMap(b)
	if err != nil {
		return nil, fmt.Errorf("load map %s: %w", name, err)
	}
	return m, nil
}
