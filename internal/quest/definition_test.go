package quest_test

import (
	"reflect"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/Elpulgo/critterkingdom/internal/quest"
)

// missingHatJSON is the quest sketch from Appendix A of docs/PLAN.md.
const missingHatJSON = `{
  "id": "missing_hat",
  "giver": "grandma_hedgehog",
  "title": {"en": "The Missing Hat", "sv": "Den försvunna hatten"},
  "steps": [
    {"type": "goto",    "target": "meadow",                   "icon": "footsteps"},
    {"type": "collect", "item": "grandmas_hat", "count": 1,   "icon": "hat"},
    {"type": "talk",    "npc": "grandma_hedgehog",            "icon": "speech"}
  ],
  "reward": {"sticker": "hat_hero", "item": "cookie"}
}`

func TestParse_AppendixAMissingHatQuest(t *testing.T) {
	got, err := quest.Parse([]byte(missingHatJSON))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	want := quest.Definition{
		ID:    "missing_hat",
		Giver: "grandma_hedgehog",
		Title: quest.Text{"en": "The Missing Hat", "sv": "Den försvunna hatten"},
		Steps: []quest.Step{
			{Type: quest.GoTo, Target: "meadow", Icon: "footsteps"},
			{Type: quest.Collect, Item: "grandmas_hat", Count: 1, Icon: "hat"},
			{Type: quest.Talk, NPC: "grandma_hedgehog", Icon: "speech"},
		},
		Reward: quest.Reward{Sticker: "hat_hero", Item: "cookie"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Parse =\n%+v\nwant\n%+v", got, want)
	}
}

func TestParse_AllNineStepTypes_OmittedCountMeansOne(t *testing.T) {
	data := `{
	  "id": "grand_tour",
	  "steps": [
	    {"type": "talk",     "npc": "mole"},
	    {"type": "goto",     "target": "pond"},
	    {"type": "collect",  "item": "berry", "count": 5},
	    {"type": "collect",  "item": "acorn"},
	    {"type": "deliver",  "item": "plank", "npc": "beaver", "count": 3},
	    {"type": "deliver",  "item": "letter", "npc": "owl"},
	    {"type": "befriend", "critter": "firefly"},
	    {"type": "find",     "target": "lost_key"},
	    {"type": "puzzle",   "room": "mirror_room"},
	    {"type": "defeat",   "enemy": "sulky_slime", "count": 3, "icon": "stick"},
	    {"type": "defeat",   "enemy": "grumpy_bee"},
	    {"type": "clear",    "room": "cellar"}
	  ]
	}`
	got, err := quest.Parse([]byte(data))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	want := []quest.Step{
		{Type: quest.Talk, NPC: "mole"},
		{Type: quest.GoTo, Target: "pond"},
		{Type: quest.Collect, Item: "berry", Count: 5},
		{Type: quest.Collect, Item: "acorn", Count: 1},
		{Type: quest.Deliver, Item: "plank", NPC: "beaver", Count: 3},
		{Type: quest.Deliver, Item: "letter", NPC: "owl", Count: 1},
		{Type: quest.Befriend, Critter: "firefly"},
		{Type: quest.Find, Target: "lost_key"},
		{Type: quest.Puzzle, Room: "mirror_room"},
		{Type: quest.Defeat, Enemy: "sulky_slime", Count: 3, Icon: "stick"},
		{Type: quest.Defeat, Enemy: "grumpy_bee", Count: 1},
		{Type: quest.Clear, Room: "cellar"},
	}
	if !reflect.DeepEqual(got.Steps, want) {
		t.Errorf("Parse steps =\n%+v\nwant\n%+v", got.Steps, want)
	}
}

func TestParse_RejectsInvalidQuests(t *testing.T) {
	// quest1 puts the step after a valid first step, so step errors name index 1.
	quest1 := func(step string) string {
		return `{"id": "q", "steps": [{"type": "talk", "npc": "mole"}, ` + step + `]}`
	}
	tests := []struct {
		name string
		data string
		// wantErr lists substrings the error must contain.
		wantErr []string
	}{
		{"missing id", `{"steps": [{"type": "talk", "npc": "mole"}]}`, []string{"missing id"}},
		{"no steps", `{"id": "q", "steps": []}`, []string{`"q"`, "no steps"}},
		{"steps omitted", `{"id": "q"}`, []string{`"q"`, "no steps"}},
		{"unknown type", quest1(`{"type": "dance", "npc": "mole"}`), []string{`"q"`, "step 1", `"dance"`}},
		{"talk without npc", quest1(`{"type": "talk"}`), []string{"step 1", "npc"}},
		{"goto without target", quest1(`{"type": "goto"}`), []string{"step 1", "target"}},
		{"collect without item", quest1(`{"type": "collect", "count": 2}`), []string{"step 1", "item"}},
		{"deliver without item", quest1(`{"type": "deliver", "npc": "owl"}`), []string{"step 1", "item"}},
		{"deliver without npc", quest1(`{"type": "deliver", "item": "letter"}`), []string{"step 1", "npc"}},
		{"befriend without critter", quest1(`{"type": "befriend"}`), []string{"step 1", "critter"}},
		{"find without target", quest1(`{"type": "find"}`), []string{"step 1", "target"}},
		{"puzzle without room", quest1(`{"type": "puzzle"}`), []string{"step 1", "room"}},
		{"defeat without enemy", quest1(`{"type": "defeat", "count": 3}`), []string{"step 1", "enemy"}},
		{"clear without room", quest1(`{"type": "clear"}`), []string{"step 1", "room"}},
		{"negative count", quest1(`{"type": "collect", "item": "berry", "count": -2}`), []string{"step 1", "negative count"}},
		{"count on uncounted step", quest1(`{"type": "goto", "target": "pond", "count": 2}`), []string{"step 1", "count"}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			def, err := quest.Parse([]byte(tc.data))
			if err == nil {
				t.Fatalf("Parse succeeded with %+v; want error", def)
			}
			for _, want := range tc.wantErr {
				if !strings.Contains(err.Error(), want) {
					t.Errorf("Parse error = %q; want it to contain %q", err, want)
				}
			}
		})
	}
}

