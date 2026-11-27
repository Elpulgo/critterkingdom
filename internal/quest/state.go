package quest

import "fmt"

// State is the saveable runtime state of a Log: where every started quest
// stands, keyed by quest id. Quests that are not started are left out. It is
// a plain JSON value meant to be embedded in a save game:
//
//	{"quests": {"berry_pie": {"status": "active", "step": 1, "count": 2},
//	            "missing_hat": {"status": "done"}}}
type State struct {
	Quests map[string]QuestState `json:"quests"`
}

// QuestState is where one started quest stands.
type QuestState struct {
	// Status is "active" or "done".
	Status string `json:"status"`
	// Step is the index of the current step of an active quest.
	Step int `json:"step,omitempty"`
	// Count is the counter of that step, from 0 to one less than it needs.
	Count int `json:"count,omitempty"`
}

// The status strings of QuestState. Strings rather than the Status numbers
// keep a save readable and independent of the order of the constants.
const (
	statusActive = "active"
	statusDone   = "done"
)

// Snapshot returns the state of every started quest. A done quest records only
// its status.
func (l *Log) Snapshot() State {
	s := State{Quests: make(map[string]QuestState)}
	for i, d := range l.defs {
		switch p := l.state[i]; p.status {
		case Active:
			s.Quests[d.ID] = QuestState{Status: statusActive, Step: p.step, Count: p.count}
		case Done:
			s.Quests[d.ID] = QuestState{Status: statusDone}
		}
	}
	return s
}

// Restore returns a Log over defs with the quests in s set to where s says
// they stand and every other quest not started. Restoring a done quest does
// not report its reward again. An unknown status, a step index outside an
// active quest's steps, or a counter outside 0 to one less than what the
// active step needs is an error. An entry for a quest id that is not in defs
// is dropped, so removing a quest from the data does not break old saves.
func Restore(defs []Definition, s State) (*Log, error) {
	l := NewLog(defs)
	for id, qs := range s.Quests {
		i, ok := l.index[id]
		if !ok {
			continue
		}
		switch qs.Status {
		case statusDone:
			l.state[i] = progress{status: Done}
		case statusActive:
			steps := l.defs[i].Steps
			if qs.Step < 0 || qs.Step >= len(steps) {
				return nil, fmt.Errorf("quest %q: step %d outside its %d steps", id, qs.Step, len(steps))
			}
			if need := steps[qs.Step].need(); qs.Count < 0 || qs.Count >= need {
				return nil, fmt.Errorf("quest %q: step %d: count %d outside 0..%d", id, qs.Step, qs.Count, need-1)
			}
			l.state[i] = progress{status: Active, step: qs.Step, count: qs.Count}
		default:
			return nil, fmt.Errorf("quest %q: unknown status %q", id, qs.Status)
		}
	}
	return l, nil
}
