package assets

import (
	"fmt"
	"io/fs"
	"path"

	embedded "github.com/Elpulgo/critterkingdom/assets"
	"github.com/Elpulgo/critterkingdom/internal/dialogue"
)

// dialogueDir is the embedded directory holding one dialogue file per
// villager.
const dialogueDir = "data/dialogue"

// LoadDialogues loads every villager's dialogue from the embedded
// data/dialogue/*.json files, keyed by the npc each file names. It fails,
// naming the file, when a file does not parse or names the same npc as
// another file.
func LoadDialogues() (map[string]*dialogue.Dialogue, error) {
	entries, err := fs.ReadDir(embedded.FS, dialogueDir)
	if err != nil {
		return nil, fmt.Errorf("load dialogues: %w", err)
	}
	dialogues := make(map[string]*dialogue.Dialogue)
	fileOf := make(map[string]string)
	for _, e := range entries {
		if e.IsDir() || path.Ext(e.Name()) != ".json" {
			continue
		}
		b, err := fs.ReadFile(embedded.FS, path.Join(dialogueDir, e.Name()))
		if err != nil {
			return nil, fmt.Errorf("load dialogues: %w", err)
		}
		d, err := dialogue.Parse(b)
		if err != nil {
			return nil, fmt.Errorf("load dialogues: %s: %w", e.Name(), err)
		}
		if first, ok := fileOf[d.NPC]; ok {
			return nil, fmt.Errorf("load dialogues: %s: npc %q: duplicate, also in %s", e.Name(), d.NPC, first)
		}
		fileOf[d.NPC] = e.Name()
		dialogues[d.NPC] = d
	}
	return dialogues, nil
}
