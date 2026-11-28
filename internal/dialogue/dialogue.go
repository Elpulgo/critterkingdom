// Package dialogue holds villager conversations: the dialogue file format and
// a runner that steps through it. It is pure Go and only parses bytes; reading
// the data files belongs to the caller.
//
// A dialogue file names its npc and holds nodes, kept in file order. A node
// has lines (string-table keys with a portrait face), optional `if`
// conditions and at most two choices, each with an icon, `do` actions and a
// `goto` to the next node. Conditions and actions stay raw JSON: the caller
// decides what they mean.
package dialogue

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
)

// endNode is the reserved goto target that ends the conversation. A file may
// not define a node with this id.
const endNode = "end"

// maxChoices is how many choices a node may offer: two big buttons.
const maxChoices = 2

var (
	// ErrUnknownNode is wrapped by Parse when a choice's goto names a node
	// that is not in the file.
	ErrUnknownNode = errors.New("unknown node")
	// ErrTooManyChoices is wrapped by Parse when a node has more than two
	// choices.
	ErrTooManyChoices = errors.New("too many choices")
)

// Condition is one element of a node's `if` array, kept as its raw JSON
// object, such as {"quest": "missing_hat", "is": "not_started"}.
type Condition json.RawMessage

// UnmarshalJSON keeps a copy of the raw JSON.
func (c *Condition) UnmarshalJSON(data []byte) error {
	*c = append((*c)[:0], data...)
	return nil
}

// Action is one element of a choice's `do` array, kept as its raw JSON
// object, such as {"startQuest": "missing_hat"}.
type Action json.RawMessage

// UnmarshalJSON keeps a copy of the raw JSON.
func (a *Action) UnmarshalJSON(data []byte) error {
	*a = append((*a)[:0], data...)
	return nil
}

// Line is one line a villager says.
type Line struct {
	// Text is a string-table key, such as "grandma.hat.1".
	Text string `json:"text"`
	// Face names the portrait expression, such as "sad".
	Face string `json:"face"`
}

// Choice is one answer button the player can pick.
type Choice struct {
	// Text is a string-table key, such as "ui.yes".
	Text string `json:"text"`
	// Icon names the picture on the button, such as "check".
	Icon string `json:"icon"`
	// Do lists the actions to run when the choice is picked, in order.
	Do []Action `json:"do"`
	// Goto is the node to go to next; "end" or "" ends the conversation.
	Goto string `json:"goto"`
}

// Node is one step of a conversation.
type Node struct {
	// If lists the conditions that must all hold for the conversation to
	// start at this node.
	If []Condition
	// HasIf reports whether the file gave the node an `if` key. Only such
	// nodes can start a conversation; "if": [] always holds.
	HasIf bool
	// Lines are said one after another; a node has at least one.
	Lines []Line
	// Choices are offered on the node's last line; at most two.
	Choices []Choice
}

// UnmarshalJSON decodes a node and records whether its `if` key was present.
func (n *Node) UnmarshalJSON(data []byte) error {
	var raw struct {
		If      *[]Condition `json:"if"`
		Lines   []Line       `json:"lines"`
		Choices []Choice     `json:"choices"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	*n = Node{HasIf: raw.If != nil, Lines: raw.Lines, Choices: raw.Choices}
	if raw.If != nil {
		n.If = *raw.If
	}
	return nil
}

// Dialogue is one villager's dialogue file.
type Dialogue struct {
	// NPC is the villager's id, such as "grandma_hedgehog".
	NPC   string
	order []string
	nodes map[string]Node
}

// UnmarshalJSON decodes a dialogue file, keeping the nodes in file order. A
// node id that appears twice is an error.
func (d *Dialogue) UnmarshalJSON(data []byte) error {
	var raw struct {
		NPC   string          `json:"npc"`
		Nodes json.RawMessage `json:"nodes"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	*d = Dialogue{NPC: raw.NPC, nodes: map[string]Node{}}
	if len(raw.Nodes) == 0 || string(raw.Nodes) == "null" {
		return nil
	}
	dec := json.NewDecoder(bytes.NewReader(raw.Nodes))
	if tok, err := dec.Token(); err != nil {
		return err
	} else if tok != json.Delim('{') {
		return fmt.Errorf("nodes: want an object, got %v", tok)
	}
	for dec.More() {
		tok, err := dec.Token()
		if err != nil {
			return err
		}
		id := tok.(string) // object keys are always strings
		var n Node
		if err := dec.Decode(&n); err != nil {
			return fmt.Errorf("node %q: %w", id, err)
		}
		if _, dup := d.nodes[id]; dup {
			return fmt.Errorf("node %q: duplicate id", id)
		}
		d.order = append(d.order, id)
		d.nodes[id] = n
	}
	return nil
}

// Parse parses a dialogue file and checks its structure, so a runner can
// rely on it: a non-empty npc, at least one node, at least one node with an
// `if` key, no node named "end", no duplicate node ids, at least one line per
// node, no empty line or choice text, at most two choices per node, and every
// choice's goto naming a node in the file, "end", or nothing.
func Parse(data []byte) (*Dialogue, error) {
	var d Dialogue
	if err := json.Unmarshal(data, &d); err != nil {
		return nil, fmt.Errorf("parse dialogue: %w", err)
	}
	if err := d.validate(); err != nil {
		return nil, fmt.Errorf("parse dialogue %q: %w", d.NPC, err)
	}
	return &d, nil
}

func (d *Dialogue) validate() error {
	switch {
	case d.NPC == "":
		return errors.New("empty npc")
	case len(d.order) == 0:
		return errors.New("no nodes")
	}
	if _, ok := d.nodes[endNode]; ok {
		return fmt.Errorf("node %q: the id is reserved for ending a conversation", endNode)
	}
	entry := false
	for _, id := range d.order {
		n := d.nodes[id]
		entry = entry || n.HasIf
		if err := d.validateNode(n); err != nil {
			return fmt.Errorf("node %q: %w", id, err)
		}
	}
	if !entry {
		return errors.New("no node with an if key to start at")
	}
	return nil
}

func (d *Dialogue) validateNode(n Node) error {
	if len(n.Lines) == 0 {
		return errors.New("no lines")
	}
	for i, l := range n.Lines {
		if l.Text == "" {
			return fmt.Errorf("line %d: empty text", i)
		}
	}
	if len(n.Choices) > maxChoices {
		return fmt.Errorf("%d choices, at most %d: %w", len(n.Choices), maxChoices, ErrTooManyChoices)
	}
	for i, c := range n.Choices {
		if c.Text == "" {
			return fmt.Errorf("choice %d: empty text", i)
		}
		if c.Goto == "" || c.Goto == endNode {
			continue
		}
		if _, ok := d.nodes[c.Goto]; !ok {
			return fmt.Errorf("choice %d: goto %q: %w", i, c.Goto, ErrUnknownNode)
		}
	}
	return nil
}

// NodeIDs returns the node ids in file order. The slice is a copy.
func (d *Dialogue) NodeIDs() []string {
	return slices.Clone(d.order)
}

// Node returns the node with the given id, and false when the file has no
// such node.
func (d *Dialogue) Node(id string) (Node, bool) {
	n, ok := d.nodes[id]
	return n, ok
}
