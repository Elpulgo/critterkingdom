package quest

import "fmt"

// Status is where a quest stands for the player.
type Status int

// The statuses a quest moves through, in order.
const (
	// NotStarted: nobody has handed the quest to the player yet.
	NotStarted Status = iota
	// Active: the player is working on one of the quest's steps.
	Active
	// Done: every step is finished and the reward was handed out.
	Done
)

// ChangeKind says how a quest moved on an event.
type ChangeKind int

// The ways a quest can move on one event.
const (
	// Progressed: the counter of a counted step went up, the step is not
	// finished yet.
	Progressed ChangeKind = iota + 1
	// StepAdvanced: a step was finished and the quest is now on the next one.
	StepAdvanced
	// Completed: the last step was finished; the change carries the reward.
	Completed
)

// Change is how one quest moved on one event.
type Change struct {
	Quest string
	Kind  ChangeKind
	// Step is the index of the step the quest is on after the event (the last
	// step for Completed).
	Step int
	// Count and Of are that step's counter and how many it needs.
	Count, Of int
	// Reward is the quest's reward on a Completed change, empty otherwise.
	Reward Reward
}

// progress is the runtime state of one quest.
type progress struct {
	status Status
	step   int
	count  int
}

// Log is the player's quest log: a fixed set of quest definitions and where
// each of them stands. Feed it events with Handle and it reports how the
// quests moved.
type Log struct {
	defs  []Definition
	state []progress
	index map[string]int
}

// NewLog returns a Log over defs with every quest not started. The ids in
// defs are expected to be unique, as LoadDir returns them; a later definition
// with an id seen before is ignored.
func NewLog(defs []Definition) *Log {
	l := &Log{index: make(map[string]int, len(defs))}
	for _, d := range defs {
		if _, ok := l.index[d.ID]; ok {
			continue
		}
		l.index[d.ID] = len(l.defs)
		l.defs = append(l.defs, d)
	}
	l.state = make([]progress, len(l.defs))
	return l
}

// Start makes a not-started quest active on its first step with its counter
// at 0. An unknown id, or a quest that is already active or done, is an error
// and changes nothing.
func (l *Log) Start(id string) error {
	i, ok := l.index[id]
	if !ok {
		return fmt.Errorf("quest %q: unknown quest", id)
	}
	switch l.state[i].status {
	case Active:
		return fmt.Errorf("quest %q: already active", id)
	case Done:
		return fmt.Errorf("quest %q: already done", id)
	}
	l.state[i] = progress{status: Active}
	return nil
}

// Handle feeds one event to the current step of every active quest and
// returns, in definition order, one Change per quest that moved. A quest
// moves at most one step per event; not-started and done quests, and steps
// other than the current one, ignore the event.
func (l *Log) Handle(e Event) []Change {
	var changes []Change
	for i, d := range l.defs {
		p := &l.state[i]
		if p.status != Active {
			continue
		}
		step := d.Steps[p.step]
		if !step.matches(e) {
			continue
		}
		need := step.need()
		p.count = min(p.count+e.amount(), need)
		switch {
		case p.count < need:
			changes = append(changes, Change{Quest: d.ID, Kind: Progressed, Step: p.step, Count: p.count, Of: need})
		case p.step+1 < len(d.Steps):
			p.step++
			p.count = 0
			changes = append(changes, Change{Quest: d.ID, Kind: StepAdvanced, Step: p.step, Count: 0, Of: d.Steps[p.step].need()})
		default:
			p.status = Done
			changes = append(changes, Change{Quest: d.ID, Kind: Completed, Step: p.step, Count: need, Of: need, Reward: d.Reward})
		}
	}
	return changes
}

// Status reports where quest id stands. An unknown id is NotStarted.
func (l *Log) Status(id string) Status {
	i, ok := l.index[id]
	if !ok {
		return NotStarted
	}
	return l.state[i].status
}

// Progress reports the current step of an active quest, its counter and how
// many the step needs (1 on a step that is not counted). ok is false when the
// quest is unknown or not active.
func (l *Log) Progress(id string) (step, count, of int, ok bool) {
	i, found := l.index[id]
	if !found || l.state[i].status != Active {
		return 0, 0, 0, false
	}
	p := l.state[i]
	return p.step, p.count, l.defs[i].Steps[p.step].need(), true
}

// Active lists the ids of the active quests in definition order.
func (l *Log) Active() []string {
	var ids []string
	for i, d := range l.defs {
		if l.state[i].status == Active {
			ids = append(ids, d.ID)
		}
	}
	return ids
}
