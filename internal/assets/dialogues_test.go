package assets_test

import (
	"errors"
	"slices"
	"testing"

	"github.com/Elpulgo/critterkingdom/internal/assets"
	"github.com/Elpulgo/critterkingdom/internal/dialogue"
	"github.com/Elpulgo/critterkingdom/internal/i18n"
)

// noConditionHolds is the quest state of a new game: no `if` condition holds
// and no action may run. It stands in for the quest system, which is not
// built yet.
type noConditionHolds struct{}

func (noConditionHolds) Holds(dialogue.Condition) (bool, error) { return false, nil }

func (noConditionHolds) Do(a dialogue.Action) error {
	return errors.New("no action may run: " + string(a))
}

func TestLoadDialogues_TheThreeVillagersCanEachStartAConversationWhenNoConditionHolds(t *testing.T) {
	dialogues, err := assets.LoadDialogues()
	if err != nil {
		t.Fatalf("LoadDialogues: %v", err)
	}

	var npcs []string
	for npc := range dialogues {
		npcs = append(npcs, npc)
	}
	slices.Sort(npcs)
	want := []string{"baker_badger", "grandma_hedgehog", "mayor_owl"}
	if !slices.Equal(npcs, want) {
		t.Fatalf("LoadDialogues() npcs = %q; want %q", npcs, want)
	}

	for _, npc := range want {
		d := dialogues[npc]
		if d.NPC != npc {
			t.Errorf("dialogues[%q].NPC = %q; want the key", npc, d.NPC)
		}
		r, err := dialogue.Start(d, noConditionHolds{}, noConditionHolds{})
		if err != nil {
			t.Errorf("%s: Start: %v", npc, err)
			continue
		}
		if r.Line().Text == "" {
			t.Errorf("%s: the conversation starts with no line", npc)
		}
	}
}

func TestLoadDialogues_EveryTextTheVillagersShowIsInBothStringTables(t *testing.T) {
	dialogues, err := assets.LoadDialogues()
	if err != nil {
		t.Fatalf("LoadDialogues: %v", err)
	}
	tables, err := assets.LoadStringTables()
	if err != nil {
		t.Fatalf("LoadStringTables: %v", err)
	}

	for npc, d := range dialogues {
		keys := []string{"npc." + npc + ".name"}
		for _, id := range d.NodeIDs() {
			n, _ := d.Node(id)
			for _, l := range n.Lines {
				keys = append(keys, l.Text)
			}
			for _, c := range n.Choices {
				keys = append(keys, c.Text)
			}
		}
		for _, key := range keys {
			for _, lang := range []i18n.Lang{i18n.Swedish, i18n.English} {
				if tables[lang][key] == "" {
					t.Errorf("%s: key %q has no %s text", npc, key, lang)
				}
			}
		}
	}
}
