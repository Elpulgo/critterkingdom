package engine

import (
	"testing"

	"github.com/Elpulgo/critterkingdom/internal/geom"
)

func TestViewFor_ScalesTheWorldByTheLargestWholeNumberAndFillsTheScreen(t *testing.T) {
	tests := []struct {
		name   string
		sw, sh int
		want   View
	}{
		{"full HD shows 320 by 180 at 6x", 1920, 1080, View{Scale: 6, W: 320, H: 180}},
		{"a 4:3 tablet sees a taller world", 2048, 1536, View{Scale: 8, W: 256, H: 192}},
		{"a smaller screen keeps the world size at a lower scale", 960, 540, View{Scale: 3, W: 320, H: 180}},
		{"a window smaller than the minimum world still draws at 1x", 200, 150, View{Scale: 1, W: 200, H: 150}},
		{"an ultrawide screen caps the width and centres the world", 5120, 1440, View{Scale: 8, W: 480, H: 180, OffX: 640}},
		{"a portrait phone caps the height and centres the world", 1170, 2532, View{Scale: 4, W: 293, H: 240, OffX: -1, OffY: 786}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := viewFor(tt.sw, tt.sh); got != tt.want {
				t.Errorf("viewFor(%d, %d) = %+v, want %+v", tt.sw, tt.sh, got, tt.want)
			}
		})
	}
}

func TestScreenToView_FindsTheWorldPixelUnderAFinger(t *testing.T) {
	v := View{Scale: 4, W: 293, H: 240, OffX: -1, OffY: 786}
	if got, want := v.ScreenToView(585, 1186), (geom.Vec{X: 146.5, Y: 100}); got != want {
		t.Errorf("ScreenToView(585, 1186) = %v, want %v", got, want)
	}
}
