package dialogue_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"os"
	"testing"

	"github.com/Elpulgo/critterkingdom/internal/dialogue"
)

var (
	errUnknownCondition = errors.New("unknown condition")
	errUnknownAction    = errors.New("unknown action")
)

// questStates is an in-memory stand-in for the quest system, which does not
// exist yet. It decodes {"quest": ..., "is": ...} conditions and
// {"startQuest": ...} / {"finishQuest": ...} actions itself; a quest it has
// no state for is "not_started".
type questStates map[string]string

func (q questStates) Do(a dialogue.Action) error {
	var act struct {
		StartQuest  string `json:"startQuest"`
		FinishQuest string `json:"finishQuest"`
	}
	if err := json.Unmarshal(a, &act); err != nil {
		return err
	}
	switch {
	case act.StartQuest != "":
		q[act.StartQuest] = "started"
	case act.FinishQuest != "":
		q[act.FinishQuest] = "done"
	default:
		return fmt.Errorf("%s: %w", a, errUnknownAction)
	}
	return nil
}

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

func parseFixture(t *testing.T) *dialogue.Dialogue {
	t.Helper()
	data, err := os.ReadFile("testdata/grandma_hedgehog.json")
	if err != nil {
		t.Fatal(err)
	}
	return mustParse(t, string(data))
}

