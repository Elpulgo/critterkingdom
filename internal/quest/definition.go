// Package quest holds the quest rules: quest definitions loaded from
// quests/*.json, and the event-driven state that moves a quest from step to
// step until it is done and its reward is handed out. It is pure Go so every
// rule can be tested without a window.
package quest

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"path"
)

// Text is a string in each language, keyed by language code such as "en".
type Text map[string]string

// StepType is the kind of thing a quest step asks the player to do.
type StepType string

// The step types a quest file can use in a step's "type".
const (
	// Talk asks the player to talk to an NPC (field npc).
	Talk StepType = "talk"
	// GoTo asks the player to go to an area (field target).
	GoTo StepType = "goto"
	// Collect asks the player to pick up count of an item (fields item,
	// count).
	Collect StepType = "collect"
	// Deliver asks the player to hand count of an item to an NPC (fields
	// item, npc, count).
	Deliver StepType = "deliver"
	// Befriend asks the player to befriend a critter (field critter).
	Befriend StepType = "befriend"
	// Find asks the player to find a hidden thing (field target).
	Find StepType = "find"
	// Puzzle asks the player to solve a puzzle room (field room).
	Puzzle StepType = "puzzle"
	// Defeat asks the player to calm count grumpies of a kind (fields enemy,
	// count).
	Defeat StepType = "defeat"
	// Clear asks the player to clear a room of grumpies (field room).
	Clear StepType = "clear"
)

// Step is one thing a quest asks the player to do. Which id fields are used
// depends on Type; the others are empty.
type Step struct {
	Type StepType `json:"type"`
	// Target is the area of a goto step or the hidden thing of a find step.
	Target string `json:"target,omitempty"`
	// Item is the item of a collect or deliver step.
	Item string `json:"item,omitempty"`
	// Count is how many a collect, deliver or defeat step needs. After
	// loading it is at least 1 on those steps and 0 on every other step.
	Count int `json:"count,omitempty"`
	// NPC is the NPC of a talk step, or who receives a deliver step's item.
	NPC string `json:"npc,omitempty"`
	// Enemy is the grumpy kind of a defeat step.
	Enemy string `json:"enemy,omitempty"`
	// Critter is the species of a befriend step.
	Critter string `json:"critter,omitempty"`
	// Room is the room of a puzzle or clear step.
	Room string `json:"room,omitempty"`
	// Icon is the HUD icon shown for the step.
	Icon string `json:"icon,omitempty"`
}

// Reward is what finishing a quest gives; both parts are optional.
type Reward struct {
	Sticker string `json:"sticker,omitempty"`
	Item    string `json:"item,omitempty"`
}

// Definition is one quest as described in a quests/<id>.json file.
type Definition struct {
	ID string `json:"id"`
	// Giver is the NPC id who hands out the quest.
	Giver  string `json:"giver"`
	Title  Text   `json:"title"`
	Steps  []Step `json:"steps"`
	Reward Reward `json:"reward"`
}

// Parse reads one quest file and checks it for the mistakes it can show on
// its own: a missing id, no steps, an unknown step type, a step missing an id
// field its type needs, a negative count, and a count on a step type that is
// not counted. An omitted count on a collect, deliver or defeat step becomes
// 1.
func Parse(data []byte) (Definition, error) {
	var def Definition
	if err := json.Unmarshal(data, &def); err != nil {
		return Definition{}, fmt.Errorf("parse quest: %w", err)
	}
	if def.ID == "" {
		return Definition{}, errors.New("quest: missing id")
	}
	if len(def.Steps) == 0 {
		return Definition{}, fmt.Errorf("quest %q: no steps", def.ID)
	}
	for i := range def.Steps {
		if err := def.Steps[i].normalize(); err != nil {
			return Definition{}, fmt.Errorf("quest %q step %d: %w", def.ID, i, err)
		}
	}
	return def, nil
}

// normalize checks the step against its type and fills in the default count.
func (s *Step) normalize() error {
	type field struct {
		name, value string
	}
	var required []field
	counted := false
	switch s.Type {
	case Talk:
		required = []field{{"npc", s.NPC}}
	case GoTo:
		required = []field{{"target", s.Target}}
	case Collect:
		required = []field{{"item", s.Item}}
		counted = true
	case Deliver:
		required = []field{{"item", s.Item}, {"npc", s.NPC}}
		counted = true
	case Befriend:
		required = []field{{"critter", s.Critter}}
	case Find:
		required = []field{{"target", s.Target}}
	case Puzzle:
		required = []field{{"room", s.Room}}
	case Defeat:
		required = []field{{"enemy", s.Enemy}}
		counted = true
	case Clear:
		required = []field{{"room", s.Room}}
	default:
		return fmt.Errorf("unknown type %q", s.Type)
	}
	for _, f := range required {
		if f.value == "" {
			return fmt.Errorf("%s step missing %s", s.Type, f.name)
		}
	}
	switch {
	case s.Count < 0:
		return fmt.Errorf("negative count %d", s.Count)
	case !counted && s.Count != 0:
		return fmt.Errorf("%s step does not take a count", s.Type)
	case counted && s.Count == 0:
		s.Count = 1
	}
	return nil
}

// LoadDir parses every *.json file in dir of fsys as one quest each and
// returns them sorted by file name. Other files and subdirectories are
// ignored. Two files with the same quest id are an error.
func LoadDir(fsys fs.FS, dir string) ([]Definition, error) {
	entries, err := fs.ReadDir(fsys, dir)
	if err != nil {
		return nil, fmt.Errorf("load quests: %w", err)
	}
	var defs []Definition
	fileOf := make(map[string]string)
	for _, e := range entries {
		if e.IsDir() || path.Ext(e.Name()) != ".json" {
			continue
		}
		data, err := fs.ReadFile(fsys, path.Join(dir, e.Name()))
		if err != nil {
			return nil, fmt.Errorf("load quests: %w", err)
		}
		def, err := Parse(data)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", e.Name(), err)
		}
		if first, ok := fileOf[def.ID]; ok {
			return nil, fmt.Errorf("%s: quest %q: duplicate id, also in %s", e.Name(), def.ID, first)
		}
		fileOf[def.ID] = e.Name()
		defs = append(defs, def)
	}
	return defs, nil
}
