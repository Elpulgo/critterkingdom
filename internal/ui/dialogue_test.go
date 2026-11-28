package ui_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/Elpulgo/critterkingdom/internal/assets"
	"github.com/Elpulgo/critterkingdom/internal/dialogue"
	"github.com/Elpulgo/critterkingdom/internal/engine"
	"github.com/Elpulgo/critterkingdom/internal/geom"
	"github.com/Elpulgo/critterkingdom/internal/ui"
)

// The screen is a 568x320 phone at unit 1, the narrowest the game supports.
// A string-table key fits on one line of it.
const screenW, screenH, unit = 568, 320, 1

// newGame is the quest state of a new game: no `if` condition holds and no
// action may run. It stands in for the quest system, which is not built yet.
type newGame struct{}

func (newGame) Holds(dialogue.Condition) (bool, error) { return false, nil }

func (newGame) Do(a dialogue.Action) error { return errors.New("no action may run: " + string(a)) }

// keys shows every string-table key as itself, so a test can tell which
// line or choice is up.
func keys(key string) string { return key }

// open starts the conversation of the embedded dialogue file for npc and
// returns its overlay, showing text through lookup.
func open(t *testing.T, npc string, lookup func(string) string) *ui.Dialogue {
	t.Helper()
	dialogues, err := assets.LoadDialogues()
	if err != nil {
		t.Fatalf("LoadDialogues: %v", err)
	}
	return start(t, dialogues[npc], lookup)
}

func start(t *testing.T, d *dialogue.Dialogue, lookup func(string) string) *ui.Dialogue {
	t.Helper()
	font, err := assets.LoadFont()
	if err != nil {
		t.Fatalf("LoadFont: %v", err)
	}
	r, err := dialogue.Start(d, newGame{}, newGame{})
	if err != nil {
		t.Fatalf("Start: %v", err)
	}
	return ui.NewDialogue(font, r, d.NPC, lookup)
}

// step runs one tick at 60 ticks per second with the input in.
func step(t *testing.T, d *ui.Dialogue, in engine.Input) {
	t.Helper()
	f := engine.Frame{Input: in, DT: 1.0 / 60, ScreenW: screenW, ScreenH: screenH, Unit: unit}
	if err := d.Update(f); err != nil {
		t.Fatalf("Update: %v", err)
	}
}

// tapAt is a tap at p, in screen pixels.
func tapAt(p geom.Vec) engine.Input { return engine.Input{Tap: true, TapAt: p} }

// corner is a tap far from the dialogue box and its buttons.
var corner = tapAt(geom.Vec{X: 1, Y: 1})

func TestDialogue_TapShowsATypingLineAtOnceAndTheNextTapMovesOn(t *testing.T) {
	d := open(t, "mayor_owl", keys)
	step(t, d, engine.Input{})
	if got := d.Text(); got == "mayor.chat.1" {
		t.Fatalf("after one tick the whole line %q is shown, want it still typing", got)
	}

	step(t, d, corner)
	if got := d.Text(); got != "mayor.chat.1" {
		t.Fatalf("after a tap while typing shows %q, want the whole line %q", got, "mayor.chat.1")
	}

	step(t, d, corner)
	step(t, d, corner)
	if got := d.Text(); got != "mayor.chat.2" {
		t.Fatalf("after tapping on the shown line shows %q, want the next line %q", got, "mayor.chat.2")
	}
	if d.Done() {
		t.Fatal("done before the last line was tapped past")
	}

	step(t, d, corner)
	if !d.Done() {
		t.Fatal("not done after tapping past the last line")
	}
}

// question taps the baker's conversation up to the yes/no question with its
// buttons showing.
func question(t *testing.T) *ui.Dialogue {
	t.Helper()
	d := open(t, "baker_badger", keys)
	step(t, d, corner) // shows baker.chat.1
	step(t, d, corner) // moves on to baker.chat.2, the question
	if got := d.Choices(); len(got) != 0 {
		t.Fatalf("buttons %v show while the question is typing, want none", got)
	}
	step(t, d, corner) // shows the question and its buttons
	if got := d.Text(); got != "baker.chat.2" || len(d.Choices()) != 2 {
		t.Fatalf("shows %q with %d buttons, want the question baker.chat.2 with 2", got, len(d.Choices()))
	}
	return d
}

func TestDialogue_AQuestionStaysUpUntilAButtonIsPicked(t *testing.T) {
	d := question(t)

	step(t, d, corner)
	step(t, d, engine.Input{Paw: true})

	if got := d.Text(); got != "baker.chat.2" || len(d.Choices()) != 2 {
		t.Fatalf("after a tap outside the buttons and a paw, shows %q with %d buttons; want the question with 2", got, len(d.Choices()))
	}
}

func TestDialogue_TheSecondButtonTakesTheSecondChoicesGoto(t *testing.T) {
	cases := []struct {
		name   string
		inputs []engine.Input
	}{
		{"tap on the second button", []engine.Input{tapAt(secondButton())}},
		{"paw after moving right", []engine.Input{{Move: geom.Vec{X: 1}}, {Paw: true}}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			d := question(t)

			for _, in := range tc.inputs {
				step(t, d, in)
			}
			step(t, d, corner) // shows the reply

			if got := d.Text(); got != "baker.no.1" {
				t.Fatalf("shows %q, want the second choice's reply baker.no.1", got)
			}
		})
	}
}

// secondButton is the centre of the second of two choice buttons.
func secondButton() geom.Vec {
	r := ui.LayoutDialogue(screenW, screenH, unit, 2).Choices[1]
	return r.Min.Add(r.Max).Scale(0.5)
}

func TestDialogue_ALineTooLongForTwoLinesIsShownOverTwoPagesBeforeTheNextLine(t *testing.T) {
	// About 95 characters: three lines on the phone, so two pages.
	const long = "Det var en gång en liten igelkott som bodde under en stor ek, " +
		"nära ängen där blåbären växer."
	d, err := dialogue.Parse([]byte(`{"npc": "teller", "nodes": {"start": {"if": [],
		"lines": [{"text": "long", "face": "happy"}, {"text": "short", "face": "happy"}]}}}`))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	lookup := func(key string) string {
		if key == "long" {
			return long
		}
		return key
	}
	o := start(t, d, lookup)

	step(t, o, corner) // shows the first page
	first := o.Text()
	step(t, o, corner) // moves on to the second page
	step(t, o, corner) // shows the second page
	second := o.Text()
	step(t, o, corner) // moves on to the next line
	step(t, o, corner)

	if strings.Count(first, "\n") != 1 {
		t.Errorf("first page %q is not two lines", first)
	}
	if got := strings.ReplaceAll(first+" "+second, "\n", " "); got != long {
		t.Errorf("the two pages read %q, want the whole line %q", got, long)
	}
	if got := o.Text(); got != "short" {
		t.Errorf("after the second page shows %q, want the next line %q", got, "short")
	}
}
