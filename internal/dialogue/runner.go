package dialogue

import (
	"errors"
	"fmt"
)

// ErrNoEntry is returned by Start when no node with an `if` key has all its
// conditions holding, so the villager has nothing to say.
var ErrNoEntry = errors.New("no entry node holds")

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
// portrait face.
func (r *Runner) Line() Line {
	return r.d.nodes[r.node].Lines[r.line]
}
