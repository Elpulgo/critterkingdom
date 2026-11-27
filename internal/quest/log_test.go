package quest_test

import (
	"reflect"
	"testing"

	"github.com/Elpulgo/critterkingdom/internal/quest"
)

// afterwards is a closing step so the step under test is not the last one.
var afterwards = quest.Step{Type: quest.Talk, NPC: "afterwards_npc"}

// startedLog returns a Log over defs with the quests in start already started.
func startedLog(t *testing.T, defs []quest.Definition, start ...string) *quest.Log {
	t.Helper()
	l := quest.NewLog(defs)
	for _, id := range start {
		if err := l.Start(id); err != nil {
			t.Fatalf("Start(%q): %v", id, err)
		}
	}
	return l
}

type progress struct {
	step, count, of int
	ok              bool
}

func progressOf(l *quest.Log, id string) progress {
	step, count, of, ok := l.Progress(id)
	return progress{step, count, of, ok}
}

func TestLog_EachStepType_AdvancesOnItsEventAndIgnoresAllOthers(t *testing.T) {
	allKinds := []quest.EventKind{
		quest.TalkedTo, quest.EnteredArea, quest.ItemCollected, quest.ItemDelivered,
		quest.Befriended, quest.HiddenFound, quest.PuzzleSolved, quest.EnemyDefeated,
		quest.RoomCleared,
	}
	tests := []struct {
		name  string
		step  quest.Step
		right quest.Event
	}{
		{"talk", quest.Step{Type: quest.Talk, NPC: "mole"}, quest.Event{Kind: quest.TalkedTo, ID: "mole"}},
		{"goto", quest.Step{Type: quest.GoTo, Target: "meadow"}, quest.Event{Kind: quest.EnteredArea, ID: "meadow"}},
		{"collect", quest.Step{Type: quest.Collect, Item: "berry", Count: 1}, quest.Event{Kind: quest.ItemCollected, ID: "berry"}},
		{"deliver", quest.Step{Type: quest.Deliver, Item: "plank", NPC: "beaver", Count: 1}, quest.Event{Kind: quest.ItemDelivered, ID: "plank", To: "beaver"}},
		{"befriend", quest.Step{Type: quest.Befriend, Critter: "firefly"}, quest.Event{Kind: quest.Befriended, ID: "firefly"}},
		{"find", quest.Step{Type: quest.Find, Target: "lost_key"}, quest.Event{Kind: quest.HiddenFound, ID: "lost_key"}},
		{"puzzle", quest.Step{Type: quest.Puzzle, Room: "mirror_room"}, quest.Event{Kind: quest.PuzzleSolved, ID: "mirror_room"}},
		{"defeat", quest.Step{Type: quest.Defeat, Enemy: "sulky_slime", Count: 1}, quest.Event{Kind: quest.EnemyDefeated, ID: "sulky_slime"}},
		{"clear", quest.Step{Type: quest.Clear, Room: "cellar"}, quest.Event{Kind: quest.RoomCleared, ID: "cellar"}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			defs := []quest.Definition{{ID: "q", Steps: []quest.Step{tc.step, afterwards}}}
			l := startedLog(t, defs, "q")

			// Every other kind, carrying the step's own ids, and the right
			// kind with a wrong id leave the step where it is.
			var wrong []quest.Event
			for _, k := range allKinds {
				if k != tc.right.Kind {
					wrong = append(wrong, quest.Event{Kind: k, ID: tc.right.ID, To: tc.right.To})
				}
			}
			wrongID := tc.right
			wrongID.ID = "something_else"
			wrong = append(wrong, wrongID)
			for _, e := range wrong {
				if got := l.Handle(e); len(got) != 0 {
					t.Errorf("Handle(%+v) = %+v, want no change", e, got)
				}
				if got, want := progressOf(l, "q"), (progress{0, 0, 1, true}); got != want {
					t.Errorf("after %+v Progress = %+v, want %+v", e, got, want)
				}
			}

			got := l.Handle(tc.right)
			want := []quest.Change{{Quest: "q", Kind: quest.StepAdvanced, Step: 1, Count: 0, Of: 1}}
			if !reflect.DeepEqual(got, want) {
				t.Errorf("Handle(%+v) = %+v, want %+v", tc.right, got, want)
			}
			if got, want := progressOf(l, "q"), (progress{1, 0, 1, true}); got != want {
				t.Errorf("Progress = %+v, want %+v", got, want)
			}
		})
	}
}

