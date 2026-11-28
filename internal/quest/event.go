package quest

// EventKind says what happened in the game. Each step type listens to exactly
// one kind.
type EventKind string

// The things that can happen in the game that a quest step listens for.
const (
	// TalkedTo: the player talked to NPC ID. Advances a talk step.
	TalkedTo EventKind = "talked_to"
	// EnteredArea: the player walked into area ID. Advances a goto step.
	EnteredArea EventKind = "entered_area"
	// ItemCollected: the player picked up N of item ID. Advances a collect
	// step.
	ItemCollected EventKind = "item_collected"
	// ItemDelivered: the player handed N of item ID to NPC To. Advances a
	// deliver step.
	ItemDelivered EventKind = "item_delivered"
	// Befriended: the player befriended a critter of species ID. Advances a
	// befriend step.
	Befriended EventKind = "befriended"
	// HiddenFound: the player found hidden thing ID. Advances a find step.
	HiddenFound EventKind = "hidden_found"
	// PuzzleSolved: the player solved puzzle room ID. Advances a puzzle step.
	PuzzleSolved EventKind = "puzzle_solved"
	// EnemyDefeated: the player calmed N grumpies of kind ID. Advances a
	// defeat step.
	EnemyDefeated EventKind = "enemy_defeated"
	// RoomCleared: the player cleared room ID of grumpies. Advances a clear
	// step.
	RoomCleared EventKind = "room_cleared"
)

// Event is something that happened in the game, fed to Log.Handle.
type Event struct {
	Kind EventKind
	// ID is the npc, area, item, critter, hidden thing, room or enemy id the
	// event is about.
	ID string
	// To is the NPC who received the item of an ItemDelivered event.
	To string
	// N is how many, for ItemCollected, ItemDelivered and EnemyDefeated; 0
	// means 1. Other kinds ignore it.
	N int
}

// amount is how many the event counts for, at least 1.
func (e Event) amount() int {
	if e.N < 1 {
		return 1
	}
	return e.N
}

// matches reports whether e is the event that moves step s: the kind that
// step type listens to, with the step's id (and, for deliver, the receiving
// NPC).
func (s Step) matches(e Event) bool {
	switch s.Type {
	case Talk:
		return e.Kind == TalkedTo && e.ID == s.NPC
	case GoTo:
		return e.Kind == EnteredArea && e.ID == s.Target
	case Collect:
		return e.Kind == ItemCollected && e.ID == s.Item
	case Deliver:
		return e.Kind == ItemDelivered && e.ID == s.Item && e.To == s.NPC
	case Befriend:
		return e.Kind == Befriended && e.ID == s.Critter
	case Find:
		return e.Kind == HiddenFound && e.ID == s.Target
	case Puzzle:
		return e.Kind == PuzzleSolved && e.ID == s.Room
	case Defeat:
		return e.Kind == EnemyDefeated && e.ID == s.Enemy
	case Clear:
		return e.Kind == RoomCleared && e.ID == s.Room
	}
	return false
}

// need is how many matching events (counting N) finish the step: its count
// on counted steps, 1 on every other step.
func (s Step) need() int {
	if s.Count < 1 {
		return 1
	}
	return s.Count
}
