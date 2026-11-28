package assets_test

import (
	"reflect"
	"testing"

	"github.com/Elpulgo/critterkingdom/internal/assets"
	"github.com/Elpulgo/critterkingdom/internal/quest"
)

func TestLoadQuests_LoadsTheMissingHatQuestFromTheEmbeddedFiles(t *testing.T) {
	defs, err := assets.LoadQuests()
	if err != nil {
		t.Fatalf("LoadQuests: %v", err)
	}

	want := []quest.Definition{{
		ID:    "missing_hat",
		Giver: "grandma_hedgehog",
		Title: quest.Text{"en": "The Missing Hat", "sv": "Den försvunna hatten"},
		Steps: []quest.Step{
			{Type: quest.GoTo, Target: "meadow", Icon: "footsteps"},
			{Type: quest.Collect, Item: "grandmas_hat", Count: 1, Icon: "hat"},
			{Type: quest.Talk, NPC: "grandma_hedgehog", Icon: "speech"},
		},
		Reward: quest.Reward{Sticker: "hat_hero", Item: "cookie"},
	}}
	if !reflect.DeepEqual(defs, want) {
		t.Errorf("LoadQuests() =\n%+v\nwant\n%+v", defs, want)
	}
}