func TestLog_DeliverStep_NeedsTheRightItemAndTheRightReceiver(t *testing.T) {
	defs := []quest.Definition{{ID: "q", Steps: []quest.Step{
		{Type: quest.Deliver, Item: "plank", NPC: "beaver", Count: 1}, afterwards,
	}}}
	l := startedLog(t, defs, "q")

	for _, e := range []quest.Event{
		{Kind: quest.ItemDelivered, ID: "plank", To: "owl"},
		{Kind: quest.ItemDelivered, ID: "acorn", To: "beaver"},
		{Kind: quest.ItemDelivered, ID: "plank"},
	} {
		if got := l.Handle(e); len(got) != 0 {
			t.Errorf("Handle(%+v) = %+v, want no change", e, got)
		}
	}
	if got, want := progressOf(l, "q"), (progress{0, 0, 1, true}); got != want {
		t.Errorf("Progress = %+v, want %+v", got, want)
	}
}

func TestLog_CountedSteps_CountUpCapAtTheirCountAndTheNextStepStartsAtZero(t *testing.T) {
	tests := []struct {
		name  string
		steps []quest.Step
		event quest.Event
	}{
		{
			name: "collect",
			steps: []quest.Step{
				{Type: quest.Collect, Item: "berry", Count: 5},
				{Type: quest.Collect, Item: "berry", Count: 3},
				{Type: quest.Collect, Item: "berry", Count: 2},
			},
			event: quest.Event{Kind: quest.ItemCollected, ID: "berry"},
		},
		{
			name: "deliver",
			steps: []quest.Step{
				{Type: quest.Deliver, Item: "plank", NPC: "beaver", Count: 5},
				{Type: quest.Deliver, Item: "plank", NPC: "beaver", Count: 3},
				{Type: quest.Deliver, Item: "plank", NPC: "beaver", Count: 2},
			},
			event: quest.Event{Kind: quest.ItemDelivered, ID: "plank", To: "beaver"},
		},
		{
			name: "defeat",
			steps: []quest.Step{
				{Type: quest.Defeat, Enemy: "sulky_slime", Count: 5},
				{Type: quest.Defeat, Enemy: "sulky_slime", Count: 3},
				{Type: quest.Defeat, Enemy: "sulky_slime", Count: 2},
			},
			event: quest.Event{Kind: quest.EnemyDefeated, ID: "sulky_slime"},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			defs := []quest.Definition{{ID: "q", Steps: append(tc.steps, afterwards)}}
			l := startedLog(t, defs, "q")
			n := func(n int) quest.Event { e := tc.event; e.N = n; return e }

			plays := []struct {
				event quest.Event
				want  quest.Change
			}{
				// N of 0 counts as one.
				{n(0), quest.Change{Quest: "q", Kind: quest.Progressed, Step: 0, Count: 1, Of: 5}},
				{n(1), quest.Change{Quest: "q", Kind: quest.Progressed, Step: 0, Count: 2, Of: 5}},
			}
			for _, p := range plays {
				got := l.Handle(p.event)
				if want := []quest.Change{p.want}; !reflect.DeepEqual(got, want) {
					t.Fatalf("Handle(%+v) = %+v, want %+v", p.event, got, want)
				}
			}
			if got, want := progressOf(l, "q"), (progress{0, 2, 5, true}); got != want {
				t.Fatalf("Progress after two = %+v, want %+v", got, want)
			}

			plays = []struct {
				event quest.Event
				want  quest.Change
			}{
				{n(2), quest.Change{Quest: "q", Kind: quest.Progressed, Step: 0, Count: 4, Of: 5}},
				// One left but three handed in: the step finishes and the
				// extra two do not carry over.
				{n(3), quest.Change{Quest: "q", Kind: quest.StepAdvanced, Step: 1, Count: 0, Of: 3}},
				// Ten on a step of three finishes only that step.
				{n(10), quest.Change{Quest: "q", Kind: quest.StepAdvanced, Step: 2, Count: 0, Of: 2}},
			}
			for _, p := range plays {
				got := l.Handle(p.event)
				if want := []quest.Change{p.want}; !reflect.DeepEqual(got, want) {
					t.Fatalf("Handle(%+v) = %+v, want %+v", p.event, got, want)
				}
			}
			if got, want := progressOf(l, "q"), (progress{2, 0, 2, true}); got != want {
				t.Errorf("Progress after the oversized events = %+v, want %+v", got, want)
			}
		})
	}
}

