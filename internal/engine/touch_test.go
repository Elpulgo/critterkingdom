package engine

import (
	"testing"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/Elpulgo/critterkingdom/internal/geom"
)

// screen is a 1000 by 600 touch screen at one screen pixel per CSS pixel.
// Its paw button is centred at 924,524.
var screen = touchLayout{screenW: 1000, screenH: 600, unit: 1}

func touchAt(id ebiten.TouchID, x, y float64) touchPoint {
	return touchPoint{id: id, pos: geom.Vec{X: x, Y: y}}
}

func TestTouch_TheNewestFingerOnTheWorldSteers(t *testing.T) {
	var tc touchControls
	first, second := touchAt(1, 600, 200), touchAt(2, 700, 300)
	ticks := []struct {
		name    string
		fingers []touchPoint
		want    geom.Vec
	}{
		{"one finger", []touchPoint{first}, geom.Vec{X: 600, Y: 200}},
		{"a second finger takes over", []touchPoint{second, first}, geom.Vec{X: 700, Y: 300}},
		{"the first steers again when the second lifts", []touchPoint{first}, geom.Vec{X: 600, Y: 200}},
	}
	for _, tk := range ticks {
		in := tc.step(tk.fingers, screen)
		if !in.pointing || in.point != tk.want {
			t.Fatalf("%s: pointing %v at %v, want pointing at %v", tk.name, in.pointing, in.point, tk.want)
		}
	}
}

func TestTouch_ThePawButtonPressesOnceWhenAFingerLandsOnIt(t *testing.T) {
	var tc touchControls
	tc.step([]touchPoint{touchAt(1, 600, 200)}, screen) // the first touch shows the buttons
	tc.step(nil, screen)

	onButton := []touchPoint{touchAt(2, 924, 524)}
	if in := tc.step(onButton, screen); !in.paw || in.pointing {
		t.Fatalf("finger lands on the button: paw %v, pointing %v; want paw and no walking", in.paw, in.pointing)
	}
	if in := tc.step(onButton, screen); in.paw {
		t.Fatalf("finger held on the button: paw pressed again, want one press per tap")
	}
}

func TestTouch_TheLeftHalfSteersWithAPadOnlyWhenThePadIsEnabled(t *testing.T) {
	// The finger touches down at 100,500 and slides right by 0.9 of the pad's
	// radius of 64.
	down, slid := touchAt(1, 100, 500), touchAt(1, 157.6, 500)
	tests := []struct {
		name         string
		pad          bool
		wantMove     geom.Vec
		wantPointing bool
	}{
		{"with the pad the finger steers right", true, geom.Vec{X: 1}, false},
		{"without the pad the finger walks to where it is", false, geom.Vec{}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tc := touchControls{padEnabled: tt.pad}
			tc.step([]touchPoint{down}, screen)
			in := tc.step([]touchPoint{slid}, screen)
			if in.move != tt.wantMove || in.pointing != tt.wantPointing {
				t.Errorf("move %v, pointing %v; want move %v, pointing %v", in.move, in.pointing, tt.wantMove, tt.wantPointing)
			}
		})
	}
}

func TestTouch_AFingerLandingAnywhereTapsOnceWhereItLanded(t *testing.T) {
	tests := []struct {
		name string
		pad  bool
		at   touchPoint
	}{
		{"on the world", false, touchAt(2, 600, 200)},
		{"on the paw button", false, touchAt(2, 924, 524)},
		{"on the pad half with the pad enabled", true, touchAt(2, 100, 500)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tc := touchControls{padEnabled: tt.pad}
			tc.step([]touchPoint{touchAt(1, 700, 100)}, screen) // the first touch shows the buttons
			tc.step(nil, screen)

			if in := tc.step([]touchPoint{tt.at}, screen); !in.tap || in.tapAt != tt.at.pos {
				t.Fatalf("finger lands: tap %v at %v, want a tap at %v", in.tap, in.tapAt, tt.at.pos)
			}
			if in := tc.step([]touchPoint{tt.at}, screen); in.tap {
				t.Fatalf("finger held: tapped again, want one tap per touch")
			}
		})
	}
}
