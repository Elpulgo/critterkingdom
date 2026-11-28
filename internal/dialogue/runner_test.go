package dialogue_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/Elpulgo/critterkingdom/internal/dialogue"
)

var errUnknownCondition = errors.New("unknown condition")

// questStates is an in-memory stand-in for the quest system, which does not
// exist yet. It decodes {"quest": ..., "is": ...} conditions itself; a quest
// it has no state for is "not_started".
type questStates map[string]string

func (q questStates) Holds(c dialogue.Condition) (bool, error) {
	var cond struct {
		Quest string `json:"quest"`
		Is    string `json:"is"`
	}
	if err := json.Unmarshal(c, &cond); err != nil {
		return false, err
	}
	if cond.Quest == "" {
		return false, fmt.Errorf("%s: %w", c, errUnknownCondition)
	}
	state, ok := q[cond.Quest]
	if !ok {
		state = "not_started"
	}
	return state == cond.Is, nil
}

func mustParse(t *testing.T, data string) *dialogue.Dialogue {
	t.Helper()
	d, err := dialogue.Parse([]byte(data))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	return d
}

func TestStart_ConditionsPickTheStartNode(t *testing.T) {
	d := mustParse(t, `{"npc": "grandma_hedgehog", "nodes": {
		"both_found": {
			"if": [{"quest": "missing_hat", "is": "done"}, {"quest": "lost_bell", "is": "done"}],
			"lines": [{"text": "grandma.both", "face": "happy"}]},
		"hat_found": {
			"if": [{"quest": "missing_hat", "is": "done"}],
			"lines": [{"text": "grandma.hat.done", "face": "happy"}]},
		"chatter": {
			"if": [],
			"lines": [{"text": "grandma.chatter", "face": "neutral"}]}}}`)

	tests := []struct {
		name   string
		quests questStates
		want   dialogue.Line
	}{
		{"first node in file order when its conditions hold",
			questStates{"missing_hat": "done", "lost_bell": "done"},
			dialogue.Line{Text: "grandma.both", Face: "happy"}},
		{"node needing two conditions is skipped when only one holds",
			questStates{"missing_hat": "done"},
			dialogue.Line{Text: "grandma.hat.done", Face: "happy"}},
		{"fallback when only the second condition of the first node holds",
			questStates{"lost_bell": "done"},
			dialogue.Line{Text: "grandma.chatter", Face: "neutral"}},
		{"fallback when nothing else holds",
			questStates{},
			dialogue.Line{Text: "grandma.chatter", Face: "neutral"}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r, err := dialogue.Start(d, tc.quests, nil)
			if err != nil {
				t.Fatalf("Start: %v", err)
			}
			if got := r.Line(); got != tc.want {
				t.Errorf("Line() = %+v, want %+v", got, tc.want)
			}
		})
	}
}

func TestStart_NodeWithoutIfKeyIsNeverTheStart(t *testing.T) {
	d := mustParse(t, `{"npc": "fox", "nodes": {
		"reply": {"lines": [{"text": "fox.reply", "face": "happy"}]},
		"hello": {"if": [], "lines": [{"text": "fox.hello", "face": "neutral"}],
			"choices": [{"text": "ui.yes", "icon": "check", "goto": "reply"}]}}}`)

	r, err := dialogue.Start(d, questStates{}, nil)
	if err != nil {
		t.Fatalf("Start: %v", err)
	}
	if got, want := r.Line(), (dialogue.Line{Text: "fox.hello", Face: "neutral"}); got != want {
		t.Errorf("Line() = %+v, want %+v", got, want)
	}
}

func TestStart_AppendixADialogueStartsOnlyBeforeTheQuest(t *testing.T) {
	data, err := os.ReadFile("testdata/grandma_hedgehog.json")
	if err != nil {
		t.Fatal(err)
	}
	d := mustParse(t, string(data))

	r, err := dialogue.Start(d, questStates{}, nil)
	if err != nil {
		t.Fatalf("Start before the quest: %v", err)
	}
	if got, want := r.Line(), (dialogue.Line{Text: "grandma.hat.1", Face: "sad"}); got != want {
		t.Errorf("Line() = %+v, want %+v", got, want)
	}

	if r, err := dialogue.Start(d, questStates{"missing_hat": "started"}, nil); !errors.Is(err, dialogue.ErrNoEntry) {
		t.Errorf("Start once the quest is started = %+v, %v; want %v", r, err, dialogue.ErrNoEntry)
	}
}

func TestStart_FailingConditionCheckIsReturned(t *testing.T) {
	d := mustParse(t, `{"npc": "fox", "nodes": {
		"rainy": {"if": [{"weather": "rain"}], "lines": [{"text": "fox.rain"}]},
		"chatter": {"if": [], "lines": [{"text": "fox.chatter"}]}}}`)

	r, err := dialogue.Start(d, questStates{}, nil)
	if !errors.Is(err, errUnknownCondition) {
		t.Errorf("Start = %+v, %v; want %v", r, err, errUnknownCondition)
	}
}