func questFile(id string) *fstest.MapFile {
	return &fstest.MapFile{Data: []byte(`{"id": "` + id + `", "steps": [{"type": "talk", "npc": "mole"}]}`)}
}

func TestLoadDir_ReturnsJSONQuestsInFileNameOrder(t *testing.T) {
	fsys := fstest.MapFS{
		"data/quests/c_pie.json":     questFile("berry_pie"),
		"data/quests/a_hat.json":     questFile("missing_hat"),
		"data/quests/b_kite.json":    questFile("lost_kite"),
		"data/quests/README.md":      {Data: []byte("not a quest")},
		"data/quests/draft.json.bak": {Data: []byte("{")},
		"data/quests/old/x.json":     questFile("old_quest"),
		"data/critters.json":         {Data: []byte("[]")},
	}
	defs, err := quest.LoadDir(fsys, "data/quests")
	if err != nil {
		t.Fatalf("LoadDir: %v", err)
	}
	var ids []string
	for _, d := range defs {
		ids = append(ids, d.ID)
	}
	want := []string{"missing_hat", "lost_kite", "berry_pie"}
	if !reflect.DeepEqual(ids, want) {
		t.Errorf("LoadDir ids = %v; want %v", ids, want)
	}
}

func TestLoadDir_RejectsTwoFilesWithTheSameQuestID(t *testing.T) {
	fsys := fstest.MapFS{
		"quests/hat.json":      questFile("missing_hat"),
		"quests/hat_copy.json": questFile("missing_hat"),
	}
	defs, err := quest.LoadDir(fsys, "quests")
	if err == nil {
		t.Fatalf("LoadDir succeeded with %d quests; want error", len(defs))
	}
	for _, want := range []string{`"missing_hat"`, "duplicate id"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("LoadDir error = %q; want it to contain %q", err, want)
		}
	}
}
