// Package ui draws the game's interface over the world and turns taps and
// paw presses on it into actions: for now the dialogue box's typewriter text
// and its two-line pages.
package ui

import "math"

// CharsPerSecond is how fast dialogue text types out.
const CharsPerSecond = 40

// Typewriter reveals one page of text a character at a time. A character is
// a rune, so "å" counts as one.
type Typewriter struct {
	page    []rune
	elapsed float64
	shown   int
}

// NewTypewriter starts typing page from its first character.
func NewTypewriter(page string) *Typewriter {
	return &Typewriter{page: []rune(page)}
}

// Update moves the typing on by dt seconds.
func (t *Typewriter) Update(dt float64) {
	if t.Done() {
		return
	}
	t.elapsed += dt
	// The small nudge keeps sums of frame times such as 30 × 1/60 from
	// landing a hair under a whole character.
	n := int(math.Floor(t.elapsed*CharsPerSecond + 1e-9))
	t.shown = min(max(n, 0), len(t.page))
}

// Skip shows the whole page at once.
func (t *Typewriter) Skip() {
	t.shown = len(t.page)
}

// Done reports whether the whole page is shown.
func (t *Typewriter) Done() bool {
	return t.shown >= len(t.page)
}

// Visible returns the part of the page typed so far.
func (t *Typewriter) Visible() string {
	return string(t.page[:t.shown])
}