func TestLog_EventsOnlyReachTheCurrentStepOfActiveQuests(t *testing.T) {
	defs := []quest.Definition{{ID: "hat", Steps: []quest.Step{
		{Type: quest.GoTo, Target: "meadow"},
		{Type: quest.Collect, Item: "grandmas_hat", Count: 1},
		{Type: quest.Talk, NPC: "grandma_hedgehog"},
	}}}
	l := quest.NewLog(defs)
	meadow := quest.Event{Kind: quest.EnteredArea, ID: "meadow"}
	hat := quest.Event{Kind: quest.ItemCollected, ID: "grandmas_hat"}
	grandma := quest.Event{Kind: quest.TalkedTo, ID: "grandma_hedgehog"}

	if got := l.Handle(meadow); len(got) != 0 {
		t.Errorf("not started: Handle(meadow) = %+v, want no change", got)
	}
	if got := l.Status("hat"); got != quest.NotStarted {
		t.Errorf("not started: Status = %v, want NotStarted", got)
	}

	if err := l.Start("hat"); err != nil {
		t.Fatalf("Start: %v", err)
	}
	for _, e := range []quest.Event{hat, grandma} {
		if got := l.Handle(e); len(got) != 0 {
			t.Errorf("on step 0: Handle(%+v) of a later step = %+v, want no change", e, got)
		}
	}
	if got, want := progressOf(l, "hat"), (progress{0, 0, 1, true}); got != want {
		t.Errorf("on step 0: Progress = %+v, want %+v", got, want)
	}

	if got := l.Handle(meadow); len(got) != 1 || got[0].Kind != quest.StepAdvanced {
		t.Fatalf("Handle(meadow) = %+v, want one StepAdvanced", got)
	}
	for _, e := range []quest.Event{meadow, grandma} {
		if got := l.Handle(e); len(got) != 0 {
			t.Errorf("on step 1: Handle(%+v) of another step = %+v, want no change", e, got)
		}
	}
	if got, want := progressOf(l, "hat"), (progress{1, 0, 1, true}); got != want {
		t.Errorf("on step 1: Progress = %+v, want %+v", got, want)
	}
}

func TestLog_OneEventMovesEveryMatchingQuestByOneStepInDefinitionOrder(t *testing.T) {
	owl := quest.Step{Type: quest.Talk, NPC: "owl"}
	defs := []quest.Definition{
		{ID: "letters", Steps: []quest.Step{owl, owl, afterwards}},
		{ID: "not_yet", Steps: []quest.Step{owl, afterwards}},
		{ID: "stars", Steps: []quest.Step{owl, {Type: quest.Find, Target: "star"}}},
	}
	// Started out of definition order on purpose.
	l := startedLog(t, defs, "stars", "letters")

	got := l.Handle(quest.Event{Kind: quest.TalkedTo, ID: "owl"})
	want := []quest.Change{
		{Quest: "letters", Kind: quest.StepAdvanced, Step: 1, Count: 0, Of: 1},
		{Quest: "stars", Kind: quest.StepAdvanced, Step: 1, Count: 0, Of: 1},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Handle(talked to owl) = %+v, want %+v", got, want)
	}
	if got, want := progressOf(l, "letters"), (progress{1, 0, 1, true}); got != want {
		t.Errorf("letters Progress = %+v, want %+v (one step, not two)", got, want)
	}
	if got, want := l.Active(), []string{"letters", "stars"}; !reflect.DeepEqual(got, want) {
		t.Errorf("Active = %v, want %v", got, want)
	}
}

func TestLog_Start_MakesANewQuestActiveAndRejectsUnknownOrActiveQuests(t *testing.T) {
	defs := []quest.Definition{
		{ID: "hat", Steps: []quest.Step{{Type: quest.Collect, Item: "berry", Count: 5}, afterwards}},
		{ID: "pie", Steps: []quest.Step{afterwards}},
	}
	l := quest.NewLog(defs)
	for _, id := range []string{"hat", "pie"} {
		if got := l.Status(id); got != quest.NotStarted {
			t.Errorf("new log: Status(%q) = %v, want NotStarted", id, got)
		}
	}
	if got := l.Active(); len(got) != 0 {
		t.Errorf("new log: Active = %v, want none", got)
	}

	if err := l.Start("hat"); err != nil {
		t.Fatalf("Start(hat): %v", err)
	}
	if got := l.Status("hat"); got != quest.Active {
		t.Errorf("Status(hat) = %v, want Active", got)
	}
	if got, want := progressOf(l, "hat"), (progress{0, 0, 5, true}); got != want {
		t.Errorf("Progress(hat) = %+v, want %+v", got, want)
	}

	l.Handle(quest.Event{Kind: quest.ItemCollected, ID: "berry", N: 2})
	if err := l.Start("hat"); err == nil {
		t.Error("Start of an active quest: want an error")
	}
	if got, want := progressOf(l, "hat"), (progress{0, 2, 5, true}); got != want {
		t.Errorf("Progress(hat) after the second Start = %+v, want %+v", got, want)
	}

	if err := l.Start("no_such_quest"); err == nil {
		t.Error("Start of an unknown quest: want an error")
	}
	if got, want := l.Active(), []string{"hat"}; !reflect.DeepEqual(got, want) {
		t.Errorf("Active = %v, want %v", got, want)
	}
}
