package combat

import (
	"testing"

	"github.com/Elpulgo/critterkingdom/internal/geom"
)

func TestHurtBox_IsTheMiddleSixtyPercentOfTheSprite(t *testing.T) {
	sprite := geom.Rect{Min: geom.Vec{X: 0, Y: 0}, Max: geom.Vec{X: 20, Y: 10}}
	want := geom.Rect{Min: geom.Vec{X: 4, Y: 2}, Max: geom.Vec{X: 16, Y: 8}}
	if got := HurtBox(sprite); got != want {
		t.Errorf("HurtBox(%v) = %v, want %v", sprite, got, want)
	}
}

func TestSwingBox_IsWiderThanTheStickAndReachesEnemiesBesideAndJustPastIt(t *testing.T) {
	swing := SwingBox(geom.Vec{X: 100, Y: 100}, geom.Vec{X: 1}, 12)
	wantSwing := geom.Rect{Min: geom.Vec{X: 96, Y: 90}, Max: geom.Vec{X: 116, Y: 110}}
	if swing != wantSwing {
		t.Fatalf("SwingBox = %v, want %v", swing, wantSwing)
	}

	tests := []struct {
		name   string
		sprite geom.Rect
		want   bool
	}{
		{
			"hits an enemy standing beside the stick's line",
			geom.Rect{Min: geom.Vec{X: 104, Y: 84}, Max: geom.Vec{X: 120, Y: 100}},
			true,
		},
		{
			"hits an enemy standing just past the stick's tip",
			geom.Rect{Min: geom.Vec{X: 112, Y: 92}, Max: geom.Vec{X: 128, Y: 108}},
			true,
		},
		{
			"misses an enemy too far to the side",
			geom.Rect{Min: geom.Vec{X: 110, Y: 112}, Max: geom.Vec{X: 126, Y: 128}},
			false,
		},
		{
			"misses an enemy behind the player",
			geom.Rect{Min: geom.Vec{X: 70, Y: 92}, Max: geom.Vec{X: 86, Y: 108}},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hurt := HurtBox(tt.sprite)
			if got := Overlap(swing, hurt); got != tt.want {
				t.Errorf("Overlap(%v, HurtBox(%v)=%v) = %v, want %v", swing, tt.sprite, hurt, got, tt.want)
			}
		})
	}
}

func TestOverlap_IncludesTouchingEdgesAndCornersAndRejectsAnyGap(t *testing.T) {
	a := geom.Rect{Min: geom.Vec{X: 10, Y: 10}, Max: geom.Vec{X: 20, Y: 20}}
	tests := []struct {
		name string
		b    geom.Rect
		want bool
	}{
		{"overlapping interiors", geom.Rect{Min: geom.Vec{X: 15, Y: 15}, Max: geom.Vec{X: 25, Y: 25}}, true},
		{"one box inside the other", geom.Rect{Min: geom.Vec{X: 12, Y: 12}, Max: geom.Vec{X: 18, Y: 18}}, true},
		{"touching along an edge", geom.Rect{Min: geom.Vec{X: 20, Y: 10}, Max: geom.Vec{X: 30, Y: 20}}, true},
		{"touching at a single corner", geom.Rect{Min: geom.Vec{X: 20, Y: 20}, Max: geom.Vec{X: 30, Y: 30}}, true},
		{"gap on X", geom.Rect{Min: geom.Vec{X: 21, Y: 10}, Max: geom.Vec{X: 30, Y: 20}}, false},
		{"gap on Y", geom.Rect{Min: geom.Vec{X: 10, Y: 21}, Max: geom.Vec{X: 20, Y: 30}}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Overlap(a, tt.b); got != tt.want {
				t.Errorf("Overlap(%v, %v) = %v, want %v", a, tt.b, got, tt.want)
			}
			if got := Overlap(tt.b, a); got != tt.want {
				t.Errorf("Overlap(%v, %v) = %v, want %v", tt.b, a, got, tt.want)
			}
		})
	}
}
