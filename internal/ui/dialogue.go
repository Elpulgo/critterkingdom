package ui

import (
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"

	"github.com/Elpulgo/critterkingdom/internal/dialogue"
	"github.com/Elpulgo/critterkingdom/internal/engine"
)

// Dialogue text style: size in CSS pixels and font weight.
const (
	textSize   = 26
	textWeight = 500
)

// noHighlight means no choice button is highlighted.
const noHighlight = -1

// Choice is a choice button as shown: its text, looked up, and its icon
// name, such as "check".
type Choice struct {
	Text string
	Icon string
}

// Dialogue is the overlay that shows a conversation with a villager: their
// portrait and name, and the current line typed out a page of at most two
// lines at a time, with the choice buttons under a question.
//
// A tap anywhere or paw shows the rest of a page that is still typing, or
// moves on to the next page or line. On a question only the buttons answer:
// a tap on one, or paw once Move left or right has highlighted one.
type Dialogue struct {
	runner *dialogue.Runner
	npcID  string
	text   func(key string) string
	// face measures the text for wrapping; its size follows the unit.
	face *text.GoTextFace

	// pages are the current line's pages, wrapped to width screen pixels.
	// They are nil until the first Update, which knows the screen size.
	pages [][]string
	width float64
	page  int
	typer *Typewriter

	highlight int
}

// A Dialogue is an overlay on the scene stack.
var _ engine.Overlay = (*Dialogue)(nil)

// NewDialogue returns the overlay for a conversation that runner has
// started with the villager npcID. text looks up string-table keys, such as
// the line texts and "npc.<npcID>.name", in the player's language. font is
// the game's font.
func NewDialogue(font *text.GoTextFaceSource, runner *dialogue.Runner, npcID string, text func(key string) string) *Dialogue {
	return &Dialogue{
		runner:    runner,
		npcID:     npcID,
		text:      text,
		face:      engine.NewFace(font, textSize, textWeight),
		highlight: noHighlight,
	}
}

// Name returns the villager's name.
func (d *Dialogue) Name() string {
	return d.text("npc." + d.npcID + ".name")
}

// Face returns the portrait expression of the current line, such as
// "happy", or "" once the conversation is over.
func (d *Dialogue) Face() string {
	return d.runner.Line().Face
}

// Text returns the part of the current page typed so far, its lines joined
// by "\n".
func (d *Dialogue) Text() string {
	if d.typer == nil {
		return ""
	}
	return d.typer.Visible()
}

// Choices returns the choice buttons to show. It is empty until the last
// page of a question has been typed out in full.
func (d *Dialogue) Choices() []Choice {
	if !d.answering() {
		return nil
	}
	var out []Choice
	for _, c := range d.runner.Choices() {
		out = append(out, Choice{Text: d.text(c.Text), Icon: c.Icon})
	}
	return out
}

// Highlighted returns the index of the choice button Move has highlighted,
// or -1 when none is.
func (d *Dialogue) Highlighted() int {
	if !d.answering() {
		return noHighlight
	}
	return d.highlight
}

// Done reports whether the conversation is over.
func (d *Dialogue) Done() bool {
	return d.runner.Done()
}

// Update turns this tick's taps, paw presses and moves into the next page,
// line or choice, then types the page on. Menu and Pointing are ignored. It
// returns the runner's error when advancing or choosing fails.
func (d *Dialogue) Update(f engine.Frame) error {
	if d.runner.Done() {
		return nil
	}
	l := LayoutDialogue(f.ScreenW, f.ScreenH, f.Unit, len(d.runner.Choices()))
	d.wrap(l, f.Unit)

	if err := d.handle(f.Input, l); err != nil {
		return err
	}
	if d.typer != nil {
		d.typer.Update(f.DT)
	}
	return nil
}

// handle reacts to the input against what was shown on the last frame.
func (d *Dialogue) handle(in engine.Input, l DialogueLayout) error {
	press := in.Tap || in.Paw
	switch {
	case !d.typer.Done():
		if press {
			d.typer.Skip()
		}
		return nil
	case d.page < len(d.pages)-1:
		if press {
			d.startPage(d.page + 1)
		}
		return nil
	case d.answering():
		return d.answer(in, l)
	case press:
		if err := d.runner.Advance(); err != nil {
			return err
		}
		d.startLine()
	}
	return nil
}

// answer handles input on a question whose buttons are showing.
func (d *Dialogue) answer(in engine.Input, l DialogueLayout) error {
	n := len(d.runner.Choices())
	switch {
	case in.Move.X < 0:
		d.highlight = 0
	case in.Move.X > 0:
		d.highlight = n - 1
	}

	pick := noHighlight
	if in.Tap {
		for i, r := range l.Choices {
			if r.Contains(in.TapAt) {
				pick = i
			}
		}
	}
	if pick == noHighlight && in.Paw {
		pick = d.highlight
	}
	if pick == noHighlight {
		return nil
	}
	if err := d.runner.Choose(pick); err != nil {
		return err
	}
	d.startLine()
	return nil
}

// answering reports whether the buttons of a question are showing: the
// step has choices and its last page is fully typed.
func (d *Dialogue) answering() bool {
	return d.typer != nil && d.typer.Done() && d.page == len(d.pages)-1 &&
		len(d.runner.Choices()) > 0
}

// wrap makes sure the current line is split into pages for the text area
// in l. When the area's width has changed, as when a tablet is turned, the
// line is wrapped again and starts over from its first page.
func (d *Dialogue) wrap(l DialogueLayout, unit float64) {
	width := l.Text.Max.X - l.Text.Min.X
	if d.pages != nil && width == d.width {
		return
	}
	if unit <= 0 {
		unit = 1
	}
	d.face.Size = textSize * unit
	d.width = width
	d.paginate()
}

// startLine begins the runner's current line with no button highlighted.
// It is wrapped at the width last used, so Text and Choices are right
// before the next Update.
func (d *Dialogue) startLine() {
	d.highlight = noHighlight
	if !d.runner.Done() {
		d.paginate()
	}
}

// paginate splits the current line into pages at the width last used and
// starts typing the first.
func (d *Dialogue) paginate() {
	measure := func(s string) float64 { return text.Advance(s, d.face) }
	d.pages = Paginate(d.text(d.runner.Line().Text), d.width, measure)
	if len(d.pages) == 0 {
		// A line with no words still gets a page, so it can be tapped past.
		d.pages = [][]string{{""}}
	}
	d.startPage(0)
}

// startPage starts typing page i of the current line.
func (d *Dialogue) startPage(i int) {
	d.page = i
	d.typer = NewTypewriter(strings.Join(d.pages[i], "\n"))
}

// DrawWorld draws nothing: the dialogue box is all interface.
func (d *Dialogue) DrawWorld(engine.Canvas) {}

// DrawUI draws nothing yet.
func (d *Dialogue) DrawUI(*ebiten.Image, engine.UI) {}