func TestStart_AppendixADialogueHasNothingToSayOnceTheQuestIsStarted(t *testing.T) {
	if r, err := dialogue.Start(parseFixture(t), questStates{"missing_hat": "started"}, nil); !errors.Is(err, dialogue.ErrNoEntry) {
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

func TestRunner_AppendixAConversationPlaysToTheEndWhenThePlayerSaysYes(t *testing.T) {
	quests := questStates{}
	r, err := dialogue.Start(parseFixture(t), quests, quests)
	if err != nil {
		t.Fatalf("Start: %v", err)
	}
	if got, want := r.Line(), (dialogue.Line{Text: "grandma.hat.1", Face: "sad"}); got != want {
		t.Errorf("first Line() = %+v, want %+v", got, want)
	}
	if got := r.Choices(); len(got) != 0 {
		t.Errorf("Choices() on the first line = %+v, want none", got)
	}

	if err := r.Advance(); err != nil {
		t.Fatalf("Advance to the question: %v", err)
	}
	if got, want := r.Line(), (dialogue.Line{Text: "grandma.hat.2", Face: "neutral"}); got != want {
		t.Errorf("question Line() = %+v, want %+v", got, want)
	}
	choices := r.Choices()
	if len(choices) != 2 ||
		choices[0].Text != "ui.yes" || choices[0].Icon != "check" ||
		choices[1].Text != "ui.later" || choices[1].Icon != "cross" {
		t.Fatalf("Choices() on the question = %+v, want ui.yes/check and ui.later/cross", choices)
	}

	if err := r.Choose(0); err != nil {
		t.Fatalf("Choose yes: %v", err)
	}
	if got := quests["missing_hat"]; got != "started" {
		t.Errorf("quest missing_hat = %q after yes, want started", got)
	}
	if got, want := r.Line(), (dialogue.Line{Text: "grandma.hat.thanks", Face: "happy"}); got != want {
		t.Errorf("Line() after yes = %+v, want %+v", got, want)
	}
	if r.Done() {
		t.Fatal("Done() on the thanks line, want false")
	}

	if err := r.Advance(); err != nil {
		t.Fatalf("Advance past thanks: %v", err)
	}
	if !r.Done() {
		t.Error("Done() after the thanks line = false, want true")
	}
}

func TestChoose_RunsTheChoicesActionsAndGoesToItsNode(t *testing.T) {
	// The "after" node's `if` never holds: a goto lands on it anyway.
	const file = `{"npc": "fox", "nodes": {
		"ask": {"if": [], "lines": [{"text": "fox.ask", "face": "neutral"}],
			"choices": [
				{"text": "ui.yes", "icon": "check",
					"do": [{"startQuest": "berries"}, {"finishQuest": "berries"}], "goto": "after"},
				{"text": "ui.later", "icon": "cross", "goto": "end"}]},
		"after": {"if": [{"quest": "berries", "is": "never"}],
			"lines": [{"text": "fox.thanks", "face": "happy"}]}}}`

	tests := []struct {
		name       string
		choice     int
		wantQuests questStates
		wantDone   bool
		wantLine   dialogue.Line
	}{
		{"actions run in order and the goto node is next", 0,
			questStates{"berries": "done"}, false,
			dialogue.Line{Text: "fox.thanks", Face: "happy"}},
		{"goto end runs no action and ends the conversation", 1,
			questStates{}, true, dialogue.Line{}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			quests := questStates{}
			r, err := dialogue.Start(mustParse(t, file), quests, quests)
			if err != nil {
				t.Fatalf("Start: %v", err)
			}
			if err := r.Choose(tc.choice); err != nil {
				t.Fatalf("Choose(%d): %v", tc.choice, err)
			}
			if !maps.Equal(quests, tc.wantQuests) {
				t.Errorf("quests = %v, want %v", quests, tc.wantQuests)
			}
			if r.Done() != tc.wantDone {
				t.Errorf("Done() = %v, want %v", r.Done(), tc.wantDone)
			}
			if got := r.Line(); got != tc.wantLine {
				t.Errorf("Line() = %+v, want %+v", got, tc.wantLine)
			}
		})
	}
}

func TestChoose_WithoutGotoEndsTheConversation(t *testing.T) {
	r, err := dialogue.Start(mustParse(t, `{"npc": "fox", "nodes": {
		"ask": {"if": [], "lines": [{"text": "fox.ask"}],
			"choices": [{"text": "ui.ok", "icon": "check"}]}}}`), questStates{}, questStates{})
	if err != nil {
		t.Fatalf("Start: %v", err)
	}
	if err := r.Choose(0); err != nil {
		t.Fatalf("Choose(0): %v", err)
	}
	if !r.Done() {
		t.Error("Done() after a choice without goto = false, want true")
	}
}

func TestRunner_RefusedStepsLeaveTheConversationWhereItWas(t *testing.T) {
	// The first choice's action is one the fake does not know, so it fails.
	const file = `{"npc": "fox", "nodes": {
		"ask": {"if": [], "lines": [{"text": "fox.ask.1"}, {"text": "fox.ask.2"}],
			"choices": [
				{"text": "ui.yes", "icon": "check", "do": [{"dance": "jig"}], "goto": "after"},
				{"text": "ui.later", "icon": "cross", "goto": "end"}]},
		"after": {"lines": [{"text": "fox.thanks"}]}}}`
	toQuestion := func(r *dialogue.Runner) error { return r.Advance() }
	toEnd := func(r *dialogue.Runner) error {
		if err := r.Advance(); err != nil {
			return err
		}
		return r.Choose(1)
	}

	tests := []struct {
		name     string
		prepare  func(*dialogue.Runner) error
		act      func(*dialogue.Runner) error
		wantErr  error
		wantLine dialogue.Line
	}{
		{"tapping to continue on the question is refused", toQuestion,
			func(r *dialogue.Runner) error { return r.Advance() },
			dialogue.ErrChoiceRequired, dialogue.Line{Text: "fox.ask.2"}},
		{"a failing action keeps the conversation on the choice", toQuestion,
			func(r *dialogue.Runner) error { return r.Choose(0) },
			errUnknownAction, dialogue.Line{Text: "fox.ask.2"}},
		{"a choice index past the last is refused", toQuestion,
			func(r *dialogue.Runner) error { return r.Choose(2) },
			dialogue.ErrNoSuchChoice, dialogue.Line{Text: "fox.ask.2"}},
		{"a negative choice index is refused", toQuestion,
			func(r *dialogue.Runner) error { return r.Choose(-1) },
			dialogue.ErrNoSuchChoice, dialogue.Line{Text: "fox.ask.2"}},
		{"choosing on a line without choices is refused",
			func(*dialogue.Runner) error { return nil },
			func(r *dialogue.Runner) error { return r.Choose(0) },
			dialogue.ErrNoSuchChoice, dialogue.Line{Text: "fox.ask.1"}},
		{"advancing after the end is refused", toEnd,
			func(r *dialogue.Runner) error { return r.Advance() },
			dialogue.ErrDone, dialogue.Line{}},
		{"choosing after the end is refused", toEnd,
			func(r *dialogue.Runner) error { return r.Choose(0) },
			dialogue.ErrDone, dialogue.Line{}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r, err := dialogue.Start(mustParse(t, file), questStates{}, questStates{})
			if err != nil {
				t.Fatalf("Start: %v", err)
			}
			if err := tc.prepare(r); err != nil {
				t.Fatalf("prepare: %v", err)
			}
			wasDone, hadChoices := r.Done(), len(r.Choices())

			if err := tc.act(r); !errors.Is(err, tc.wantErr) {
				t.Fatalf("err = %v, want %v", err, tc.wantErr)
			}
			if got := r.Line(); got != tc.wantLine {
				t.Errorf("Line() = %+v, want %+v", got, tc.wantLine)
			}
			if r.Done() != wasDone {
				t.Errorf("Done() = %v, want %v", r.Done(), wasDone)
			}
			if got := len(r.Choices()); got != hadChoices {
				t.Errorf("len(Choices()) = %d, want %d", got, hadChoices)
			}
		})
	}
}
