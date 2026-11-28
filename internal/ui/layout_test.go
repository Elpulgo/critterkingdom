package ui_test

import (
	"testing"

	"github.com/Elpulgo/critterkingdom/internal/ui"
)

func TestLayoutDialogue_ChoiceButtonsAreBigTouchTargetsOnScreenAndApart(t *testing.T) {
	cases := []struct {
		name       string
		cssW, cssH float64
		unit       float64
	}{
		{"phone at unit 1", 568, 320, 1},
		{"phone at unit 2", 568, 320, 2},
		{"tablet at unit 1", 1024, 768, 1},
		{"tablet at unit 2", 1024, 768, 2},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			screenW, screenH := tc.cssW*tc.unit, tc.cssH*tc.unit
			least := 64 * tc.unit

			l := ui.LayoutDialogue(screenW, screenH, tc.unit, 2)

			if len(l.Choices) != 2 {
				t.Fatalf("got %d choice buttons, want 2", len(l.Choices))
			}
			for i, r := range l.Choices {
				if w, h := r.Max.X-r.Min.X, r.Max.Y-r.Min.Y; w < least || h < least {
					t.Errorf("button %d is %.0fx%.0f screen px, want at least %.0f each way", i, w, h, least)
				}
				if r.Min.X < 0 || r.Min.Y < 0 || r.Max.X > screenW || r.Max.Y > screenH {
					t.Errorf("button %d at %v is off the %.0fx%.0f screen", i, r, screenW, screenH)
				}
			}
			a, b := l.Choices[0], l.Choices[1]
			if a.Min.X < b.Max.X && b.Min.X < a.Max.X && a.Min.Y < b.Max.Y && b.Min.Y < a.Max.Y {
				t.Errorf("buttons %v and %v overlap", a, b)
			}
		})
	}
}
