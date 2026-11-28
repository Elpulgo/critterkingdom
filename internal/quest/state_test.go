package quest_test

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/Elpulgo/critterkingdom/internal/quest"
)

// saveDefs are three quests for the save tests: one the player never starts,
// one left mid-counter on its second step, and one finished.
var saveDefs = []quest.Definition{
	{ID: "not_yet", Steps: []quest.Step{{Type: quest.Talk, NPC: "owl"}}},
	{ID: "berry_pie", Steps: []quest.Step{
		{Type: quest.Talk, NPC: "mole"},
		{Type: quest.Collect, Item: "berry", Count: 5},
		{Type: quest.Talk, NPC: "mole"},
	}},
	{ID: "lost_key", Steps: []quest.Step{{Type: quest.Find, Target: "lost_key"}},
		Reward: quest.Reward{Sticker: "key_keeper"}},
}

// roundTrip marshals s to JSON and back, as a save and load would.
func roundTrip(t *testing.T, s quest.State) (quest.State, string) {
	t.Helper()
	data, err := json.Marshal(s)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	var back quest.State
	if err := json.Unmarshal(data, &back); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	return back, string(data)
}

func TestState_SurvivesAJSONRoundTripMidCounterAndPlayContinues(t *testing.T) {
	l := startedLog(t, saveDefs, "berry_pie", "lost_key")
	l.Handle(quest.Event{Kind: quest.TalkedTo, ID: "mole"})
	l.Handle(quest.Event{Kind: quest.ItemCollected, ID: "berry"})
	l.Handle(quest.Event{Kind: quest.ItemCollected, ID: "berry"})
	l.Handle(quest.Event{Kind: quest.HiddenFound, ID: "lost_key"})

	saved, data := roundTrip(t, l.Snapshot())
	const wantJSON = `{"quests":{"berry_pie":{"status":"active","step":1,"count":2},"lost_key":{"status":"done"}}}`
	if data != wantJSON {
		t.Errorf("saved JSON = %s, want %s", data, wantJSON)
	}

	r, err := quest.Restore(saveDefs, saved)
	if err != nil {
		t.Fatalf("Restore: %v", err)
	}
	wantStatus := map[string]quest.Status{"not_yet": quest.NotStarted, "berry_pie": quest.Active, "lost_key": quest.Done}
	wantProgress := map[string]progress{"not_yet": {}, "berry_pie": {1, 2, 5, true}, "lost_key": {}}
	for id := range wantStatus {
		if got := r.Status(id); got != wantStatus[id] {
			t.Errorf("restored Status(%s) = %v, want %v", id, got, wantStatus[id])
		}
		if got := progressOf(r, id); got != wantProgress[id] {
			t.Errorf("restored Progress(%s) = %+v, want %+v", id, got, wantProgress[id])
		}
	}

	got := r.Handle(quest.Event{Kind: quest.ItemCollected, ID: "berry"})
	want := []quest.Change{{Quest: "berry_pie", Kind: quest.Progressed, Step: 1, Count: 3, Of: 5}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("next berry after restore: Handle = %+v, want %+v", got, want)
	}
}

func TestRestore_DoneQuest_DoesNotGrantItsRewardAgain(t *testing.T) {
	def, err := quest.Parse([]byte(missingHatJSON))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	var saved quest.State
	if err := json.Unmarshal([]byte(`{"quests":{"missing_hat":{"status":"done"}}}`), &saved); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	l, err := quest.Restore([]quest.Definition{def}, saved)
	if err != nil {
		t.Fatalf("Restore: %v", err)
	}

	for _, e := range []quest.Event{
		{Kind: quest.EnteredArea, ID: "meadow"},
		{Kind: quest.ItemCollected, ID: "grandmas_hat"},
		{Kind: quest.TalkedTo, ID: "grandma_hedgehog"},
	} {
		if got := l.Handle(e); len(got) != 0 {
			t.Errorf("Handle(%+v) on a restored done quest = %+v, want no change", e, got)
		}
	}
	if err := l.Start("missing_hat"); err == nil {
		t.Error("Start of a restored done quest: want an error")
	}
	if got := l.Status("missing_hat"); got != quest.Done {
		t.Errorf("Status = %v, want Done", got)
	}
}

func TestRestore_RejectsStateThatDoesNotFitTheQuests(t *testing.T) {
	tests := []struct {
		name  string
		entry quest.QuestState
	}{
		{"unknown status", quest.QuestState{Status: "finished"}},
		{"not started is not a saved status", quest.QuestState{Status: "not_started"}},
		{"negative step", quest.QuestState{Status: "active", Step: -1}},
		{"step past the last", quest.QuestState{Status: "active", Step: 3}},
		{"counter at the step's count", quest.QuestState{Status: "active", Step: 1, Count: 5}},
		{"negative counter", quest.QuestState{Status: "active", Step: 1, Count: -1}},
		{"counter on an uncounted step", quest.QuestState{Status: "active", Step: 0, Count: 1}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			s := quest.State{Quests: map[string]quest.QuestState{"berry_pie": tc.entry}}
			if l, err := quest.Restore(saveDefs, s); err == nil {
				t.Errorf("Restore(%+v) = %v, want an error", tc.entry, l)
			}
		})
	}
}

func TestRestore_DropsTheStateOfAQuestNoLongerInTheData(t *testing.T) {
	s := quest.State{Quests: map[string]quest.QuestState{
		"removed_quest": {Status: "active", Step: 7, Count: 3},
		"berry_pie":     {Status: "active", Step: 1, Count: 2},
	}}
	l, err := quest.Restore(saveDefs, s)
	if err != nil {
		t.Fatalf("Restore: %v", err)
	}
	if got, want := l.Snapshot(), (quest.State{Quests: map[string]quest.QuestState{
		"berry_pie": {Status: "active", Step: 1, Count: 2},
	}}); !reflect.DeepEqual(got, want) {
		t.Errorf("Snapshot after Restore = %+v, want %+v", got, want)
	}
}
