package dialogue_test

import (
	"errors"
	"os"
	"reflect"
	"testing"

	"github.com/Elpulgo/critterkingdom/internal/dialogue"
)

func TestParse_AppendixADialogueKeepsNodeOrderLinesAndChoices(t *testing.T) {
	data, err := os.ReadFile("testdata/grandma_hedgehog.json")
	if err != nil {
		t.Fatal(err)
	}
	d, err := dialogue.Parse(data)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if d.NPC != "grandma_hedgehog" {
		t.Errorf("NPC = %q, want grandma_hedgehog", d.NPC)
	}
	if ids := d.NodeIDs(); !reflect.DeepEqual(ids, []string{"start", "thanks"}) {
		t.Errorf("NodeIDs() = %q, want [start thanks]", ids)
	}

	start, ok := d.Node("start")
	if !ok {
		t.Fatal(`Node("start") not found`)
	}
	want := dialogue.Node{
		If:    []dialogue.Condition{dialogue.Condition(`{"quest": "missing_hat", "is": "not_started"}`)},
		HasIf: true,
		Lines: []dialogue.Line{
			{Text: "grandma.hat.1", Face: "sad"},
			{Text: "grandma.hat.2", Face: "neutral"},
		},
		Choices: []dialogue.Choice{
			{Text: "ui.yes", Icon: "check", Do: []dialogue.Action{dialogue.Action(`{"startQuest": "missing_hat"}`)}, Goto: "thanks"},
			{Text: "ui.later", Icon: "cross", Goto: "end"},
		},
	}
	if !reflect.DeepEqual(start, want) {
		t.Errorf("start node = %+v\nwant %+v", start, want)
	}

	thanks, ok := d.Node("thanks")
	if !ok {
		t.Fatal(`Node("thanks") not found`)
	}
	if thanks.HasIf {
		t.Error("thanks node has HasIf = true, want false: it has no if key")
	}
}

func TestParse_ChoiceGotoMustNameANodeInTheFileOrEnd(t *testing.T) {
	tests := []struct {
		name    string
		choice  string
		wantErr error
	}{
		{"missing node", `{"text": "ui.yes", "icon": "check", "goto": "nowhere"}`, dialogue.ErrUnknownNode},
		{"end", `{"text": "ui.yes", "icon": "check", "goto": "end"}`, nil},
		{"omitted goto", `{"text": "ui.yes", "icon": "check"}`, nil},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			data := `{"npc": "fox", "nodes": {"start": {"if": [],
				"lines": [{"text": "fox.1"}], "choices": [` + tc.choice + `]}}}`
			_, err := dialogue.Parse([]byte(data))
			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("Parse error = %v, want %v", err, tc.wantErr)
			}
		})
	}
}

func TestParse_NodeWithMoreThanTwoChoicesIsRejected(t *testing.T) {
	data := `{"npc": "fox", "nodes": {"start": {"if": [], "lines": [{"text": "fox.1"}], "choices": [
		{"text": "ui.yes", "icon": "check"},
		{"text": "ui.later", "icon": "cross"},
		{"text": "ui.maybe", "icon": "star"}]}}}`
	_, err := dialogue.Parse([]byte(data))
	if !errors.Is(err, dialogue.ErrTooManyChoices) {
		t.Fatalf("Parse error = %v, want %v", err, dialogue.ErrTooManyChoices)
	}
}

func TestParse_RejectsStructurallyInvalidFiles(t *testing.T) {
	tests := []struct {
		name string
		data string
	}{
		{"empty npc", `{"npc": "", "nodes": {"start": {"if": [], "lines": [{"text": "fox.1"}]}}}`},
		{"no nodes", `{"npc": "fox", "nodes": {}}`},
		{"no entry node", `{"npc": "fox", "nodes": {"start": {"lines": [{"text": "fox.1"}]}}}`},
		{"node named end", `{"npc": "fox", "nodes": {
			"start": {"if": [], "lines": [{"text": "fox.1"}]},
			"end": {"lines": [{"text": "fox.2"}]}}}`},
		{"duplicate node id", `{"npc": "fox", "nodes": {
			"start": {"if": [], "lines": [{"text": "fox.1"}]},
			"start": {"if": [], "lines": [{"text": "fox.2"}]}}}`},
		{"node without lines", `{"npc": "fox", "nodes": {"start": {"if": [], "lines": []}}}`},
		{"empty line text", `{"npc": "fox", "nodes": {"start": {"if": [], "lines": [{"text": "", "face": "happy"}]}}}`},
		{"empty choice text", `{"npc": "fox", "nodes": {"start": {"if": [], "lines": [{"text": "fox.1"}],
			"choices": [{"text": "", "icon": "check"}]}}}`},
		{"malformed json", `{"npc": "fox", "nodes": {"start": {`},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if d, err := dialogue.Parse([]byte(tc.data)); err == nil {
				t.Fatalf("Parse = %+v, nil; want error", d)
			}
		})
	}
}
