package assets

import (
	"fmt"
	"io/fs"
	"path"
	"sort"
	"strings"

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

// MapNames lists the names of every embedded map, the files maps/*.tmj
// without their extension, sorted. Each name loads with LoadMap.
func MapNames() ([]string, error) {
	files, err := fs.Glob(embedded.FS, "maps/*.tmj")
	if err != nil {
		return nil, fmt.Errorf("list maps: %w", err)
	}
	names := make([]string, len(files))
	for i, f := range files {
		names[i] = strings.TrimSuffix(path.Base(f), ".tmj")
	}
	sort.Strings(names)
	return names, nil
}
