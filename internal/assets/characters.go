package assets

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"path"

	embedded "github.com/Elpulgo/critterkingdom/assets"
	"github.com/Elpulgo/critterkingdom/internal/anim"
)

// LoadCharacter loads the character with the given id, such as "fox": its
// motion data from the embedded data/rigs/<id>.json, joined with the part
// sheet that data's sprite names, sprites/<sprite>/<name>.json (for
// "critters/fox" that is sprites/critters/fox/fox.json). The error names the
// file that is missing or invalid.
func LoadCharacter(id string) (anim.Character, error) {
	rigFile := "data/rigs/" + id + ".json"
	rig, err := fs.ReadFile(embedded.FS, rigFile)
	if err != nil {
		return anim.Character{}, fmt.Errorf("read %s: %w", rigFile, err)
	}
	var head struct {
		Sprite string `json:"sprite"`
	}
	if err := json.Unmarshal(rig, &head); err != nil {
		return anim.Character{}, fmt.Errorf("load %s: %w", rigFile, err)
	}
	if head.Sprite == "" {
		return anim.Character{}, fmt.Errorf("load %s: no sprite named", rigFile)
	}

	sheetFile := "sprites/" + head.Sprite + "/" + path.Base(head.Sprite) + ".json"
	sheet, err := fs.ReadFile(embedded.FS, sheetFile)
	if err != nil {
		return anim.Character{}, fmt.Errorf("read %s: %w", sheetFile, err)
	}
	layout, err := anim.ParseLayout(sheet)
	if err != nil {
		return anim.Character{}, fmt.Errorf("load %s: %w", sheetFile, err)
	}

	c, err := anim.ParseCharacter(layout, rig)
	if err != nil {
		return anim.Character{}, fmt.Errorf("load %s: %w", rigFile, err)
	}
	return c, nil
}
