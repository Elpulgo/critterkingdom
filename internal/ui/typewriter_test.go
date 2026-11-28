package ui

import "testing"

// page is 40 characters but 43 bytes: å and ä count as one character each.
const page = "Hej på dig! Vill du ha en bulle här, då?"

func TestTypewriter_TypesFortyCharactersPerSecondCountingSwedishLettersAsOne(t *testing.T) {
	tw := NewTypewriter(page)

	tw.Update(0.5)
	if got, want := tw.Visible(), "Hej på dig! Vill du "; got != want {
		t.Fatalf("after 0.5 s shows %q, want %q", got, want)
	}
	if tw.Done() {
		t.Fatal("done after 0.5 s of a 40-character page, want still typing")
	}

	tw.Update(0.5)
	if got := tw.Visible(); got != page || !tw.Done() {
		t.Fatalf("after 1 s shows %q (done %v), want the whole page and done", got, tw.Done())
	}
}
