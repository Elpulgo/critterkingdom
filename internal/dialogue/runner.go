package dialogue

import (
	"errors"
	"fmt"
)

var (
	// ErrNoEntry is returned by Start when no node with an `if` key has all
	// its conditions holding, so the villager has nothing to say.
	ErrNoEntry = errors.New("no entry node holds")
	// ErrChoiceRequired is returned by Advance on a step that offers
	// choices: the player has to pick one.
	ErrChoiceRequired = errors.New("a choice is required")
	// ErrNoSuchChoice is returned by Choose on a step without choices or
	// for an index out of range.
	ErrNoSuchChoice = errors.New("no such choice")
	// ErrDone is returned by Advance and Choose once the conversation has
	// ended.
	ErrDone = errors.New("conversation is over")
)

// Conditions decides whether a node's `if` conditions hold. The runner never
// looks inside a condition: it hands each one over as the raw JSON from the
// file, and the caller decides what it means.
type Conditions interface {
	Holds(Condition) (bool, error)
}

// Actions runs a choice's `do` actions. The runner hands each one over as
// the raw JSON from the file.
type Actions interface {
	Do(Action) error
}

// Runner steps through one conversation of a dialogue.
type Runner struct {
	d       *Dialogue
	actions Actions
	node    string
	line    int
	done    bool
}

// Start begins a conversation at the first node, in file order, that has an
// `if` key and whose conditions all hold; "if": [] always holds. Nodes
// without an `if` key are only reached by a choice's goto. It returns
// ErrNoEntry when no such node holds, and a wrapped error when checking a
// condition fails.
func Start(d *Dialogue, c Conditions, a Actions) (*Runner, error) {
	for _, id := range d.order {
		n := d.nodes[id]
		if !n.HasIf {
			continue
		}
		ok, err := allHold(c, n.If)
		if err != nil {
			return nil, fmt.Errorf("start dialogue %q: node %q: %w", d.NPC, id, err)
		}
		if ok {
			return &Runner{d: d, actions: a, node: id}, nil
		}
	}
	return nil, fmt.Errorf("start dialogue %q: %w", d.NPC, ErrNoEntry)
}

// allHold reports whether every condition holds, stopping at the first that
// does not.
func allHold(c Conditions, conds []Condition) (bool, error) {
	for i, cond := range conds {
		ok, err := c.Holds(cond)
		if err != nil {
			return false, fmt.Errorf("condition %d: %w", i, err)
		}
		if !ok {
			return false, nil
		}
	}
	return true, nil
}

// Line returns the line the villager is saying now: its string-table key and
// portrait face. It returns the zero Line once the conversation has ended.
func (r *Runner) Line() Line {
	if r.done {
		return Line{}
	}
	return r.d.nodes[r.node].Lines[r.line]
}

// Choices returns the answers offered on this step. It is empty except on
// the last line of a node that has choices, so the question and its buttons
// come together.
func (r *Runner) Choices() []Choice {
	if r.done {
		return nil
	}
	n := r.d.nodes[r.node]
	if r.line != len(n.Lines)-1 || len(n.Choices) == 0 {
		return nil
	}
	return append([]Choice(nil), n.Choices...)
}

// Done reports whether the conversation has ended.
func (r *Runner) Done() bool {
	return r.done
}

// Advance moves on to the node's next line, or ends the conversation after
// the last line of a node without choices. It returns ErrChoiceRequired on
// a step that offers choices and ErrDone after the end, and does not move.
func (r *Runner) Advance() error {
	if r.done {
		return ErrDone
	}
	if len(r.Choices()) > 0 {
		return ErrChoiceRequired
	}
	if r.line < len(r.d.nodes[r.node].Lines)-1 {
		r.line++
		return nil
	}
	r.done = true
	return nil
}

// Choose picks choice i on this step: it runs the choice's actions in order,
// then moves to the first line of its goto node, or ends the conversation
// for "end" or an omitted goto. The goto node's `if` is not checked. It
// returns ErrDone after the end and ErrNoSuchChoice on a step without
// choices or for an index out of range. When an action fails, Choose
// returns the wrapped error and stays on the same step; actions that
// already ran are not undone.
func (r *Runner) Choose(i int) error {
	if r.done {
		return ErrDone
	}
	choices := r.Choices()
	if i < 0 || i >= len(choices) {
		return fmt.Errorf("choose %d in dialogue %q: node %q: %w", i, r.d.NPC, r.node, ErrNoSuchChoice)
	}
	ch := choices[i]
	for j, act := range ch.Do {
		if err := r.actions.Do(act); err != nil {
			return fmt.Errorf("choose %d in dialogue %q: node %q: action %d: %w", i, r.d.NPC, r.node, j, err)
		}
	}
	if ch.Goto == "" || ch.Goto == endNode {
		r.done = true
		return nil
	}
	r.node, r.line = ch.Goto, 0
	return nil
}
