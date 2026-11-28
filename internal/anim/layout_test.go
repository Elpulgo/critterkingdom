package anim_test

import (
	"reflect"
	"testing"

	"github.com/Elpulgo/critterkingdom/internal/anim"
	"github.com/Elpulgo/critterkingdom/internal/geom"
)

// exportedSheet is a part sheet as `make art` writes it, with two views. The
// parts are listed tail before body on purpose: drawing order is file order,
// not alphabetical.
const exportedSheet = `{
  "pxPerWorldPixel": 8,
  "views": {
    "down": {
      "parts": [
        {"name": "tail", "image": "down_tail.png", "x": 2, "y": -9, "w": 5, "h": 6, "pivotX": 3, "pivotY": -4},
        {"name": "body", "image": "down_body.png", "x": -6, "y": -11.25, "w": 12, "h": 11.375, "pivotX": 0, "pivotY": -5.5625}
      ]
    },
    "side": {
      "parts": [
        {"name": "head", "image": "side_head.png", "x": -8, "y": -25, "w": 16, "h": 15, "pivotX": 0, "pivotY": -11.5}
      ]
    }
  }
}`

func TestParseLayout_ReadsEachViewsPartsInFileOrderWithBoxesAndPivotsFromTheFeet(t *testing.T) {
	layout, err := anim.ParseLayout([]byte(exportedSheet))
	if err != nil {
		t.Fatalf("ParseLayout: %v", err)
	}
	if layout.PxPerWorldPixel != 8 {
		t.Errorf("PxPerWorldPixel = %d, want 8", layout.PxPerWorldPixel)
	}

	tests := []struct {
		view string
		want anim.View
	}{
		{"down", anim.View{Name: "down", Parts: []anim.Part{
			{
				Name:  "tail",
				Image: "down_tail.png",
				Box:   geom.Rect{Min: geom.Vec{X: 2, Y: -9}, Max: geom.Vec{X: 7, Y: -3}},
				Pivot: geom.Vec{X: 3, Y: -4},
			},
			{
				Name:  "body",
				Image: "down_body.png",
				Box:   geom.Rect{Min: geom.Vec{X: -6, Y: -11.25}, Max: geom.Vec{X: 6, Y: 0.125}},
				Pivot: geom.Vec{X: 0, Y: -5.5625},
			},
		}}},
		{"side", anim.View{Name: "side", Parts: []anim.Part{
			{
				Name:  "head",
				Image: "side_head.png",
				Box:   geom.Rect{Min: geom.Vec{X: -8, Y: -25}, Max: geom.Vec{X: 8, Y: -10}},
				Pivot: geom.Vec{X: 0, Y: -11.5},
			},
		}}},
	}
	for _, tt := range tests {
		t.Run(tt.view, func(t *testing.T) {
			if got := layout.View(tt.view); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("View(%q) =\n%+v\nwant\n%+v", tt.view, got, tt.want)
			}
		})
	}
}

func TestLayoutView_AMissingUpViewShowsTheDownParts(t *testing.T) {
	layout, err := anim.ParseLayout([]byte(exportedSheet))
	if err != nil {
		t.Fatalf("ParseLayout: %v", err)
	}
	got := layout.View("up")
	if got.Name != "down" || len(got.Parts) != 2 || got.Parts[0].Name != "tail" || got.Parts[1].Name != "body" {
		t.Errorf("View(\"up\") = %+v, want the down view with tail and body", got)
	}
}

func TestParseLayout_RejectsASheetItCannotDrawFrom(t *testing.T) {
	tests := []struct {
		name  string
		sheet string
	}{
		{"broken JSON", `{"pxPerWorldPixel": 8, "views": {`},
		{"no down view", `{"pxPerWorldPixel": 8, "views": {"side": {"parts": [
			{"name": "head", "image": "side_head.png", "x": -8, "y": -25, "w": 16, "h": 15, "pivotX": 0, "pivotY": -11.5}
		]}}}`},
		{"two parts with the same name in one view", `{"pxPerWorldPixel": 8, "views": {"down": {"parts": [
			{"name": "ear", "image": "down_ear.png", "x": -6, "y": -17, "w": 4, "h": 5, "pivotX": -4, "pivotY": -12},
			{"name": "ear", "image": "down_ear.png", "x": 2, "y": -17, "w": 4, "h": 5, "pivotX": 4, "pivotY": -12}
		]}}}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := anim.ParseLayout([]byte(tt.sheet)); err == nil {
				t.Errorf("ParseLayout accepted a sheet with %s", tt.name)
			}
		})
	}
}
