package main

import (
	"errors"
	"fmt"
	"log"

	"github.com/hajimehoshi/ebiten/v2/text/v2"

	"github.com/Elpulgo/critterkingdom/internal/dialogue"
	"github.com/Elpulgo/critterkingdom/internal/engine"
	"github.com/Elpulgo/critterkingdom/internal/i18n"
	"github.com/Elpulgo/critterkingdom/internal/ui"
	"github.com/Elpulgo/critterkingdom/internal/world"
)

// talker opens a villager's dialogue as an overlay on the scene stack when
// the critter starts an interaction with them.
type talker struct {
	font      *text.GoTextFaceSource
	dialogues map[string]*dialogue.Dialogue
	stack     *engine.Stack
}

// interact is the world's interaction sink: it starts a conversation with
// the villager and pushes its dialogue box. A villager without a dialogue
// file, or with nothing to say right now, is logged and opens nothing.
func (t *talker) interact(in world.Interaction) {
	d, ok := t.dialogues[in.NPC]
	if !ok {
		log.Printf("talk to %q: no dialogue file", in.NPC)
		return
	}
	r, err := dialogue.Start(d, noQuests{}, noQuests{})
	if err != nil {
		log.Printf("talk to %q: %v", in.NPC, err)
		return
	}
	t.stack.Push(ui.NewDialogue(t.font, r, in.NPC, i18n.T))
}

// errNoQuests is returned for every dialogue action until quests are wired
// in.
var errNoQuests = errors.New("quest actions are not wired in yet")

// noQuests stands in for the quest system as the dialogues' conditions and
// actions until it is wired in: no condition holds, so a quest-gated node
// stays dormant, and every action fails naming itself, so nothing is
// pretended to have happened.
type noQuests struct{}

// Holds reports false for every condition.
func (noQuests) Holds(dialogue.Condition) (bool, error) {
	return false, nil
}

// Do fails, naming the action.
func (noQuests) Do(a dialogue.Action) error {
	return fmt.Errorf("action %s: %w", a, errNoQuests)
}
